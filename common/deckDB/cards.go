package deckDB

import(

	"./nameNorm"
	"github.com/jackc/pgx"

	"fmt"

)

// Given an card name, returns all archetypes it has appeared within.
func GetCardArchetypes(pool *pgx.ConnPool,
	card string) ([]string, error) {

	
	rows, err := pool.Query(cardArchetypes, card)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Unclean archetype names fresh off mtgtop8
	dirty:= make([]string, 0)
	for rows.Next() {
		a:= ""

		err = rows.Scan(&a)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row", err)
		}

		dirty = append(dirty, a)
	}

	clean:= make([]string, 0)
	for _, d:= range dirty{
		n, err:= nameNorm.CleanEst(d)
		if err !=nil {
			continue
		}

		clean = append(clean, n)
	}

	// Deduplicate
	deduper:= make(map[string]struct{})
	for _, c:= range clean{
		deduper[c] = struct{}{}
	}
	archetypes:= make([]string, 0)
	for k, _:= range deduper{
		archetypes = append(archetypes, k)
	}

	if len(archetypes) == 0 {
		err = fmt.Errorf("no archetypes found")
	}

	return archetypes, err

}
