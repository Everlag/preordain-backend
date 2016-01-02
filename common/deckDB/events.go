package deckDB

import(

	"github.com/jackc/pgx"

)

// Fetch the contents of a deck denoted by a deckid
func HasEvent(pool *pgx.ConnPool,
	eventid string) (bool, error) {

	var present bool

	err:= pool.QueryRow(eventExists, eventid).Scan(&present)
	if err != nil {
		return false, err
	}

	return present, nil
}