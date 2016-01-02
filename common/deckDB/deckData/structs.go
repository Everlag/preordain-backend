package deckData
// A centralized Source for all structures that deckDB uses internally,
// takes as input, or returns.
//


import(

	"fmt"

	"bytes"

	"time"

)

// Structures used solely for output

// A deck with desirable metadata.
type TaggedDeck struct{
	Event string
	Happened Timestamp
	Deck *Deck
}


// Bidirectional structs

type Card struct {
	Name     string
	// Quantity can be summed, might as well
	// avoid future sanity loss
	Quantity int64
}

// Turn a deck into a pretty string for easier debugging
func (card *Card) String() string {
	return fmt.Sprintf("%s %d\n", card.Name, card.Quantity)
}

type Event struct{
	// mtgtop8 event id
	Name string
	EventID string `json:",omitempty"`

	Happened Timestamp

	Decks []*Deck `json:",omitempty"`
}

// A single deck that gets transformed into one meta row
// and len(Maindeck) + len(Sideboard) cheap card rows
type Deck struct {
	// mtgtop8 deck id
	Name string

	DeckID string `json:",omitempty"`

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

// Declare a wrapper for time.Time so we can
// marshal it to a timestamp rather than a string
type Timestamp time.Time

func (t Timestamp) MarshalJSON() ([]byte, error) {

	ts := time.Time(t).Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil

}