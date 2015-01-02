package UserStructs

import (
	"./databaseHandling"
	"./MailHandling"

	"log"
	"sync"

	"fmt"
)

// Constants to use for instantiating a database.
const snapshotRotationTime int64 = 3600 * 24
const snapshotCount int = 14

// A singleton object that provides access to a user database.
//
// Users are stored in the database under the following, per user, key
// *UserManager.Suffix*-user-*userName*
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

	// Contains the mailing abstractions we require to work
	mailer *MailHandling.Mailer

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

// Runs the daemon that ensures the on disk metadata matches the current
func (aManager *UserManager) runDaemon() {
	if !aManager.daemonRunning {
		go aManager.metadataDaemon()
		aManager.daemonRunning = true
	}
}

// Derives a key used for indexing into the db for users
func (aManager *UserManager) deriveUserDBKey(name string) []byte {
	return []byte(aManager.Suffix + "-user-" + name)
}

// Commits the user, as defined by their name which indexes into the .users map
// of the manger, to persistent storage
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

// Handles committing the user to the persistent storage.
//
// Call this after all changes to a user have been performed
func (aManager *UserManager) userModified(name string) error {
	return aManager.userToStorage(name)
}

// Gets the user from the manager, returns an error if the user doesn't exist
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