package userDB

import(

	"testing"

	"fmt"
	"time"

)

const CollectionCountPerUser int = 2

// Add a few collections to each user
// then try to read them back
func TestCollections(t *testing.T) {
	t.Parallel()

	var users []string
	var keys [][]byte
	var collections [][]string

	var user string
	var key []byte
	var collection string
	var err error
	for i := 0; i < testCount; i++ {
		// Each user has a random name of length < 256
		user = randString(int(randByte()))
		users = append(users, user)

		// They need a session key to add or look at collections
		key, err:= AddUser(pool, user, "bar", "foo")
		if err!=nil {
			t.Fatal("failed to add user ", err)
		}

		keys = append(keys, key)
		
		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		// Add each collection using a sane session
		userCollections:= make([]string, CollectionCountPerUser)
		collections = append(collections, userCollections)
		for k:= 0; k < CollectionCountPerUser; k++ {

			collection = randString(int(randByte()))

			err = AddCollection(pool, key, user, collection)
			if err!=nil {
				t.Fatal(err)
			}

			collections[i][k] = collection

		}
	}

	time.Sleep(testSleepTime)
	
	var coll *Collection
	var userCollections []string
	for i := 0; i < testCount; i++ {
		
		user = users[i]
		userCollections = collections[i]
		key = keys[i]

		for _, collName:= range userCollections{
			t.Log(user, collName)

			coll, err = GetCollectionMeta(pool, key, user, collName)
			if err!=nil {
				t.Fatal(err)
			}

			if coll.Name != collName {
				t.Fatal(fmt.Errorf("invalid collection returned"))
			}
		}
	}

}

// Tests to ensure a user is incapable of adding more than their alloted
// collections.
func TestCollPermissions(t *testing.T) {
	t.Parallel()

	user:= randString(int(randByte()))
	key, err:= AddUser(pool, user, "bar", "foo")
	if err!=nil {
		t.Fatal("failed to add user ", err)
	}

	collection:= randString(int(randByte()))

	err = AddCollection(pool, key, user, collection)
	if err!=nil {
		t.Fatal("valid collection was denied", err)
	}

	// History but not viewing should fail
	err = SetCollectionPrivacy(pool, key, user, collection,
		"Boots")
	if err == nil {
		t.Fatal("was allowed to set invalid permissions")
	}

	// Viewing but no history should work
	err = SetCollectionPrivacy(pool, key, user, collection,
		"Private")
	if err != nil {
		t.Fatal("failed to set valid permissions", err)
	}
	
	// No public access should work
	err = SetCollectionPrivacy(pool, key, user, collection,
		"History")
	if err != nil {
		t.Fatal("failed to set valid permissions", err)
	}

}

// Tests to ensure a user is incapable of adding a collection with a name
// longer than reasonable
func TestInvalidCollName(t *testing.T) {
	t.Parallel()

	user:= randString(int(randByte()))
	collection:= randString(int(randByte()) * 256)

	key, err:= AddUser(pool, user, "bar", "foo")
	if err!=nil {
		t.Fatal("failed to add user ", err)
	}

	err = AddCollection(pool, key, user, collection)
	if err == nil {
		t.Fatal(fmt.Errorf("collection name was too long and accepted"))
	}

}

// Tests to ensure a user is incapable of adding more than their alloted
// collections.
func TestInvalidCollCount(t *testing.T) {
	t.Parallel()

	user:= randString(int(randByte()))

	key, err:= AddUser(pool, user, "bar", "foo")
	if err!=nil {
		t.Fatal("failed to add user ", err)
	}

	// Set a static value for testing purposes
	err = SetMaxCollections(pool, user, 1)
	if err!=nil {
		t.Fatal("failed to set collection max", err)
	}

	// Two random collections
	collections:= []string{randString(int(randByte())),
		randString(int(randByte())),}

	err = AddCollection(pool, key, user, collections[0])
	if err!=nil {
		t.Fatal("valid collection was denied", err)
	}

	err = AddCollection(pool, key, user, collections[1])
	if err==nil {
		t.Fatal("collection beyond maximum was allowed")
	}
	
}