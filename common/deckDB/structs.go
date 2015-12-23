package deckDB

import(

	"fmt"

	"bytes"

)

type Card struct {
	Name     string
	Quantity int
}

type Event struct{
	// mtgtop8 event id
	Name, EventID string

	Happened Timestamp

	Decks []*Deck
}

// A single deck that gets transformed into one meta row
// and len(Maindeck) + len(Sideboard) cheap card rows
type Deck struct {
	// mtgtop8 deck id
	Name, DeckID string

	// The specific person piloting the deck
	Player string

	// Deck contents
	Maindeck  []*Card
	Sideboard []*Card
}

// Turn a deck into a pretty string for easier debugging
func (deck *Deck) String() string {
	s := bytes.Buffer{}

	s.WriteString(fmt.Sprintf("Name %s\n", deck.Name))
	s.WriteString(fmt.Sprintf("Player %s\n", deck.Player))
	
	for _, c := range deck.Maindeck {
		s.WriteString(fmt.Sprintf("%d %s\n", c.Quantity, c.Name))
	}
	s.WriteString("\nSideboard\n")
	for _, c := range deck.Sideboard {
		s.WriteString(fmt.Sprintf("%d %s\n", c.Quantity, c.Name))
	}
	return s.String()
}