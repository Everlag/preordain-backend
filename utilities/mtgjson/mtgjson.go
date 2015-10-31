package mtgjson

import (
	"fmt"

	"encoding/json"
	"io/ioutil"

	"strings"
)

const AllSetsXLoc string = "AllSets-x.json"
const AllSetsLoc string = "AllSets.json"
const AllCardsXLoc string = "AllCards-x.json"
const AllCardsLoc string = "AllCards.json"

type Cards map[string]*Card

// The full card mtgjson exposes with additions
type Card struct {
	Name                                   string
	Text                                   string
	ManaCost                               string
	Colors                                 []string
	Power, Toughness, Type, ImageName      string
	Printings, Types, SuperTypes, SubTypes []string
	Legalities                             []struct {
		Format   string
		Legality string
	}
	Rarity   string
	Reserved bool
	Loyalty  int

	// Extra flag incase it must be removed before being
	// passed to any clients
	invalid bool
}

// Fetches and unmarshals AllCards.json
func AllCards() (Cards, error) {
	return genericCards(AllCardsLoc)
}

// Fetches and unmarshals AllCards-X.json
func AllCardsX() (Cards, error) {
	return genericCards(AllCardsXLoc)
}

// Attempts to fetch and unmarhsal name into some Cards
func genericCards(name string) (Cards, error) {
	raw, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read", name, err)
	}

	var cs Cards
	err = json.Unmarshal(raw, &cs)
	if err != nil {
		return nil, fmt.Errorf("to unmarshal", name, err)
	}

	// Clean up the data for our specific uses
	for _, c:= range cs{
		c.clean()
		if c.invalid {
			delete(cs, c.Name)
		}
	}

	return cs, nil
}

type Sets map[string]*Set

// The full set metadata mtgjson exposes
type Set struct {
	Name        string
	ReleaseDate string
	Type        string
	Code        string
	Booster     interface{}

	Cards       []Card

	Timestamp int64
}

// Fetches and unmarshals AllSets.json
func AllSets() (Sets, error) {
	return genericSets(AllSetsLoc)
}

// Fetches and unmarshals AllSets-X.json
func AllSetsX() (Sets, error) {
	return genericSets(AllSetsXLoc)
}

// Attempts to fetch and unmarshal name into some Sets
func genericSets(name string) (Sets, error) {
	raw, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read", name, err)
	}

	var s Sets
	err = json.Unmarshal(raw, &s)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal", name, err)
	}

	// Clean the sets and their contents for our specific use case
	for _, s:= range s{
		s.clean()
	}

	return s, nil
}

// Deal with special cases regarding cards
func (c *Card) clean() {
	for i, s:= range c.Printings{
		// Wizards had a bad inital version on gatherer
		if s == "Zendikar Expeditions" {
			c.Printings[i] = "Zendikar Expedition"
		}
	}

	// Flag the avatar cards for removal
	if strings.Contains(c.Name, "Avatar") {
		c.invalid = true
	}
}

// Deal with special caes regarding cards
func (s *Set) clean() {
	if s.Name == "Zendikar Expeditions" {
		s.Name = "Zendikar Expedition"
	}

	// Remove invalid cards
	cleaned:= make([]Card, 0)
	for _, c:= range s.Cards{
		c.clean()

		if !c.invalid {
			cleaned = append(cleaned, c)
		}
	}
	s.Cards = cleaned
}