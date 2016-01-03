package deckDB

import(

	"./deckData"
	"./nameNorm"

	"github.com/jackc/pgx"

	"fmt"

)

// Fetch and return a deck
func GetDeck(pool *pgx.ConnPool,
	deckid string) (*deckData.Deck, error) {

	// Fetch contents
	maindeck, sideboard, err:= GetDeckContents(pool, deckid)
	if err!=nil {
		return nil, err
	}

	player, archetype, err:= GetDeckMeta(pool, deckid)
	if err!=nil {
		return nil, err
	}

	d:= &deckData.Deck{
		Name: archetype,
		Player: player,
		DeckID: deckid,
		Maindeck: maindeck,
		Sideboard: sideboard,
	}

	// Turn the name into something we're comfortable with
	err = nameNorm.Clean(d)

	return d, err
}

// Acquire the player and mtgtop8 name for the deck
func GetDeckMeta(pool *pgx.ConnPool,
	deckid string) (player, archetype string, err error) {

	err = pool.QueryRow(deckMeta, deckid).Scan(&player, &archetype)

	return

}

// Fetch the contents of a deck denoted by a deckid
func GetDeckContents(pool *pgx.ConnPool,
	deckid string) (maindeck, sideboard []*deckData.Card, err error) {

	rows, err := pool.Query(deckContents, deckid)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var cards []*deckData.Card
	var sideboardStatus []bool
	for rows.Next() {
		c := deckData.Card{}

		sideboard:= false

		err = rows.Scan(&c.Name, &c.Quantity, &sideboard)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan row", err)
		}

		cards = append(cards, &c)
		sideboardStatus = append(sideboardStatus, sideboard)
	}

	// Decks may not contain exactly 75 cards, we
	// can handle the extra allocations
	for i, c:= range cards{
		if sideboardStatus[i] {
			sideboard = append(sideboard, c)
			continue
		}

		maindeck = append(maindeck, c)
	}

	return
}