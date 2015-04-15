package userDB

import(

	"testing"

	"time"

)

// Add some sessions to the remote db
// then test each one for the return.
// Following that remove each session.
func TestSessions(t *testing.T) {
	t.Parallel()

	var users []string
	var keys [][]byte

	var user string
	var key []byte
	var err error
	for i := 0; i < testCount; i++ {
		// Each user has a random name of length < 256
		user = randString(int(randByte()))
		users = append(users, user)

		key, err = AddUser(pool, user, "bar", "foo")
		if err!=nil {
			t.Fatal("failed to add user ", err)
		}

		keys = append(keys, key)
	}

	time.Sleep(testSleepTime)
	
	for i := 0; i < testCount; i++ {
		
		user = users[i]
		key = keys[i]

		err = SessionAuth(pool, user, key)
		if err!=nil {
			t.Fatal("failed to authenticate the session", err)
		}

		err = Logout(pool, user, key)
		if err!=nil {
			t.Fatal("failed to logout", err)
		}
	}

}

// Add some sessions to the remote db
// then test each one for the return.
func TestResets(t *testing.T) {
	t.Parallel()

	var users []string
	var keys []string

	var user string
	var key string
	var err error
	for i := 0; i < testCount; i++ {
		// Each user has a random name of length < 256
		user = randString(int(randByte()))
		users = append(users, user)

		_, err = AddUser(pool, user, "bar", "foo")
		if err!=nil {
			t.Fatal("failed to add user ", err)
		}

		key, err = RequestReset(pool, users[i])
		if err!=nil {
			t.Fatal(err)
		}

		keys = append(keys, key)
	}

	time.Sleep(testSleepTime)
	
	for i := 0; i < testCount; i++ {
		
		user = users[i]
		key = keys[i]

		err = ValidateReset(pool, user, key)
		if err!=nil {
			t.Fatal(err)
		}
	}

}