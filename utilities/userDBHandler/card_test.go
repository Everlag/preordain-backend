package userDB

import(

	"testing"

	"time"

)

var CardsPerCollection int = 10

// Add a collection to each user then try
// to add some cards to that collection. Tests whether
// they appear in just the history.
//
// This lets us JUST generate a set of random cards
// as each one represents a unique historical item.
func TestCardsHistory(t *testing.T) {
	t.Parallel()

	users, keys, collections, contents:= addSomeCards(t)

	time.Sleep(testSleepTime)
	

	var user string
	var key []byte
	var collection string

	for i := 0; i < testCount; i++ {
		
		user = users[i]
		collection = collections[i]
		key = keys[i]


		acquired, err:= GetCollectionHistory(pool, key, user, collection)
		if err!=nil {
			t.Fatal(err)
		}

		if !equalCardContents(contents[i], acquired, t){
			t.Fatal("Contents of collections did not match for", user)
		}
	}

}

// Adds a collection to each user then try to add some cards to that
// collection. Add a card and then change it multiple times to ensure
// the latest version is in the collection
func TestCardsContents(t *testing.T) {
	t.Parallel()
	
	user:= randString(int(randByte()) % 31)
	key, err:= AddUser(pool, user, "bar", "foo")
	if err!=nil {
		t.Fatal("failed to add user ", err)
	}
	collection:= randString(int(randByte()))
	err = AddCollection(pool, key, user, collection)
	if err!=nil {
		t.Fatal(err)
	}

	cards:= randomCards(1)

	err = AddCards(pool, key, user, collection, cards)
	if err!= nil {
		t.Fatal(err)
	}

	transitions:= randInts(testCount)
	latestQuantity:= int32(transitions[len(transitions) - 1])
	templateCard:= cards[len(cards) - 1]

	for _, transition:= range transitions{
		templateCard.Quantity = int32(transition)
		templateCard.LastUpdate = randomTime()
		err = AddCards(pool, key, user, collection, []Card{templateCard})
		if err!= nil {
			t.Fatal(err)
		}

		// Wait for the db to catch up
		time.Sleep(stepSleepTime)
	}

	// Grab the entire collection
	acquired, err:= GetCollectionContents(pool, key, user, collection)
	if err!=nil {
		t.Fatal(err)
	}

	// Find the card.
	for _, c:= range acquired{
		if c.Name == templateCard.Name && c.Set == templateCard.Set &&
		   c.Comment == templateCard.Comment &&
		   c.Quality == templateCard.Quality {
			
		   	if c.Quantity != latestQuantity {
		   		t.Fatal("Latest quantity did not match")
		   	}else{
		   		return
		   	}

		}
	}

	t.Fatal("Couldn't find card")
}


func addSomeCards(t *testing.T) (users []string, keys [][]byte,
	collections[]string, contents [][]Card) {

	var user string
	var key []byte
	var collection string
	var cards []Card
	var err error
	for i := 0; i < testCount; i++ {
		// Each user has a random name of length < 256
		user = randString(int(randByte()) % 31)
		users = append(users, user)

		// They need a session key to add or look at collections
		key, err = AddUser(pool, user, "bar", "foo")
		if err!=nil {
			t.Fatal("failed to add user ", err)
		}

		keys = append(keys, key)
		
		// Wait for the db to catch up
		time.Sleep(stepSleepTime)

		// Add each collection using a sane session
		collection = randString(int(randByte()))
		collections = append(collections, collection)

		err = AddCollection(pool, key, user, collection)
		if err!=nil {
			t.Fatal(err)
		}

		// Add a bunch of cards to the collection
		cards = randomCards(CardsPerCollection)
		err = AddCards(pool, key, user, collection, cards)
		if err!= nil {
			t.Fatal(err)
		}

		contents = append(contents, cards)
		
	}

	return

	
}