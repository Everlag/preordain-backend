package main

import (
	"fmt"

	"os"

	"./../../common/deckDB"

	// "io/ioutil"
	// "encoding/json"
)


// The format this scraper is concerned with
//
// Currently does nothing except tweak the fetcher
const Format string = "Modern"

func main() {


	

	return
	// e, err:= FetchEvent("11159")
	// if err!=nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// t, _:= e.Happened.MarshalJSON()

	// fmt.Println(string(t))
	// return

	pool, err:= deckDB.Connect()
	if err!=nil {
		fmt.Println("failed db connection", err)
		os.Exit(1)
	}


	eventList, err:= FetchEventList()
	if err!=nil {
		fmt.Println("failed event list fetch", err)
		os.Exit(1)
	}

	// Fetch the event content and upload
	for _, e:= range eventList{

		// Fetch
		fresh, err:= FetchEvent(e)
		if err!=nil {
			fmt.Println("failed event fetch", e, err)
			os.Exit(1)
		}

		// Upload
		err = deckDB.SendEvent(pool, fresh)
		if err!=nil {
			fmt.Println("failed event upload", err)
			os.Exit(1)	
		}
	}

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