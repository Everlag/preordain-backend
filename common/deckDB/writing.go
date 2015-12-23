package deckDB

import (
	"time"

	"fmt"

	"github.com/jackc/pgx"
)

// Sends a single event to the remote server
func SendEvent(pool *pgx.ConnPool, e *Event) error {
	
	tx, err := pool.Begin()
	if err != nil {
		return ConnError
	}

	// We can exit anytime before tx.Commit is called
	// and avoid any changes to the db
	defer tx.Rollback()

	fmt.Println("sending event", e.EventID)

	// Insert the meta event record
	err = insertEvent(tx, e)
	if err !=nil{
		return err
	}

	fmt.Println("Sent event")


	// Send each deck up
	for _, d := range e.Decks {

		fmt.Println("Sending", d.Name, " - " ,d.Player)

		err = insertDeck(tx, d, e.EventID)
		if err != nil {
			return err
		}

		fmt.Println("Sent", d.Name, " - " ,d.Player)


	}

	tx.Commit()

	return nil

}

// Inserts a deck into the db using a passed transaction
func insertDeck(tx *pgx.Tx, d *Deck, EventID string) (err error) {

	// Insert the meta row
	_, err = tx.Exec(deckInsert, d.Name, d.Player,
		d.DeckID, EventID)
	if err!=nil {
		return err
	}

	// Insert the contents of the deck
	for _, c:= range d.Maindeck{
		fmt.Println("Sending", c.Name)
		err = insertCard(tx, c, d.DeckID, false)
		fmt.Println("Sent", c.Name)
	}
	for _, c:= range d.Sideboard{
		fmt.Println("Sending", c.Name)
		err = insertCard(tx, c, d.DeckID, true)
		fmt.Println("Sent", c.Name)
	}

	return err

}

// Inserts a deck into the db using a passed transaction
//
// Decks contain the necessary metadata to populate the
// row of event data
func insertEvent(tx *pgx.Tx, e *Event) (err error) {

	_, err = tx.Exec(eventInsert, e.Name, e.EventID,
		time.Time(e.Happened))

	return err

}

// Inserts a deck into the db using a passed transaction
func insertCard(tx *pgx.Tx, c *Card,
	deckid string, sideboard bool) (err error) {

	_, err = tx.Exec(cardInsert,
		c.Name, c.Quantity,
		sideboard,
		deckid)

	return err

}