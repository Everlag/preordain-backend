package userDB

import(

	"testing"
	"time"
)

// Attempt the complete workflow of a single user.
// This handles every facet of usage.
func TestComplete(t *testing.T) {

	var err error

	var key []byte
	var acquired []Card
	var reset string

	for i := 0; i < 2; i++ {

		name:= randString(int(randByte()))
		email:= randString(int(randByte()))
		password:= randString(int(randByte()))
		collection:= randString(int(randByte()))
		cards:= randomCards(testCount)

		// Grab a new session
		key, err = AddUser(pool, name, email, password)
		if err != nil {
			t.Fatal("failed to add user ", err)
		}

		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		// Add a fresh collection
		err = AddCollection(pool, key, name, collection)
		if err!=nil {
			t.Fatal("failed to add a collection", err)
		}

		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		// Add some cards to that collection
		err = AddCards(pool, key, name, collection, cards)
		if err!= nil {
			t.Fatal("failed to add cards", err)
		}

		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		// Logout
		err = Logout(pool, name, key)
		if err!=nil {
			t.Fatal("failed to logout", err)
		}

		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		// Log back in
		key, err = Login(pool, name, password)
		if err!=nil {
			t.Fatal("failed to reacquire the user", err)
		}

		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		// Grab every change we made
		acquired, err = GetCollectionHistory(pool, key, name, collection)
		if err!=nil {
			t.Fatal("failed to get history back", err)
		}
		if !equalCardContents(acquired, cards, t){
			t.Fatal("history of collection did not match", err)
		}

		// Omitting checking the contents as we do that in another test
		// and it adds a lot of complication.

		// How forgetful!
		password = randString(int(randByte()))

		// Make sure we can't do anything with an incorrect password
		key, err = Login(pool, name, password)
		if err==nil {
			t.Fatal("logged in with invalid password", err)
		}

		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		key = []byte{randByte(), randByte(), randByte(),}

		// Make sure we can't do anything with an incorrect key
		acquired, err = GetCollectionHistory(pool, key, name, collection)
		if err==nil {
			t.Fatal("got history back with invalid key", err)
		}

		// Send a password reset request
		reset, err = RequestReset(pool, name)
		if err!=nil {
			t.Fatal("failed to get reset request", err)
		}

		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		err = ChangePassword(pool, name, password, reset)
		if err!=nil {
			t.Fatal("failed to change user password", err)
		}

		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		// Log back in with the new password
		key, err = Login(pool, name, password)
		if err!=nil {
			t.Fatal("failed to reacquire the user", err)
		}

		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		// Read the collection contents again!
		acquired, err = GetCollectionHistory(pool, key, name, collection)
		if err!=nil {
			t.Fatal("failed to get history back", err)
		}
		if !equalCardContents(acquired, cards, t){
			t.Fatal("history of collection did not match", err)
		}
		
	}


}