package main

import (
	"os"
	"testing"

	"fmt"

	"io/ioutil"
	"encoding/json"

	"path/filepath"

	"reflect"

	"./../../common/deckDB"
)

const testLoc string = "testDecks"

// Deck cases are located in testDecks/ and are named as
// case.mwDeck.txt for deck to parse
// case.json for the serialized reference
var deckCases = []string{"vanilla", "splitCards", "aether"}

func TestParser(t *testing.T) {

	// Check each case
	for _, d:= range deckCases{
		err:= deckCase(d)
		if err!=nil {
			t.Error(err, d)
		}
	}

}

func deckCase(d string) error {
	
	// Fetch reference data structure
	reference, err:= getDeckReference(d)
	if err!=nil {
		return fmt.Errorf("failed to acquire reference", err)
	}

	// Parse the deck
	rawName:= fmt.Sprintf("%s.mwDeck.txt", d)
	rawLoc:= filepath.Join(testLoc, rawName)	
	raw, err:= os.Open(rawLoc)
	if err!=nil {
		return fmt.Errorf("failed to open deck test", err)
	}
	parsed, err:= NewDeck(raw)
	if err!=nil {
		return err
	}

	if !deckEquality(reference, parsed) {
		fmt.Println(parsed.String())
		return fmt.Errorf("reference and parsed didn't match")
	}

	return nil
}

func deckEquality(a, b *deckDB.Deck) bool {
	return reflect.DeepEqual(a, b)
}


// Get a reference value set for a deck
func getDeckReference(deck string) (*deckDB.Deck, error) {
	
	refName:= fmt.Sprintf("%s.json", deck)
	refLoc:= filepath.Join(testLoc, refName)

	var ref deckDB.Deck

	raw, err:= ioutil.ReadFile(refLoc)
	if err!=nil {
		return nil, err
	}
	err = json.Unmarshal(raw, &ref)
	if err!=nil {
		return nil, err
	}

	return &ref, nil
}

func ParseDecklist(filename string) (*deckDB.Deck, error) {
	file, err := os.Open("sampledata/" + filename)
	if err != nil {
		return nil, err
	}
	deck, err := NewDeck(file)
	if err != nil {
		return nil, err
	}
	if CountCards(deck.Maindeck) < 60 {
		return nil, fmt.Errorf("Maindeck less than 60 cards")
	}
	if CountCards(deck.Sideboard) > 15 {
		return nil, fmt.Errorf("Sideboard more than 15 cards")
	}
	return deck, nil
}

func CountCards(deck []*deckDB.Card) (count int) {
	for _, c := range deck {
		count += c.Quantity
	}
	return count
}