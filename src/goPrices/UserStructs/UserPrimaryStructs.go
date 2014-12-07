package UserStructs

import (
	"./databaseHandling"

	"log"
	"sync"
	"time"

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
type UserManager struct {

	//the suffix identifying this user manager as a unique entity.
	//
	//appended to the generic database name and acts as an identifier of the
	//file containing this manager after serialization
	Suffix string

	//Maps from a given invite code to an array containing
	//[usesRemaining, Collections Per Use]
	InviteCodes map[string][2]int

	// Records the name of each user. Used for bookkeeping and as a way
	// to ensure that we always know the names of users in case we need to
	// initiate action without their demand.
	UserNames []string

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

	//allows us to signal to daemons that we have finished
	dead bool

	// nameLock ensures only one writer to the precious list of user names
	nameLock sync.RWMutex

	logger *log.Logger
}

//runs the daemon that ensures the on disk metadata matches the current
func (aManager *UserManager) runDaemon() {
	if !aManager.daemonRunning {
		go aManager.metadataDaemon()
		aManager.daemonRunning = true
	}
}

//derives a key used for indexing into the db for users
func (aManager *UserManager) deriveUserDBKey(name string) []byte {
	return []byte(aManager.Suffix + "-user-" + name)
}

//commits the user, as defined by their name which indexes into the .users map
//of the manger, to persistent storage
func (aManager *UserManager) userToStorage(name string) error {
	aUser, ok := aManager.users[name]
	if !ok {
		return fmt.Errorf("Failed to commit user to persistent storage")
	}

	//ensure nothing can touch the user while we deal with it
	aUser.sessionLock.Lock()
	defer aUser.sessionLock.Unlock()

	//derive the unique index into the storage for this user defined as
	// *UserManager.Suffix*-user-*userName*
	storageIndex := aManager.deriveUserDBKey(name)

	userData, err := aUser.ToJson()
	if err != nil {
		return err
	}

	aManager.storage.WriteValue(storageIndex, userData)

	return nil

}

//handles committing the user to the persistent storage.
//
//call this after all changes to a user have been performed
func (aManager *UserManager) userModified(name string) error {
	return aManager.userToStorage(name)
}

//gets the user from the manager, returns an error if the user doesn't exist
func (aManager *UserManager) getUser(name string) (*User, error) {
	//check the manager's in memory cache
	aUser, ok := aManager.users[name]
	if ok {
		return aUser, nil
	}

	//check the database
	dbKey := aManager.deriveUserDBKey(name)
	userData, failedToFind := aManager.storage.ReadValue(dbKey)
	if failedToFind == nil {
		//if the user exists in the database, then we can unmarshal it
		//and store it in memory
		aUser, err := userFromJson(userData)
		if err == nil {
			aManager.users[name] = &aUser
			return &aUser, nil
		}
	}

	return &User{}, fmt.Errorf("Failed to find user")
}

//attempts to acquire a user using the provided name. if the user exists,
//the session is checked. if the session is valid, a valid user is returned.
//if not, returns and error
func (aManager *UserManager) authenticateUser(name, session string) (*User, error) {

	aUser, err := aManager.getUser(name)
	if err != nil {
		return nil, fmt.Errorf("Failed to acquire user")
	}

	if !aUser.checkSession(session) {
		return nil, fmt.Errorf("Failed to acquire user")
	}

	//here we have a valid user with authenticated session
	return aUser, nil

}

//attempts to acquire the user using the provided name and session key.
//
//if the user exists and this is not a public context, then we authenticate
//them and return no error on success. On a public context, we simply get the
//user and its the caller's responsibility to deal with it.
func (aManager *UserManager) getUserWithAuthentication(name, sessionKey string,
	public bool) (*User, error) {

	var aUser *User
	var err error

	if !public {
		//attempt authentication if not a public attempt
		aUser, err = aManager.authenticateUser(name, sessionKey)
		if err != nil {
			return nil, err
		}
	} else {
		//a public authentication bypasses this
		aUser, err = aManager.getUser(name)
		if err != nil {
			return nil, err
		}
	}

	return aUser, nil

}

//checks for the existence of a specified user
func (aManager *UserManager) userExists(name string) bool {

	_, err := aManager.getUser(name)

	if err != nil {
		return false
	}

	return true

}

//derives the backing database name for a manager with the provided suffix
func deriveDatabaseName(suffix string) string {
	return suffix + "Users"
}

//derives the name of the manager's metadata file
func deriveMetaDataName(suffix string) string {
	return suffix + ".meta.json"
}

//creates a new, blank user manager along with its backing storage on disk.
//
//this does not return a manager as it is meant to be used a single time for
//instantiating all the necessary components for user data.
func NewUserManager(suffix string) error {
	var err error

	//create the database
	dbName := deriveDatabaseName(suffix)
	aDatabase, err := databaseHandling.NewWrappedStorage(dbName,
		snapshotRotationTime, snapshotCount)
	if err != nil {
		fmt.Println("Failed to create database, exiting")
		return err
	}

	//create the manager
	aManager := UserManager{
		Suffix:      suffix,
		InviteCodes: make(map[string][2]int),
	}
	aManager.setStorage(aDatabase)
	//start a fresh logger for said manager
	aManager.startLogger()

	fmt.Println(aManager.storage.GetLocation())

	err = aManager.save()
	if err != nil {
		fmt.Println("Failed to save manager, exiting")
		return err
	}

	aManager.Close()

	return nil

}

//closes the user manager and associated backing storage safely.
//
//any attempts to use the manager after this will have undefined behavior.
func (aManager *UserManager) Close() {
	aManager.dead = true
	aManager.logger.Println("CLOSING manager")
	aManager.logger.Println("Closing Database")
	aManager.storage.SafeClose()
	aManager.logger.Println("Saving Manager metadata")
	aManager.save()
	aManager.logger.Println("Close Completed")
}

//All methods of each user are goroutine safe via a session lock preventing
//writers from mutating state at the same time readers are looking at it
//
//This means there is a slight overhead for locking but ensures atomicity for
//the entire user. Readers can operate concurrently, writers will lock up
//everything
type User struct {

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

	//The token that allows the user to reset their password is stored as a
	//session sp timing is handled for free
	PasswordResetToken UserSession

	//valid user sessions map to their object
	//
	//when a session is stale it is removed from this map.
	//
	//keys are derived from generating a 64 byte array and then casting it to
	//a string
	Sessions map[string]*UserSession

	//we don't want to have to deal with sessions being removed after they
	//have been shown to exist so we use this lock.
	//
	//This lock ensures any operation mutating the state of the user is done
	//synchronously
	sessionLock sync.RWMutex
}

//Returns a valid user session
func (aUser *User) getNewSession() string {

	aUser.sessionLock.Lock()
	defer aUser.sessionLock.Unlock()

	freshSession := newSession()
	aUser.Sessions[freshSession.Key] = &freshSession

	return freshSession.Key
}

//Returns a valid password reset token
func (aUser *User) getResetToken() string {

	aUser.sessionLock.Lock()
	defer aUser.sessionLock.Unlock()

	aUser.PasswordResetToken = newSession()

	return aUser.PasswordResetToken.Key

}

func (aUser *User) getCollection(collName string,
	public bool) (*Collection, error) {

	aUser.sessionLock.RLock()
	defer aUser.sessionLock.RUnlock()

	if !aUser.collectionExists(collName) {
		return nil, fmt.Errorf("Collection doesn't exist")
	}

	//deal with public and private access
	if public {
		cleanedCollection, err := aUser.Collections[collName].StripToPublic()
		if err != nil {
			return nil, fmt.Errorf("Collection doesn't exist")
		}
		return cleanedCollection, nil
	}

	return aUser.Collections[collName], nil

}

func (aUser *User) collectionExists(collName string) bool {

	if aUser.Collections[collName] != nil {
		return true
	}

	return false
}

func (aUser *User) addTrade(collName string, aTrade Trade) error {
	aUser.sessionLock.Lock()
	defer aUser.sessionLock.Unlock()

	//perform the check for existing collections
	if !aUser.collectionExists(collName) {
		return fmt.Errorf("Collection doesn't exist, cannot add trade.")
	}

	aUser.Collections[collName].AddTrade(aTrade)

	return nil

}

func (aUser *User) addCollection(collName string) error {

	aUser.sessionLock.Lock()
	defer aUser.sessionLock.Unlock()

	//perform the check for existing collections
	if aUser.collectionExists(collName) {
		return fmt.Errorf("Collection already exists, cannot replace.")
	}

	//create the collection
	aCollection := CreateCollection(collName)

	//actually add the collection
	aUser.Collections[collName] = &aCollection

	return nil
}

//checks the provided session code for validity.
func (aUser *User) checkSession(sessionCode string) bool {

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

	return true

}

func (aUser *User) checkResetToken(resetTokenKey string) error {

	if !aUser.PasswordResetToken.valid() ||
		aUser.PasswordResetToken.Key != resetTokenKey {
		return fmt.Errorf("Invalid password reset token")
	}

	return nil

}

//attempts to change the user's password. returns the only valid
//session key if it works
func (aUser *User) changePassword(resetTokenKey, newPassword string) (string, error) {
	aUser.sessionLock.Lock()
	defer aUser.sessionLock.Unlock()

	//check resetTokenKey
	if aUser.checkResetToken(resetTokenKey) != nil {
		return "", fmt.Errorf("Invalid password reset token")
	}

	//check requirements
	if !passwordMeetsRequirements(newPassword) {
		return "", fmt.Errorf("Password doesn't meet minimum requirements")
	}

	//nuke the existing sessions
	aUser.Sessions = make(map[string]*UserSession)

	//grab a new one
	freshSession := newSession()
	aUser.Sessions[freshSession.Key] = &freshSession

	//replace the user's password
	nonce, passwordHash, err := passwordDerivation(newPassword)
	if err != nil {
		return "", fmt.Errorf("Failed to derive password, not changed")
	}

	aUser.Nonce, aUser.HashedPass = nonce, passwordHash

	aUser.PasswordResetToken.end()

	return freshSession.Key, nil

}

//a user session contains the timestamp for which it is valid up to and the key
//that it is associated with.
//
//everytime a session is used its end time should be incremented by
//userSessionExtensionLength
type UserSession struct {
	Key string
	End int64
}

//extends a user session if it is valid and is within
//userSessionExtensionLength of its expiry
//
//returns an error if the token is expired, otherwise returns nil
func (aSession *UserSession) extend() error {

	if !aSession.valid() {
		return fmt.Errorf("Token is expired")
	}

	now := time.Now().UTC().Unix()

	if aSession.End+userSessionExtensionLength > now {
		aSession.End += userSessionExtensionLength
	}

	return nil

}

//invalidates this session
func (aSession *UserSession) end() {

	aSession.End = -1

}

//checks if the session is past its end time.
func (aSession *UserSession) valid() bool {

	now := time.Now().UTC().Unix()
	if now > aSession.End {
		return false
	}

	return true

}

//returns a fresh session valid till userSessionExtensionLength * 2 + NOW
func newSession() UserSession {
	endTime := (userSessionExtensionLength * 2) + time.Now().UTC().Unix()

	key := randString(32)

	return UserSession{key, endTime}
}
