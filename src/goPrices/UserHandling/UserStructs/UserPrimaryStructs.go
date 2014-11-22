package UserStructs

import(

	"./databaseHandling"

	"time"
	"sync"
	"log"

	"fmt"

)


//constants to use for instantiating a database.
const snapshotRotationTime int64 = 3600 * 24
const snapshotCount int = 14

//the time between checks to ensure the manager is not yet dirty
const managerSerializationTime int = 1800

//if a user's session token is within userSessionExtensionLength of its expiry,
//it being used will increment it by the userSessionExtensionLength
const userSessionExtensionLength int64 = 3600 * 24 * 7

//A singleton object that provides access to a user database.
//
//Users are stored in the database under the following, per user, key
//*UserManager.Suffix*-user-*userName*
type UserManager struct{

	//the suffix identifying this user manager as a unique entity.
	//
	//appended to the generic database name and acts as an identifier of the
	//file containing this manager after serialization
	Suffix string

	//Maps from a given invite code to an array containing
	//[usesRemaining, Collections Per Use]
	InviteCodes map[string][]int

	//contains users who have been fetched from the database to
	//reside in memory.
	//
	//this is a write-through cache. all writes to the user are automically
	//sent to the database when the User has been updated.
	users map[string]*User

	//contains the database abstractions we need to use in order to have
	//confidence our data is safely stored
	storage *databaseHandling.WrappedStorage

	//the dirty bool allows us the ability to determine when the manager
	//has been changed so we can serialize only when necessary
	//
	//as the manager has only its InviteCodes as mutable, this changes
	//only when the invite codes change
	dirty bool
	//a background 
	daemonRunning bool

	logger log.Logger

}

//derives the backing database name for a manager with the provided suffix
func deriveDatabaseName(suffix string) string {
	return suffix + "Users"
}

//derives the name of the manager's metadata file
func deriveMetaDataName(suffix string) string {
	return suffix+".meta.json"
}

//creates a new, blank user manager along with its backing storage on disk.
//
//this does not return a manager as it is meant to be used a single time for
//instantiating all the necessary components for user data.
func NewUserManager(suffix string) {
	var err error

	//create the database
	dbName:= deriveDatabaseName(suffix)
	aDatabase, err:= databaseHandling.NewWrappedStorage(dbName,
					snapshotRotationTime, snapshotCount)
	if err!=nil {
		fmt.Println("Failed to create database, panicking")
		panic("db failure")
	}

	aDatabase.SafeClose()

	//create the manager
	aManager:= UserManager{Suffix:suffix}

	err = aManager.save()
	if err!=nil {
		fmt.Println("Failed to save manager, panicking")
		panic("db failure")
	}

}

//commits the user, as defined by their name which indexes into the .users map
//of the manger, to persistent storage
func (aManager *UserManager) userToStorage(name string) error {
	aUser, ok := aManager.users[name]
	if !ok {
		return fmt.Errorf("Failed to commit user to persistent storage")
	}

	//derive the unique index into the storage for this user defined as
	// *UserManager.Suffix*-user-*userName*
	storageIndex := []byte(aManager.Suffix + "-user-" + name)

	userData, err := aUser.ToJson()
	if err!=nil {
		return err
	}

	aManager.storage.WriteValue(storageIndex, userData)

	return nil

}

//runs the daemon that ensures the on disk metadata matches the current
func (aManager *UserManager) runDaemon() {
	if !aManager.daemonRunning {
		go aManager.metadataDaemon()
		aManager.daemonRunning = true	
	}
}

//gets the user from the manager, returns an error if the user doesn't exist
func (aManager *UserManager) GetUser(name string) (*User, error) {
	//check the manager's in memory cache
	aUser, ok:= aManager.users[name]
	if ok {
		return aUser, nil
	}

	//check the database
	userData, failedToFind:= aManager.storage.ReadValue([]byte(name))
	if failedToFind == nil {
		//if the user exists in the database, then we can unmarshal it
		//and store it in memory
		aUser, err:= userFromJson(userData)
		aManager.users[name] = &aUser
		if err==nil {
			return &aUser, nil	 
		}
	}	

	return &User{}, fmt.Errorf("Failed to find user")
}

//checks for the existence of a specified user
func (aManager *UserManager) userExists(name string) bool {

	_, err:=aManager.GetUser(name)

	if err!=nil {
		return false
	}

	return true

}

//creates a user if they don't already exist, returns an error if they do or
//if we failed to commit them. Returns a usable session key if
//this encounters no errors.
//
//they are brought into the in-memory aManager.users map automatically.
func (aManager *UserManager) AddUser(name, email, password string,
										maxCollections int) (string, error) {
	
	if aManager.userExists(name) {
		return "", fmt.Errorf("User already exists")
	}

	nonce, err := getArrayOfRandBytes(32)
	if err!=nil {
		return "", fmt.Errorf("Failed to derive password")
	}

	passwordHash, err := DerivePassword([]byte(password), nonce)
	if err != nil {
		return "", err
	}

	firstSession := NewSession()

	aFreshUser:= User{

		Name: name,
		Email: email,

		HashedPass: passwordHash,
		Nonce: nonce,

		Collections: make(map[string]*Collection),

		MaxCollections: maxCollections,

		Sessions: make(map[string]*UserSession),

	}

	//setup the first session for this user
	aFreshUser.Sessions[firstSession.Key] = &firstSession

	//setup the user in the manager's memory
	aManager.users[name] = &aFreshUser

	//write the user to persistent storage
	aManager.userToStorage(name)

	return firstSession.Key, nil

}

type User struct{

	//their name. it must be unique
	Name string

	//their email. doesn't need to be unique
	Email string

	//their password hashed using scrypt
	//
	//parameters are to be decided
	HashedPass []byte
	//we use a salt in this case because security
	Nonce []byte

	//the various collections each user can have.
	//typically, they likely have one
	Collections map[string]*Collection

	CollectionNames []string

	//how many collections a user can have at one time.
	MaxCollections int

	//this contains a token that allows a user
	//to reset their password via an api endpoin.
	//
	//length is standardized to be 16 characters
	PasswordResetToken string
		//this is when the token should no longer
		//be considered valid. utc timestamp.
	PasswordResetTokenDeathTime int64

	//valid user sessions map to their object
	//
	//when a session is stale it is removed from this map.
	//
	//keys are derived from generating a 64 byte array and then casting it to
	//a string
	Sessions map[string]*UserSession

	//we don't want to have to deal with sessions being removed after they
	//have been shown to exist so we use this lock.
	sessionLock sync.Mutex

}

//checks the provided session code for validity.
func (aUser *User) CheckSession(sessionCode string) bool {
	
	aUser.sessionLock.Lock()

	//see if the session exists
	aSession, ok := aUser.Sessions[sessionCode]
	if !ok {
		aUser.sessionLock.Unlock()
		return false
	}

	//check for expiry of the session
	now := time.Now().UTC().Unix()
	if aSession.End < now {
		delete(aUser.Sessions, sessionCode)
		aUser.sessionLock.Unlock()
		return false
	}

	aUser.sessionLock.Unlock()

	return true

}

//a user session contains the timestamp for which it is valid up to and the key
//that it is associated with.
//
//everytime a session is used its end time should be incremented by a week.
type UserSession struct{

	Key string
	End int64

}

//returns a fresh session valid till userSessionExtensionLength * 2 + NOW
func NewSession() UserSession {
	endTime:= (userSessionExtensionLength * 2) + time.Now().UTC().Unix()

	key := randString(32)

	return UserSession{key, endTime}
}