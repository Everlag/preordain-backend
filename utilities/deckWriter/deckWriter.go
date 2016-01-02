package main

import (
	"fmt"

	"os"

	"./../../common/deckDB"
	"./../../common/deckDB/nameNorm"

	// "io/ioutil"
	// "encoding/json"
)


// The format this scraper is concerned with
//
// Currently does nothing except tweak the fetcher
const Format string = "Modern"

type FetchError struct{
	Eventid string
	Err error
}

// Updates any deck and event info not already on the server
//
// The traditional err return value is a hard error, we couldn't continue
// while the primary return value is an event specific error which has
// some metadata saved for debugging problematic events.
func UpdateDecks() ([]FetchError, error) {
	pool, err:= deckDB.Connect()
	if err!=nil {
		return nil, fmt.Errorf("failed db connection", err)
	}

	eventList, err:= FetchEventList()
	if err!=nil {
		return nil, fmt.Errorf("failed event list fetch", err)
	}

	soft:= make([]FetchError, 0)

	// Fetch the event content and upload
	for _, e:= range eventList{

		// Fetch
		fresh, err:= FetchEvent(e)
		if err!=nil {
			f:= FetchError{
				Eventid: e,
				Err: fmt.Errorf("failed event fetch", err),
			}
			fmt.Println("error", f)
			soft = append(soft, f)
			continue
		}

		// Upload
		err = deckDB.SendEvent(pool, fresh)
		if err!=nil {
			f:= FetchError{
				Eventid: e,
				Err: fmt.Errorf("failed event upload", err),
			}
			fmt.Println("error", f)
			soft = append(soft, f)
			continue
		}
	}

	return soft, nil

}

func main() {

	// failures, err:= UpdateDecks()
	// if err!=nil {
	// 	fmt.Println("failed to update deckDB", err)
	// 	return
	// }

	// // Send it off to disk to peruse
	// serial, err:= json.Marshal(failures)
	// if err!=nil {
	// 	fmt.Println("", err)
	// 	os.Exit(1)
	// }
	// err = ioutil.WriteFile("nastyness.json", serial, 0777)
	// if err!=nil {
	// 	fmt.Println("", err)
	// 	os.Exit(1)
	// }

	// fmt.Println(failures)
	// return


	pool, err:= deckDB.Connect()
	if err!=nil {
		fmt.Println("failed db connection", err)
		os.Exit(1)
	}

	cards, err:= deckDB.GetArchetypeContents(pool, nameNorm.URTwin)
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cards)

	latest, err:= deckDB.GetArchetypeLatest(pool, nameNorm.URTwin)
	if err!=nil {
		fmt.Println(err)
		return
	}

	fmt.Println(latest)

	// eventList, err:= FetchEventList()
	// if err!=nil {
	// 	fmt.Println("failed event list fetch", err)
	// 	os.Exit(1)
	// }

	// // Fetch the event content and upload
	// for _, e:= range eventList{

	// 	// Fetch
	// 	fresh, err:= FetchEvent(e)
	// 	if err!=nil {
	// 		fmt.Println("failed event fetch", e, err)
	// 		os.Exit(1)
	// 	}

	// 	// Upload
	// 	err = deckDB.SendEvent(pool, fresh)
	// 	if err!=nil {
	// 		fmt.Println("failed event upload", err)
	// 		os.Exit(1)	
	// 	}
	// }

	// e, err:= FetchEvent("10803")
	// if err!=nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// err = deckDB.SendEvent(pool, e)
	// if err!=nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)	
	// }

	fmt.Println("oh baby")
	return

	// eventList, err:= FetchEventList()
	// if err!=nil {
	// 	fmt.Println("", err)
	// 	os.Exit(1)
	// }

	// // Fetch the event content
	// decks:= make([]*deckDB.Deck, 0)
	// failedEvents:= make([]string, 0)
	// for _, e:= range eventList{

	// 	freshDecks, err:= FetchEvent(e)
	// 	if err!=nil {
	// 		failedEvents = append(failedEvents, e)

	// 		continue
	// 	}

	// 	decks = append(decks, freshDecks...)
	// }

	// // Send it off to disk to peruse
	// serial, err:= json.Marshal(decks)
	// if err!=nil {
	// 	fmt.Println("", err)
	// 	os.Exit(1)
	// }
	// err = ioutil.WriteFile("wholeMeta.json", serial, 0777)
	// if err!=nil {
	// 	fmt.Println("", err)
	// 	os.Exit(1)
	// }

	// return

	// event, err:= FetchEvent("10980")
	// if err!=nil {
	// 	fmt.Println("", err)
	// 	os.Exit(1)
	// }

	// fmt.Println(event.Name)

	// return

	// deck, err:= FetchDeck("262514")
	// if err!=nil {
	// 	fmt.Println("uh oh", err)
	// 	os.Exit(1)
	// }

	// fmt.Println(deck.Maindeck[0])
}