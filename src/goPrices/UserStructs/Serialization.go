package UserStructs

import (
	"./databaseHandling"
	"./MailHandling"

	"encoding/json"
	"fmt"

	"io/ioutil"
	"time"
)

const metadataWaitTime int = 1200

func (someData *User) ToJson() ([]byte, error) {
	marshalledData, err := json.Marshal(someData)
	if err != nil {
		fmt.Println("Failed to marhsal user data")
		return nil, fmt.Errorf("Failed to marhsal user data")
	}

	return marshalledData, nil
}

func userFromJson(jsonData []byte) (User, error) {
	var aUser User
	err := json.Unmarshal(jsonData, &aUser)
	if err != nil {
		return User{}, err
	}

	return aUser, nil
}

//converts the set data to json for disk purposes
func (someData *Collection) ToJson() ([]byte, error) {
	marshalledData, err := json.Marshal(someData)
	if err != nil {
		fmt.Println("Failed to marhsal collection data")
		return nil, fmt.Errorf("Failed to marhsal collection data")
	}

	return marshalledData, nil
}

func collectionFromJson(jsonData []byte) (Collection, error) {
	var aCollection Collection
	err := json.Unmarshal(jsonData, &aCollection)
	if err != nil {
		return Collection{}, err
	}

	return aCollection, nil
}

//converts the set data to json for disk purposes
func (someData *UserManager) ToJson() ([]byte, error) {
	marshalledData, err := json.Marshal(someData)
	if err != nil {
		fmt.Println("Failed to marhsal manager metadata")
		return nil, fmt.Errorf("Failed to marhsal manager metadata")
	}

	return marshalledData, nil
}

//converts the set data to json for disk purposes
func managerFromJson(jsonData []byte) (UserManager, error) {
	var aManager UserManager
	err := json.Unmarshal(jsonData, &aManager)
	if err != nil {
		return UserManager{}, err
	}

	return aManager, nil
}

//regularly checks to ensure the on disk version of the manager is up to date
//with the in memory version
func (aManager *UserManager) metadataDaemon() {
	sleepTime := time.Duration(metadataWaitTime) * time.Second

	for {
		//a dead manager means this goroutine can exit
		if aManager.dead {
			return
		}
		//check if the manager has changed
		if aManager.dirty {

			err := aManager.save()
			if err != nil {
				//attempt again, if this fails, then we wait till the next
				//cycle
				aManager.logger.Println("Retrying metadata save")
				aManager.save()
			}
		}

		time.Sleep(sleepTime)

	}
}

//saves a manager's metadata. logs all errors created.
func (aManager *UserManager) save() error {

	//log the event
	aManager.logger.Println("Committing manager metadata")

	//serialize
	data, err := aManager.ToJson()
	if err != nil {
		//if we ran into an error, we log the event and then wait for
		//the next rotation to try again
		aManager.logger.Println("Failed to serialize manager metadata")
		return err
	}

	err = ioutil.WriteFile(deriveMetaDataName(aManager.Suffix), data, 0666)
	if err != nil {
		aManager.logger.Println("FAILURE to commit user metadata")
		return err
	}

	return nil

}

func ReacquireManager(suffix string) (*UserManager, error) {
	//reacquires the manager with the provided suffix
	loc := deriveMetaDataName(suffix)

	jsonData, err := ioutil.ReadFile(loc)
	if err != nil {
		return nil, err
	}

	aManager, err := managerFromJson(jsonData)

	//the manager has its metadata, we need to acquire the db now
	dbName := deriveDatabaseName(suffix)
	someStorage, err := databaseHandling.ReacquireWrappedStorage(dbName)
	if err != nil {
		return nil, err
	}

	aManager.setStorage(someStorage)

	//create a fresh users map to populate
	aManager.setUsers()

	//hook up a fresh logger.
	aManager.startLogger()

	//run the daemon that keeps the on disk metadata up to date!
	aManager.runDaemon()

	// setup the mail service we need
	err = aManager.startMailer()
	if err!=nil {
		return nil, err
	}

	//return the manager reincarnated from the flames of NAND
	return &aManager, nil

}

func (aManager *UserManager) startLogger() {
	aLogger := getLogger("UserStorage."+aManager.Suffix+
		".log", aManager.Suffix)

	aManager.logger = aLogger
}

//hooks up the backing database to the provided manager
func (aManager *UserManager) setStorage(someStorage *databaseHandling.WrappedStorage) {

	aManager.storage = someStorage

}

func (aManager *UserManager) setUsers() {
	aManager.users = make(map[string]*User)
}

func (aManager *UserManager) startMailer() (err error) {
	aManager.mailer, err = MailHandling.NewMailer()

	return err

}