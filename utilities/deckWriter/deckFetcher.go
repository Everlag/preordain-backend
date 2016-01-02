package main

import (
	"fmt"

	"./../../common/deckDB/deckData"

	"net/http"
)

// Fetches an mwDeck formatted deck for a given mtgtop8 deckID
// and parses it into a usable structure
func FetchDeck(id string) (*deckData.Deck, error) {

	loc := fmt.Sprintf("http://mtgtop8.com/export_files/deck%s.mwDeck", id)

	resp, err := http.Get(loc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch decklist", err)
	}

	defer resp.Body.Close()
	deck, err := NewDeck(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse decklist", err)
	}

	// Decorate the deck with its id
	deck.DeckID = id

	return deck, nil
}