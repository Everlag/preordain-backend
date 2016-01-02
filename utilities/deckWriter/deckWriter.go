package main

import (
	"fmt"

	"time"
	"os"

	"log"

	"./../../common/deckDB"

	"github.com/jackc/pgx"

	"github.com/joho/godotenv"

)


// How long to wait between full fetches
var WaitTime time.Duration = time.Hour * time.Duration(12)

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
func UpdateDecks(pool *pgx.ConnPool, l *log.Logger) ([]FetchError, error) {

	l.Println("fetching event list")

	eventList, err:= FetchEventList()
	if err!=nil {
		return nil, fmt.Errorf("failed event list fetch", err)
	}

	l.Println("fetched event list")

	soft:= make([]FetchError, 0)

	// Fetch the event content and upload
	for _, e:= range eventList{

		// Check if we already have the event in its entirety
		present, err:= deckDB.HasEvent(pool, e)
		if err!=nil {
			f:= FetchError{
				Eventid: e,
				Err: fmt.Errorf("failed event presence test", err),
			}
			l.Println("failed to test event presence", f)
			soft = append(soft, f)
			continue
		}

		// No need to fetch if we already have it!
		if present {
			continue
		}

		l.Println("fetching event", e)

		// Fetch
		fresh, err:= FetchEvent(e)
		if err!=nil {
			f:= FetchError{
				Eventid: e,
				Err: fmt.Errorf("failed event fetch", err),
			}
			l.Println("failed to fetch event", f)
			soft = append(soft, f)
			continue
		}

		l.Println("fetched event", e)
		l.Println("uploading event", e)

		// Upload
		err = deckDB.SendEvent(pool, fresh)
		if err!=nil {
			f:= FetchError{
				Eventid: e,
				Err: fmt.Errorf("failed event upload", err),
			}
			l.Println("failed to upload event", f)
			soft = append(soft, f)
			continue
		}

		l.Println("uploaded event", e)
	}

	return soft, nil

}

// Runs a fetch from mtgtop8
//
// Self throttles to ensure we aren't being abusive.
func mtgtop8(pool *pgx.ConnPool, l *log.Logger) {
	

	// Find the last time we updated
	last, err:= GetState()
	if err!=nil {
		l.Println("inital state not present, seeding")
		// No state, set some!
		err = SetState()
		if err!=nil {
			l.Fatalln("failed to set state", err)
		}
		// Get the newly formed state
		last, err = GetState()
		if err!=nil {
			l.Fatalln("failed to get state", err)
		}
	}

	// If we haven't waited at least WaitTime since the last fetch
	// we just return and get asked again later
	now:= time.Now()
	if now.Sub(last) < WaitTime {
		return
	}

	// Actually acquire the updates
	soft, err:= UpdateDecks(pool, l)
	if err!=nil {
		l.Println(err)
		return
	}

	// Soft errors can just be logged
	if len(soft) > 0 {
		l.Println(err)
	}

	// Attempt to set the next state to thr
	err = SetState()
	if err!=nil {
		l.Println(err)
	}

}

func RunDeckLoop(pool *pgx.ConnPool, l *log.Logger) {
	
	for{

		mtgtop8(pool, l)

		l.Println("sleeping for update, mtgtop8")
		time.Sleep(time.Duration(1) * time.Hour)

	}
}

func main() {

	// Grab our local config
	envError:= godotenv.Load("deckWriter.default.env")
	if envError!=nil {
		fmt.Println("failed to parse prices.default.env")
		os.Exit(1)
	}

	l:= GetLogger("deckWriter.log", "deckWriter")

	pool, err:= deckDB.Connect()
	if err!=nil {
		l.Fatalln("failed to connect to remote db", err)
	}

	RunDeckLoop(pool, l)
}