package userDB

import(

	"testing"

	"time"

)

// Add some users and change the subs. Check they match
func TestSubAdd(t *testing.T) {
	t.Parallel()

	var users []string
	sessions:= make([][]byte, 0)
	var subs []string
	var custIDs []string
	var subIDs []string

	var user string
	var session []byte
	var sub string
	var custID string
	var subID string

	var err error


	for i := 0; i < testCount; i++ {

		// Add the user
		user = randString(int(randByte()))
		session, err = AddUser(pool, user, "bar", "foo")
		if err!=nil {
			t.Fatal("failed to add user ", err)
		}

		// Change the sub only if we choose one different from
		// the default.
		sub = randomElement(SubTiers)
		for sub == DefaultSubLevel{
			sub = randomElement(SubTiers)
		}
		custID = randString(int(randByte()))
		subID = randString(int(randByte()))
		err = ModSub(pool, user, sub, custID, subID, session)
		if err!=nil {
			t.Fatal("failed to change add sub", err)
		}

		users = append(users, user)
		sessions = append(sessions, session)
		subs = append(subs, sub)
		custIDs = append(custIDs, custID)
		subIDs = append(subIDs, subID)

	}

	time.Sleep(testSleepTime)
	
	var s *Subscription
	for i := 0; i < testCount; i++ {
		
		user = users[i]
		session = sessions[i]
		sub = subs[i]
		custID = custIDs[i]
		subID = subIDs[i]

		// Make sure the sub details stuck
		s, err = GetSub(pool, user, session)
		if s.Plan != sub ||
		   s.CustomerID != custID ||
		   s.SubID != subID{
		   	t.Fatal("Subscription details did not match uploaded")
		}
		
	}

}


// Add a finite amount of users and ensure that
func TestSubEffects(t *testing.T) {
	t.Parallel()

	var user string
	var session []byte
	var err error

	for sub, count:= range SubTiersToCollections{

		// Add the user
		user = randString(int(randByte()))
		session, err = AddUser(pool, user, "bar", "foo")
		if err!=nil {
			t.Fatal("failed to add user ", err)
		}

		// Switch them to the plan we desire
		err = ModSub(pool, user, sub, "42", "12", session)
		if err!=nil {
			if sub == DefaultSubLevel{
				// We shouldn't be able to as this is where they start at
				continue
			}
			t.Fatal("failed to change add sub", err)
		}

		// Make sure the sub details stuck
		u, err:= GetUser(pool, user)
		if err!=nil {
			t.Fatal("failed to get user", err)
		}
		if u.MaxCollections != int32(count){
			t.Fatal("failed to get right max coll count back!")
		}
		// If we got here then we changed to a non-default sub plan
		// and thus their max viewtime should always be strictly before
		// the default start time. 
		if u.Longestview != noTimeLimit {
			t.Fatal("failed to get right longestview back")
		}

	}
}