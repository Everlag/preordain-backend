package databaseHandling

import(
	
	//errors
	"fmt"

	//the storage backend
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"

	//allows us the ability to safely backup the database while running
	"sync"

	//determining when to snapshot
	"time"

	//for when something needs to be said
	"log"

	//cross-platform file paths
	"os"

	//writing metadata files
	"io/ioutil"

	//unmarshalling the metadata
	"encoding/json"

)

//the package comes with a few standard snapshot names
//
//there are 17 snapshot names predefined
var SnapshotNamesDefault  = [...]string{
	"alpha","beta","gamma","delta","epsilon",
	"zeta","phi","chi", "psi", "tau", "pi",
	"rho", "sigma", "omega", "eta", "lambda",
	}

//handles the abstractions of database usage for generic purposes
//
//a wrapped storage sits in a directory with the following contents
//generated:
//snapshots, the database, a json file with metadata, and a log
type WrappedStorage struct{

	Name string

	//we rotate databases so this contains the identifers for the file names
	//of each snapshot.
	SnapshotNames []string

	//provides an index into JournalSnapshotNames which allows us the ability
	//to acquire the most up to date user database
	LastJournalSnapshot int

	//when we last rotated the database into which we write our current data.
	LastSnapshotRotation int64

	//how long we work on a specific snapshot until switching to another
	SnapshotRotationTime int64

	//the open database where we store our user data
	//
	//this is leveldb database which means it is key-value so we store only one
	//version of each user per database.
	//
	//backups are rotated under the LastSnapshotRotation
	database *leveldb.DB

	//a RW mutex allows us to run a daemon that regularly checks if we need
	//to rotate the database and then does so in a safe manner that does
	//block but is reasonable.
	locker sync.RWMutex

	dbLogger *log.Logger

	//provides a way to ensure we run only a single snapshot daemon at a time
	daemonRunning bool

}

//runs a daemon in a goroutine that regularly checks if we need to rotate
//in a new database snapshot.
func (someStorage *WrappedStorage) runDaemon() {
	if !someStorage.daemonRunning {
		go someStorage.snapshotDaemon()
		someStorage.daemonRunning = true
	}

}

//regularly checks if the SnapshotRotationTime + LastSnapshotRotation
//is less than now, which means that we must snapshot the database safely
func (someStorage *WrappedStorage) snapshotDaemon() {
	
	for{

		snapshotTime:= someStorage.SnapshotRotationTime +
						someStorage.LastSnapshotRotation

		if snapshotTime < time.Now().UTC().Unix() {
			someStorage.snapshot()
		}

		time.Sleep(time.Duration(someStorage.SnapshotRotationTime)*time.Second)

	}

}

//snapshotting follows this procedure:
//1. Acquire locker Writer lock to prevent others from trying to work with
//   the db during the backup
//2. Close the db
//3. Write a copy of the database to disk.
//4. Reopen the db.
//5. Release the lock.
func (someStorage *WrappedStorage) snapshot() {
	
	someStorage.dbLogger.Println("Starting snapshot, acquiring lock")

	someStorage.locker.Lock()

	//close the db
	err:= someStorage.database.Close()
	if err!=nil {
		//try again
		err= someStorage.database.Close()
		if err!=nil {
			someStorage.dbLogger.Println("Failed to snapshot")
		}
	}

	loc:= someStorage.getDatabaseLocation()

	//increment the snapshot index
	someStorage.LastJournalSnapshot++
	if someStorage.LastJournalSnapshot > len(someStorage.SnapshotNames) - 1 {
		someStorage.LastJournalSnapshot = 0
	}

	//perform the copy
	cp(loc + "." + someStorage.SnapshotNames[someStorage.LastJournalSnapshot],
		loc)

	//reopen the database
	freshDatabase, err := leveldb.OpenFile(loc, nil)
	if (err!=nil){
		someStorage.dbLogger.Println("Failed to reopen database after snapshot")
	}else{
		//in the event that we fail to reopen the database, we block forever
		//
		//not an optimal solution for user experience but prevents further
		//potential data loss
		//once all is done, set the database to the current database
		someStorage.database = freshDatabase
		//update the db metadata
		someStorage.saveMetaData()
		//and release the lock if we didn't have an error

		someStorage.locker.Unlock()

		someStorage.dbLogger.Println("Successful snapshot, lock released")
	

	}

}

//returns the relative path to the database
func (someStorage *WrappedStorage) getDatabaseLocation() string {
	
	//the directory relative to the current the db resides in
	pathStart:= someStorage.Name + string(os.PathSeparator)

	//the actual name of the db file
	pathEnd:= someStorage.Name + ".leveldb"

	return pathStart + pathEnd

}

//save the metadata to disk which we require to be able to revive the
//storage.
func (someStorage *WrappedStorage) saveMetaData() error {
	//grab the json
	data, err := someStorage.ToJson()
	if err!=nil {
		return err
	}

	//save it
	loc := deriveMetaLocation(someStorage.Name)

	err = ioutil.WriteFile(loc, data, 0666)
	if err!=nil {
		return err
	}

	return nil

}

//populates the wrapped storage with the administrative sections it requires
//to work that are not persisted in the json metadata file
//
//This includes hooking up the actual database database
func (someStorage *WrappedStorage) populateEphemeral() error {
	
	name := someStorage.Name

	dbLogger := getLogger(name + string(os.PathSeparator) + name + ".log", name)

	intitalDBName := name + string(os.PathSeparator) + name + ".leveldb"
	freshDatabase, err := leveldb.OpenFile(intitalDBName, nil)
	if (err!=nil){
		return err
	}

	someStorage.dbLogger = dbLogger
	someStorage.database = freshDatabase

	return nil

}

//writes the provided value to the provided key safely
func (someStorage *WrappedStorage) WriteValue(key, val []byte) {
	someStorage.locker.RLock()
	someStorage.database.Put(key, val, nil)
	someStorage.locker.RUnlock()
}

//reads the provided value from the database and returns a slice
//completely safe to modify
//
//It returns ErrNotFound if the DB does not contain the key.
func (someStorage *WrappedStorage) ReadValue(key []byte) ([]byte, error) {
	someStorage.locker.RLock()
	result, err:= someStorage.database.Get(key, nil)
	someStorage.locker.RUnlock()
	
	return result, err

}

//returns the storage after deserializing the json meta file, populating
//it with ephemeral sections rebuilt, and reacquiring the db
func ReacquireWrappedStorage(name string) (*WrappedStorage, error) {
	var err error

	//grab the json metadata first
	loc:= deriveMetaLocation(name)
	jsonData, err:= ioutil.ReadFile(loc)
	if err!=nil {
		return nil, err
	}

	//attempt to deserialize it
	var someStorage WrappedStorage
	err = json.Unmarshal(jsonData, &someStorage)
	if err!=nil {
		return nil, err
	}

	//populate the ephemeral sections
	err = someStorage.populateEphemeral()
	if err!=nil {
		return nil, err
	}

	someStorage.runDaemon()

	return &someStorage, nil
}

//safely closes the database after saving the json meta file
//
//this permanently closes the database for this instance, if you want
//to access it again, use ReacquireWrappedStorage
func (someStorage *WrappedStorage) SafeClose() {
	var err error

	err =  someStorage.saveMetaData()
	if err!=nil {
		someStorage.dbLogger.Println("Failed to commit meta file")
	}

	//acquire the database lock.
	//
	//note that we never let it go as the database should never be accessed
	//though this instance ever again
	someStorage.locker.Lock()

	err = someStorage.database.Close()
	if err!=nil {
		//if we can't close, then we try again
		err = someStorage.database.Close()
		if err!=nil {
			//if we REALLY can't close, then we report it and fail
			someStorage.dbLogger.Println("Failed to safely close database")
		}
	}
}

//creates a new wrapped storage with the provided name, rotation time,
//and quantity of snapshots to hold before overwriting.
//
//performs basic sanity checks and will fail to create the database
//if one already exists
func NewWrappedStorage(name string,
	rotationTime int64, snapshotCount int) (*WrappedStorage, error) {

	var err error
	
	if (len(name) > 40){
		return nil, fmt.Errorf("Name too long")
	}

	if (snapshotCount > len(SnapshotNamesDefault) || 
		(snapshotCount < -1)){
		return nil, fmt.Errorf("Invalid snapshot count")
	}

	//acquire a database located in the name/ directory relative to the
	//current location
	intitalDBName := name + string(os.PathSeparator) + name + ".leveldb"
	DBOptions := &opt.Options{
		ErrorIfExist: true,
	}
	freshDatabase, err := leveldb.OpenFile(intitalDBName, DBOptions)
	if (err!=nil){
		return nil, err
	}

	dbLogger := getLogger(name + string(os.PathSeparator) + name + ".log",
						  name)

	var freshStorage WrappedStorage = WrappedStorage{
		Name: name,
		LastJournalSnapshot:0,
		LastSnapshotRotation: 0,
		SnapshotRotationTime: rotationTime,

		SnapshotNames: SnapshotNamesDefault[:snapshotCount],
		database: freshDatabase,
		dbLogger: dbLogger,
	}

	freshStorage.runDaemon()

	return &freshStorage, nil
}

func deriveMetaLocation(name string) (loc string) {
	
	loc= name + string(os.PathSeparator) + name + ".meta.json"
	return
}