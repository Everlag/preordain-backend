package deckDB

import(

	"./deckData"

	"github.com/jackc/pgx"

	"fmt"

)

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