package UserStructs

import (
	"io/ioutil"
	"os"
	"strings"

	"testing"

	"time"

	//"fmt"
)

const testDir string = "testEnv"

const testManagerSuffix string = "testingManager"

//static testing paramaters for users
var names = [...]string{"John", "Jane", "Ted", "Paul"}
var emails = [...]string{
	"John@gmail.com", "Jane@hotmail.com",
	"Ted@gmail.com", "Paul@outlook.com",
}

var passwords = [...]string{
	"coolcats$3", "jazz is great!2",
	"correct horse battery stapler21@",
	"Password1?!",
}

//static testing parameters for trades
var cardNames = [...]string{
	"Baleful Eidolon",
	"Horizon Scholar",
	"Burnished Hart",
	"Ashiok, Nightmare Weaver",
	"Controvert",
	"Martyr of Sands",
	"Fury of the Horde",
	"Skred",
	"Shackles",
	"Mox Diamond",
	"Ensnaring Bridge",
	"Ransack",
	"Sol Ring",
	"Sensei's Divining Top",
	"Seat of the Synod",
	"Scroll Rack",
}

//static collection names, one per user name
var collNames = [...]string{
	"Specs",
	"Standard Binder",
	"Trash",
	"Commander Staples",
}

const newUserPasswords string = "theBestPasswordIsTheOneLocatedInYourTests13!"

//creates a temp directory and moves the current directory to that
//location
//
//returns the directory we were originally inside
func setupTestDirectory(t *testing.T) string {

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get curent directory")
	}

	dir, err := ioutil.TempDir("", testDir)
	if err != nil {
		t.Fatalf("Failed to get temp directory")
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("Failed to go to temp directory")
	}

	return currentDir

}

//moves back to the provided directory for tests
//
//nukes the directory we were last in
func closeTestDirectory(properDir string, t *testing.T) {

	//get the temp directory we are in
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get curent directory")
	}

	//ensure we are in a temporary directory
	if !strings.Contains(currentDir, os.TempDir()) {
		t.Fatalf("Not in temp directory during attempted nuke!")
	}

	//move to the directory we want to be in
	err = os.Chdir(properDir)
	if err != nil {
		t.Fatalf("Failed to exit to proper directory")
	}

	//nuke the temp directory as thoroughly as possible
	err = os.RemoveAll(currentDir)

}

//moves back to the provided directory for tests
//
//nukes the directory we were last in
func closeTestDirectoryBench(properDir string, b *testing.B) {

	//get the temp directory we are in
	currentDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get curent directory")
	}

	//ensure we are in a temporary directory
	if !strings.Contains(currentDir, os.TempDir()) {
		b.Fatalf("Not in temp directory during attempted nuke!")
	}

	//move to the directory we want to be in
	err = os.Chdir(properDir)
	if err != nil {
		b.Fatalf("Failed to exit to proper directory")
	}

	//nuke the temp directory as thoroughly as possible
	err = os.RemoveAll(currentDir)

}

//creates a temp directory and moves the current directory to that
//location
//
//returns the directory we were originally inside
func setupTestDirectoryBench(b *testing.B) string {

	currentDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get curent directory")
	}

	dir, err := ioutil.TempDir("", testDir)
	if err != nil {
		b.Fatalf("Failed to get temp directory")
	}

	err = os.Chdir(dir)
	if err != nil {
		b.Fatalf("Failed to go to temp directory")
	}

	return currentDir

}

func getFakeTrades() []Trade {
	cards := make([]OwnedCard, len(cardNames))
	trades := make([]Trade, len(cards))
	for i := 0; i < len(cards); i++ {
		//each trade contains every card created up to that point
		cards[i] = CreateCard(cardNames[i], "aSet",
			12, 2,
			"French")
		tradeContents := make([]OwnedCard, i)
		copy(tradeContents, cards)

		// It is impossible for an invalid trade to be created here
		trades[i], _ = CreateExistingTrade(tradeContents,
			time.Now().UTC().Unix(),
			"this is a grand comment with a new: "+cardNames[i])
	}
	return trades
}

//performs a sufficiently deep comparison of two trades to determine if they
//are effectively equal
func tradesEqual(tradeA, tradeB Trade) bool {
	contentsA, contentsB := tradeA.Transaction, tradeB.Transaction

	if tradeA.Comment != tradeB.Comment ||
		tradeA.TimeStamp != tradeB.TimeStamp ||
		tradeA.Revoked != tradeB.Revoked {
		return false
	}

	if len(contentsA) != len(contentsB) {
		return false
	}

	for i := 0; i < len(contentsA); i++ {

		if contentsA[i].Name != contentsB[i].Name ||
			contentsA[i].Set != contentsB[i].Set ||
			contentsA[i].Quantity != contentsB[i].Quantity {

			return false

		}
	}

	return true
}

//Tests that trades are stored in the correct quantities right after addition
//and following a close-open cycle.
func TestTrades(t *testing.T) {

	properDir := setupTestDirectory(t)

	//setup the trades
	trades := getFakeTrades()

	//create the manager
	NewUserManager(testManagerSuffix)

	//open the manager
	aManager, err := ReacquireManager(testManagerSuffix)
	if err != nil || aManager.Suffix != testManagerSuffix {
		t.Fatalf("Failed to reacquire manager")
	}

	//add a few new users
	for i, aName := range names {

		_, err := aManager.AddUser(aName, emails[i], passwords[i], 2)
		if err != nil {
			t.Fatalf("Failed to add a user ", err)
		}

	}

	//add the trades to the users
	for i, aName := range names {

		//get a session
		aSession, err := aManager.GetNewSession(aName, passwords[i])
		if err != nil {
			t.Fatalf("Failed to acquire session, ", err)
		}

		aCollName := collNames[i]

		//add a collection to each user
		err = aManager.NewCollection(aName, aCollName, aSession)
		if err != nil {
			t.Fatalf("Failed to add collection", err)
		}

		//add the entire section of trades to the user
		for a := 0; a < len(trades); a++ {
			aManager.AddTrade(aName, aCollName, aSession, trades[a])
		}

		//ensure they were actually written to first layer cache
		aColl, err := aManager.GetCollection(aName, aCollName, aSession)
		if err != nil {
			t.Fatalf("Failed to get collection")
		}

		if len(aColl.ModifyHistory) != len(trades) {
			t.Fatalf("Trade size doesn't match pre persistence!")
		}

		//compare the list of trades to the ones we set for this user
		for a := 0; a < len(trades); a++ {
			if !tradesEqual(aColl.ModifyHistory[a],
				trades[a]) {
				t.Fatalf("Trades did not persist")
			}
		}

	}

	aManager.Close()

	//test to ensure that the users persist across manager closes
	aManager, err = ReacquireManager(testManagerSuffix)
	if err != nil || aManager.Suffix != testManagerSuffix {
		t.Fatalf("Failed to reacquire manager")
	}

	//test to ensure they are there after being sent to disk
	for i, aName := range names {

		//get a session
		aSession, err := aManager.GetNewSession(aName, passwords[i])
		if err != nil {
			t.Fatalf("Failed to acquire session, ", err)
		}

		aCollName := collNames[i]

		//decode the data
		aColl, err := aManager.GetCollection(aName, aCollName, aSession)
		if err != nil {
			t.Fatalf("Failed to get collection")
		}
		
		//compare the list of trades to the ones we set for this user
		if len(aColl.ModifyHistory) != len(trades) {
			t.Fatalf("Trade size doesn't match post persistence!")
		}

		for a := 0; a < len(trades); a++ {
			if !tradesEqual(aColl.ModifyHistory[a],
				trades[a]) {
				t.Fatalf("Trades did not persist")
			}
		}

	}

	time.Sleep(time.Duration(2) * time.Second)
	closeTestDirectory(properDir, t)

}

//Tests to ensure serialization is successful and that the db returns correctly
func TestPasswordChange(t *testing.T) {

	properDir := setupTestDirectory(t)

	//create the manager
	NewUserManager(testManagerSuffix)

	//open the manager
	aManager, err := ReacquireManager(testManagerSuffix)
	if err != nil || aManager.Suffix != testManagerSuffix {
		t.Fatalf("Failed to reacquire manager")
	}

	//add a few new users
	for i, aName := range names {

		_, err := aManager.AddUser(aName, emails[i], passwords[i], 2)
		if err != nil {
			t.Fatalf("Failed to add a user ", err)
		}
	}

	//add the reset tokens
	for _, aName := range names {

		//try to change the passwords before actually setting an initial token
		_, err = aManager.ChangePassword(aName,
			"aPoorToken", newUserPasswords)
		if err == nil {
			t.Fatalf("Successful password change using incorrect token, no initial token")
		}

		//set up a valid use token
		err := aManager.GetPasswordResetToken(aName)
		if err != nil {
			t.Fatalf("Failed to get password reset token")
		}

		//ensure it sticks in the first cache layer
		aUser, err := aManager.getUser(aName)
		if err != nil {
			t.Fatalf("Invalid password reset token")
		}

		//change the user's password to a constant
		_, err = aManager.ChangePassword(aName,
			aUser.PasswordResetToken.Key, newUserPasswords)
		if err != nil {
			t.Fatalf("Failed to change password using correct token")
		}

		//try again, the password should be invalid this time
		_, err = aManager.ChangePassword(aName,
			aUser.PasswordResetToken.Key, newUserPasswords)
		if err == nil {
			t.Fatalf("Successful password change using incorrect token")
		}

	}

	aManager.Close()

	//test to ensure that the users persist across manager closes
	aManager, err = ReacquireManager(testManagerSuffix)
	if err != nil || aManager.Suffix != testManagerSuffix {
		t.Fatalf("Failed to reacquire manager")
	}

	//test to ensure that password changes exist across resets
	for _, aName := range names {

		//get a session
		_, err := aManager.GetNewSession(aName, newUserPasswords)
		if err != nil {
			t.Fatalf("Failed to acquire session with new password, ", err)
		}

	}

	time.Sleep(time.Duration(2) * time.Second)
	closeTestDirectory(properDir, t)

}

//Tests to ensure serialization is successful and that the db returns correctly
func TestPersistence(t *testing.T) {

	properDir := setupTestDirectory(t)

	//create the manager
	NewUserManager(testManagerSuffix)

	//open the manager
	aManager, err := ReacquireManager(testManagerSuffix)
	if err != nil || aManager.Suffix != testManagerSuffix {
		t.Fatalf("Failed to reacquire manager")
	}

	//add a few new users
	for i, aName := range names {

		_, err := aManager.AddUser(aName, emails[i], passwords[i], 2)
		if err != nil {
			t.Fatalf("Failed to add a user ", err)
		}
	}

	//test to ensure they are there after being added immediately
	for i, aName := range names {

		aUser, err := aManager.getUser(aName)
		if err != nil {
			t.Fatalf("Failed to find a user ", err)
		}

		if aUser.Name != aName || aUser.Email != emails[i] {
			t.Logf("user data: ", aUser.Name, aUser.Email)
			t.Fatalf("User not being stored properly")
		}

	}

	aManager.Close()

	//test to ensure that the users persist across manager closes
	aManager, err = ReacquireManager(testManagerSuffix)
	if err != nil || aManager.Suffix != testManagerSuffix {
		t.Fatalf("Failed to reacquire manager")
	}

	//test to ensure they are there after being sent to disk
	for i, aName := range names {

		aUser, err := aManager.getUser(aName)
		if err != nil {
			t.Fatalf("Failed to find a user after persistenc")
		}

		if aUser.Name != aName || aUser.Email != emails[i] {
			t.Fatalf("User not being stored properly")
		}

	}

	time.Sleep(time.Duration(2) * time.Second)
	closeTestDirectory(properDir, t)

}

//names per runOver
const nameQuantityPerRun int = 1000

//runs per benchmark iteration
const runsOver int = 1

func generateRandomUsers(runs int) (names, emails, passwords []string) {
	quantity := runs * nameQuantityPerRun
	names = make([]string, quantity)
	emails = make([]string, quantity)
	passwords = make([]string, quantity)

	for i := 0; i < quantity; i++ {
		names[i] = randString(30)
		emails[i] = randString(40)
		passwords[i] = randString(50) + "!3"
	}

	return
}

//how long it takes to add nameQuantity names to the database
//
//Note that adds are likely bottlenecked by scrypt being used to derive keys
func BenchmarkAdd(b *testing.B) {

	for n := 0; n < b.N; n++ {
		b.StopTimer()
		properDir := setupTestDirectoryBench(b)

		//create the manager
		NewUserManager(testManagerSuffix)

		//open the manager
		aManager, err := ReacquireManager(testManagerSuffix)
		if err != nil || aManager.Suffix != testManagerSuffix {
			b.Fatalf("Failed to reacquire manager")
		}

		//create a dataset
		names, emails, passwords := generateRandomUsers(runsOver)
		//start measuring time
		b.StartTimer()

		indexMax := runsOver * nameQuantityPerRun

		for i := 0; i < indexMax; i++ {
			aSession, err := aManager.AddUser(names[i], emails[i], passwords[i], 2)
			if err != nil {
				b.Fatalf("Failed to add user")
			}
			if len(aSession) == 0 {
				b.Fatalf("Failed to get valid session")
			}
		}

		//stop measuring time
		b.StopTimer()
		time.Sleep(time.Duration(2) * time.Second)
		//clean up
		closeTestDirectoryBench(properDir, b)
		b.StartTimer()
	}

}

//how long it takes to acquire nameQuantityPerRun users from the DB with a
//cold in memory cache
//
//setup is effectively the benchmark for adding users so it takes awhile
func BenchmarkGetDBCold(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		properDir := setupTestDirectoryBench(b)

		//create the manager
		NewUserManager(testManagerSuffix)

		//open the manager
		aManager, err := ReacquireManager(testManagerSuffix)
		if err != nil || aManager.Suffix != testManagerSuffix {
			b.Fatalf("Failed to reacquire manager")
		}

		//create a dataset
		names, emails, passwords := generateRandomUsers(runsOver)

		indexMax := runsOver * nameQuantityPerRun

		for i := 0; i < indexMax; i++ {
			aSession, _ := aManager.AddUser(names[i], emails[i], passwords[i], 2)
			if len(aSession) == 0 {
				b.Fatalf("Failed to get valid session")
			}
		}

		//start measuring time
		b.StartTimer()
		for i := 0; i < indexMax; i++ {
			aUser, err := aManager.getUser(names[i])
			if err != nil {
				b.Fatalf("Failed to find a user after persistenc")
			}

			if aUser.Name != names[i] || aUser.Email != emails[i] {
				b.Fatalf("User not being stored properly")
			}
		}

		//stop measuring time
		b.StopTimer()
		time.Sleep(time.Duration(2) * time.Second)
		//clean up
		closeTestDirectoryBench(properDir, b)
		b.StartTimer()
	}

}
