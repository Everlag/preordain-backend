package UserStructs

import (

	"sync"
	"time"

	"fmt"
)

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

//Returns a valid user session if the provided password is valid
func (aUser *User) getNewSession() string {

	aUser.sessionLock.Lock()
	defer aUser.sessionLock.Unlock()

	freshSession := newSession()
	aUser.Sessions[freshSession.Key] = &freshSession

	return freshSession.Key
}

// Returns nil if a new, valid password reset token has been dispatched
// for a user
func (aUser *User) getResetToken() error {

	aUser.sessionLock.Lock()
	defer aUser.sessionLock.Unlock()

	if aUser.PasswordResetToken.insideAbuseWindow(){
		return fmt.Errorf("Password token already issued inside abuse window.")		
	}

	aUser.PasswordResetToken = newSession()

	return nil

}

func (aUser *User) getCollectionList(public bool) ([]string, error) {
	
	aUser.sessionLock.RLock()
	defer aUser.sessionLock.RUnlock()

	collectionList:= make([]string, 0)

	for aCollectionName, aCollection:= range aUser.Collections{

		if public && aCollection.PublicViewing {
			collectionList = append(collectionList, aCollectionName)
		}else if !public {
			collectionList = append(collectionList, aCollectionName)
		}
	}

	if len(collectionList) == 0 {
		return nil, fmt.Errorf("No collections available")	
	}

	return collectionList, nil

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

	return aUser.Collections[collName].AddTrade(aTrade)

}

func (aUser *User) addCollection(collName string) error {

	aUser.sessionLock.Lock()
	defer aUser.sessionLock.Unlock()

	//perform the check for existing collections
	if aUser.collectionExists(collName) {
		return fmt.Errorf("Collection already exists, cannot replace.")
	}

	if len(aUser.Collections) + 1 > aUser.MaxCollections {
		return fmt.Errorf("Too many collections")
	}

	//create the collection
	aCollection := CreateCollection(collName)

	//actually add the collection
	aUser.Collections[collName] = &aCollection

	return nil
}

func (aUser *User) setPermissions(collName string,
	Viewing, History, Comments bool) (error) {
	
	aUser.sessionLock.Lock()
	defer aUser.sessionLock.Unlock()

	// Perform the check for existing collections
	if !aUser.collectionExists(collName) {
		return fmt.Errorf("Collection doesn't exist, cannot add trade.")
	}

	return aUser.Collections[collName].SetPermissions(Viewing, History, Comments)

}

//checks the provided session code for validity.
func (aUser *User) checkSession(sessionCode string) bool {

	aUser.sessionLock.RLock()
	defer aUser.sessionLock.RUnlock()

	//see if the session exists
	aSession, ok := aUser.Sessions[sessionCode]
	if !ok {
		return false
	}

	//check for expiry of the session
	now := time.Now().UTC().Unix()
	if aSession.End < now {
		delete(aUser.Sessions, sessionCode)
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