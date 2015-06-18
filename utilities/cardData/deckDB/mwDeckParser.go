package deckData

import(

	"fmt"
	
	"net/http"
	"io/ioutil"
	
	"strings"
	"strconv"

	"time"

)

var commentedLineError = fmt.Errorf("comment, not a card")

type mwDeck []mwCard

// Returns how many copies a deck plays of a card
// with a given name in their mainboard
func (deck *mwDeck) Mainboard(card string) int {
	for _, c:= range *deck{
		if c.Name == card && !c.Sideboard {
			return c.Quantity
		}
	}

	return 0
}

// Returns how many copies a deck plays of a card
// with a given name in their sideboard
func (deck *mwDeck) Sideboard(card string) int {
	for _, c:= range *deck{
		if c.Name == card && c.Sideboard {
			return c.Quantity
		}
	}

	return 0
}

// Returns how many copies a deck plays of a card
// with a given name
//
// Returns 0 if not present
func (deck *mwDeck) Copies(card string) int {
	count:= 0

	for _, c:= range *deck{
		if c.Name == card {
			count+= c.Quantity
		}
	}

	return count
}

type mwCard struct{
	Name string
	Quantity int
	Sideboard bool
}

// Acquires a mwDeck from the remote location, parses it,
// and returns the contents in the mwCard form
func getDeck(loc string) (mwDeck, error) {
	
	fmt.Print("\t fetching deck")

	start:= time.Now()

	response, err:= http.Get(loc)
	if err!=nil {
		return nil, err
	}

	end:= time.Now()

	fmt.Println(" ", end.Sub(start))

	defer response.Body.Close()

	rawDeck, err:= ioutil.ReadAll(response.Body)
	if err!=nil {
		return nil, err
	}

	return parseDeck(string(rawDeck))

}

// Parses an mwDeck formatted deck if possible
//
// An mwDeck follows the following format for our purposes:
//  Comments are denoted with a //in the start of the line
//  Each non comment line follows: 'SB: Quantity [Set] Card Name'
//  where SB: is an optional flag to denote a card in the sideboard 
func parseDeck(deck string) (mwDeck, error) {
	
	// Normalize to unix line endings
	deck = strings.Replace(deck, "\n\r", "\n", -1)

	// Remove the set names as we don't need those
	deck = ripAllBetween(deck, "[", "]")

	// Split on newlines
	lines:= strings.Split(deck, "\n")
	
	contents:= make([]mwCard, 0)
	
	for _, l:= range lines{

		card, err:= parseDeckLine(l)
		if err!=nil {
			if err == commentedLineError {
				continue
			}
			return nil, err
		}

		contents = append(contents, card)
	}

	return contents, nil

}

// Parses a line of a decklist assuming all set data has been removed
func parseDeckLine(l string) (mwCard, error) {

	sideboard:= false
	quantity:= 0
	var err error

	// Ignore commented lines or line with insufficient length
	//
	// The length requirement is basicially arbitrary but should be sufficient
	// for screening out content-less lines
	if strings.HasPrefix(l, "//") || len(l) < 5 {
		return mwCard{}, commentedLineError	
	}

	// Normalize as convention is to have maindeck prefixed by one tab
	l = strings.TrimSpace(l)

	// Check for prescence of sideboard flag
	if strings.HasPrefix(l, "SB:"){
		sideboard = true
		// Renormalize
		l = strings.Replace(l, "SB:", "", -1)
		l = strings.TrimSpace(l)
	}

	// Quantity can only ever be a one digit or two digit number
	consideredDigits:= l[:2]
	quantity, err = strconv.Atoi(consideredDigits)
	if err != nil {
		// If a two digit number was invalid
		// then we grab the first character
		consideredDigits = l[:1]
		quantity, err = strconv.Atoi(consideredDigits)
		if err!=nil {
			// A failed conversion means we skip this card
			return mwCard{}, fmt.Errorf("failed to parse quantity")
		}
	}

	// Normalize to leave only the card name remaining
	l = strings.Replace(l, consideredDigits, "", -1)
	l = strings.TrimSpace(l)

	return mwCard{
		Name: l,
		Quantity: quantity,
		Sideboard: sideboard,
	}, nil
}