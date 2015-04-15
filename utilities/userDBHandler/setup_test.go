// Testing both the functionality of the package and the usability of the schema
//
// Higher level tests overlap with lower level tests purposefully.
//
// Tests are split into two type: error testing and bulk testing.
package userDB

import(

	"testing"

	"os"
	"fmt"
	"time"

	"math/rand"

	"github.com/jackc/pgx"
)

const testCount int = 30
const testSleepTime time.Duration = time.Duration(500) * time.Millisecond
const stepSleepTime time.Duration = time.Duration(2) * time.Millisecond

var pool *pgx.ConnPool

// Wraps every test we do so we can just pass a single connpool around
//
// Bonus: t.Parallel and a single pool with sufficient testCount means
//        we get a decent indicatication of burst production performance needs
func TestMain(m *testing.M){

	if pool == nil {
	
		var err error
		pool, err = Connect()
		if err!=nil {
			fmt.Println("encountered error initializing connection pool,", err)
			os.Exit(1)
		}

	}

	os.Exit(m.Run())

}

var cardNames = []string{
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

var setNames = []string{
	"Champions of Kamigawa",
	"Legends",
	"From the Vault: Twenty",
	"Mirrodin",
	"Scars of Mirrodin",
	"Return to Ravnica",
}

var Qualities = []string{
	"NM",
	"LP",
	"HP",
}

var Langs = []string{
	"EN",
	"ZH-HANS",
	"ZH-HANT",
	"FR",
	"IT",
	"DE",
	"KO",
	"JA",
	"PT",
	"RU",
	"ES",
}

// Generates a random card
func randomCard() Card {

	return  Card{
		Name: randomElement(cardNames),
		Set: randomElement(setNames),
		Quality: randomElement(Qualities),
		Lang: randomElement(Langs),
		Comment: randString(int(randByte()) % 20),
		Quantity: int32(rand.Intn(6)),
		LastUpdate: randomTime(),
	}
}

// Generates some number of random cards
func randomCards(count int) []Card {
	cards:= make([]Card, count)

	for i := 0; i < count; i++ {
		cards[i] = randomCard()
	}

	return cards
}

// A random time offset up to 4 billion seconds from now
func randomTime() time.Time {
	
	// This should be enough entropy
	randOffset:= int(randByte()) * int(randByte()) *
				int(randByte()) * int(randByte())

	randDuration:= time.Duration(randOffset) *  time.Minute

	// We have to round or the difference in accuracy between us and
	// postgres causes BIG issues
	return time.Now().Add(randDuration).Round(time.Second)

}

func randomElement(elements []string) string {
	
	return elements[rand.Intn(len(elements))]

}

func randInts(count int) []int {
	return rand.Perm(count)
}

// Tells us if two card slices have equal contents
func equalCardContents(a, b []Card, t *testing.T) bool {

	aMap:= make(map[Card]bool)
	bMap:= make(map[Card]bool)
	
	// Ignore times
	for _, c:= range a{
		aMap[c] = true
	}
	for _, c:= range b{
		bMap[c] = true
	}

	if len(a) != len(b) {
		t.Log("unequal lengths", len(a), "=/=" ,len(b))
		return false
	}

	for card:= range aMap{
		_, ok:= bMap[card]
		if !ok {
			t.Log(card, " was not present")
			return false
		}
	}

	return true

}