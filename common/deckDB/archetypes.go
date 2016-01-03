package deckDB

import(

	"./deckData"

	"./nameNorm"
	"github.com/jackc/pgx"

	"time"

	"fmt"

)

// Given an archetype name standardized by nameNorm,
// returns all cards that have appeared in that archetype.
func GetArchetypeContents(pool *pgx.ConnPool,
	name string) ([]*deckData.Card, error) {

	// Translate our clean names into mtgtop8 names
	// with associated metadata
	archetypes, presentCards, excludedCards, err:= nameNorm.Invert(name)
	if err!=nil {
		return nil, err
	}

	rows, err := pool.Query(archetypeContents, archetypes, 
		excludedCards, presentCards)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*deckData.Card
	for rows.Next() {
		c := deckData.Card{}

		err = rows.Scan(&c.Name, &c.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row", err)
		}

		cards = append(cards, &c)
	}

	return cards, nil

}

// Given an archetype name standardized by nameNorm,
// returns all cards that have appeared in that archetype.
func GetArchetypeLatest(pool *pgx.ConnPool,
	name string) (*deckData.TaggedDeck, error) {

	// Translate our clean names into mtgtop8 names
	// with associated metadata
	archetypes, presentCards, excludedCards, err:= nameNorm.Invert(name)
	if err!=nil {
		return nil, err
	}
	
	d:= deckData.Deck{Name: name}
	// We don't want the deckid exposed to the caller
	var Event string
	var t time.Time


	err = pool.QueryRow(archetypeLatest, archetypes, 
		excludedCards, presentCards).Scan(
			&d.DeckID,
			&d.Player, &Event, &t)
	if err != nil {
		return nil, err
	}

	maindeck, sideboard, err:= GetDeckContents(pool, d.DeckID)
	if err!=nil {
		return nil, err	
	}

	d.Maindeck = maindeck
	d.Sideboard = sideboard
	
	return &deckData.TaggedDeck{
		Event: Event,
		Happened: deckData.Timestamp(t),
		Deck: &d}, nil

}