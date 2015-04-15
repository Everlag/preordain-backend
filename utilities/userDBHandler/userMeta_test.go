package userDB

import(

	"testing"

	"time"

)

// Adds a number of users with random passwords. Then attempts
// to login and collect a session.
func TestUsers(t *testing.T) {
	t.Parallel()

	var users []User
	var passwords []string

	var aUser User
	var aPass string
	var err error
	for i := 0; i < testCount; i++ {
		
		aUser = User{
			Name: randString(int(randByte())),
			Email: randString(int(randByte())),
		}
		users = append(users, aUser)

		aPass = randString(int(randByte()) + 10)
		passwords = append(passwords, aPass)

		// Ignore the session key that gets returned
		_, err = AddUser(pool, aUser.Name, aUser.Email, aPass)
		if err!=nil {
			t.Fatal("failed to add user ", err)
		}

	}

	time.Sleep(testSleepTime)

	for i := 0; i < testCount; i++ {
		
		_, err = Login(pool, users[i].Name, passwords[i])
		if err!=nil {
			t.Fatal("failed to reacquire the user", err)
		}

	}

}

func TestInvalidUsers(t *testing.T) {
	t.Parallel()

	var name string
	var email string
	var password string
	var err error

	// Invalid name
	name = randString(int(randByte()) + 280)
	email = randString(int(randByte()))
	password = randString(int(randByte()))

	_, err = AddUser(pool, name, email, password)
	if err == nil {
		t.Fatal("was to add user with invalid name", err)
	}
	
	// Invalid email
	name = randString(int(randByte()))
	email = randString(int(randByte()) + 280)
	password = randString(int(randByte()))

	_, err = AddUser(pool, name, email, password)
	if err == nil {
		t.Fatal("was able to add user with invalid email", err)
	}
	
	// Completely valid
	name = randString(int(randByte()))
	email = randString(int(randByte()))
	password = randString(int(randByte()))

	_, err = AddUser(pool, name, email, password)
	if err != nil {
		t.Fatal("failed to add user ", err)
	}
	

}