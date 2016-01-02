// mwDeck Parser courtesy of https://github.com/malthrin/mtg-aggregatedeck
//
// License is MIT, personally requested and given on reddit
package main

import (
	"io"
	"io/ioutil"

	"regexp"

	"fmt"

	"./../../common/deckDB/deckData"


	"golang.org/x/text/encoding/charmap"

	"strconv"
	"strings"
)

// Modified to accept non-specified sets rather
// than outputting a pair of brackets
var linePattern = regexp.MustCompile(`^[^0-9]*(?P<quantity>\d+) (\[.*\] )?(?P<name>.+)$`)
const createrPrefix string = "CREATOR :"
const namePrefix string = "NAME :"


const MainDeckSize int = 60

func NewCard(line string) (*deckData.Card, error) {
	match := linePattern.FindStringSubmatch(line)
	card := &deckData.Card{}
	if match == nil {
		return card, fmt.Errorf("Failed to parse line '%s'", line)
	}
	for i, group := range linePattern.SubexpNames() {
		if group == "quantity" {
			quantity, err := strconv.Atoi(match[i])
			if err != nil {
				return card, err
			}
			card.Quantity = int64(quantity)
		} else if group == "name" {
			card.Name = match[i]
		}
	}
	if len(card.Name) == 0 {
		return nil, fmt.Errorf("Could not parse card name from '%s'", line)
	} else if card.Quantity == 0 {
		return nil, fmt.Errorf("Could not parse card quantity from '%s'", line)
	} else {
		return card, nil
	}
}

func ignoreDecklistLine(line string) bool {
	if len(line) == 0 ||
		line == "Sideboard" ||
		strings.HasPrefix(line, "//") ||
		strings.HasPrefix(line, "#") {
		return true
	}
	return false
}

// Converts a creater formatted line into
// the actual player's name
//
// Form "// CREATOR: player name"
func handleCreatorLine(line string) string {
	if !strings.Contains(line, createrPrefix) {
		return "?"
	}

	// Check formatting
	creatorSplit:= strings.Split(line, createrPrefix)
	if len(creatorSplit) < 2 {
		return "?"
	}
	// Fetch and clean the creator
	return strings.TrimSpace(creatorSplit[1])
}

// Converts a name formatted line into
// the actual mtgtop8 rough archetype
//
// Form "// NAME: deck name"
func handleNameLine(line string) string {
	if !strings.Contains(line, namePrefix) {
		return "?"
	}

	// Check formatting
	nameSplit:= strings.Split(line, namePrefix)
	if len(nameSplit) < 2 {
		return "?"
	}
	// Fetch and clean the creator
	return strings.TrimSpace(nameSplit[1])
}

func NewDeck(r io.Reader) (*deckData.Deck, error) {
	decklist, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	maindeck := make([]*deckData.Card, 0)
	sideboard := make([]*deckData.Card, 0)
	count := 0
	name:= "?"
	creator:= "?"
	for _, line := range strings.Split(string(decklist), "\n") {
		line = strings.TrimSpace(line)
		if ignoreDecklistLine(line) {
			
			// Sniff for special cases to handle them
			if strings.Contains(line, createrPrefix) {
				creator = handleCreatorLine(line)
			}

			if strings.Contains(line, namePrefix) {
				name = handleNameLine(line)
			}

			continue
		}

		card, err := NewCard(line)
		if err != nil {
			return nil, err
		}

		if count >= MainDeckSize {
			sideboard = append(sideboard, card)
		} else {
			maindeck = append(maindeck, card)
			count += int(card.Quantity)
		}
	}

	// Handle the funky encoding mtgtop8 uses...
	//
	// I feel dirty :(
	decoder1215:= charmap.Windows1252.NewDecoder()

	name, err = decoder1215.String(name)
	if err!=nil {
		return nil, err
	}
	creator, err = decoder1215.String(creator)
	if err!=nil {
		return nil, err
	}
	for _, c:= range maindeck {
		c.Name, err = decoder1215.String(c.Name)
		if err!=nil {
			return nil, err
		}
	}
	for _, c:= range sideboard {
		c.Name, err = decoder1215.String(c.Name)
		if err!=nil {
			return nil, err
		}
	}



	return &deckData.Deck{
		Name: name,
		Player: creator,
		Maindeck: maindeck,
		Sideboard: sideboard,
	}, nil
}