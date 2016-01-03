// Allows sane normalization of deck names that mtgtop8
// exposes. We have to handle some nasty shit.
//
// Begin must be called before the anything in the package is
// touched, this initializes package wide state that has
// to happen during runtime. 
package nameNorm

import (
	"./../deckData"

	"fmt"

	"strings"
	"unicode"

)

// Internal package state

// Our primary filter
var topLevel nameFilter


// Whether or not the package's internal state has
// been populated
func Ready() bool {
	return topLevel != nil	
}

// Starts our internal state.
func Begin() {

	// Consistent ordering of deck names
	// that we expose
	sortNames()

	// Mapping to and from mtgtop8 deck names
	// to the pleasant names we support
	populateTopLevel()
}

// Given a deck, performs a best effort attempt to give it a clean name
func Clean(d *deckData.Deck) error {

	if !Ready() {
		return fmt.Errorf("internal state not initialized, call Begin()!")
	}

	return topLevel.determine(d)
}

// Given an official name we support, attempts to turn it into
// every mtgtop8 name that could've produced it.
//
// This produces an InvertedTuple which allows for the necessary
// level of detail used when distinguishing one subarchetype from another
func Invert(s string) (archetypes []string,
						presentCards []string,
						excludedCards []string, err error) {
	
	if !Ready() {
		err = fmt.Errorf("internal state not initialized, call Begin()!")
		return
	}


	result:= topLevel.invert(s)
	if len(result) == 0 {
		err = fmt.Errorf("failed to find name")
		return
	}

	// Unpack the translated names to something we can
	// feed back to the caller
	archetypes = make([]string, 0)
	presentCards = make([]string, 0)
	excludedCards = make([]string, 0)
	for _, t:= range result{
		archetypes = append(archetypes, t.Name)
		if !t.Exclude {
			presentCards = append(presentCards, t.Card)
			continue
		}

		excludedCards = append(excludedCards, t.Card)
	}

	return

}

// Maps from a potentially dirty input string
// to a clean representation that ignores
// punctuation, whitespace, and case sensitivity
func Normalize(s string) string {
	return strings.Map(func(r rune) rune{
		if unicode.IsSpace(r) {
			return -1
		}
		if unicode.IsPunct(r) {
			return -1
		}
		return unicode.ToLower(r)
	}, s)
}