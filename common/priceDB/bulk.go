package priceDB

import (
	"github.com/jackc/pgx"

	"fmt"

	"time"
)

func GetBulkLatestLowest(pool *pgx.ConnPool,
	names []string, source string) (Prices, error) {

	query:= mtgPriceLatestLowest
	if source == magiccardmarket {
		query = mkmPriceLatestLowest
	}

	// We use a stored function to handle this to avoid
	// significantly more code duplication.
	rows, err := pool.Query(bulkLatest, query, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	result:= make(Prices, 0)

	failed:= 0

	for rows.Next() {
		p := Price{}

		var t time.Time
		err = rows.Scan(&p.Name, &p.Set, &t, &p.Price)
		if err != nil {
			failed++
			// Indicate a failed price,
			// we can't possibly know the name of the card that
			// failed so the client can figure that out!
			p = Price{
				Name: "Unknown",
				Set: "Unknown",
				Price: -1,
			}
		}

		p.Time = Timestamp(t)
		p.Source = source

		result = append(result, p)
	}

	// If we fail to fetch more than 1/3 of prices for this bulk
	// set, we error out
	if failed >= len(names) / 3 {
		return nil, fmt.Errorf("too many fetch failures")
	}

	return result, nil
}

func GetBulkLatestHighest(pool *pgx.ConnPool,
	names []string, source string) (Prices, error) {

	query:= mtgPriceLatestHighest
	if source == magiccardmarket {
		query = mkmPriceLatestHighest
	}

	// We use a stored function to handle this to avoid
	// significantly more code duplication.
	rows, err := pool.Query(bulkLatest, query, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	result:= make(Prices, 0)

	failed:= 0

	for rows.Next() {
		p := Price{}

		var t time.Time
		err = rows.Scan(&p.Name, &p.Set, &t, &p.Price)
		if err != nil {
			failed++
			// Indicate a failed price,
			// we can't possibly know the name of the card that
			// failed so the client can figure that out!
			p = Price{
				Name: "Unknown",
				Set: "Unknown",
				Price: -1,
			}
		}

		p.Time = Timestamp(t)
		p.Source = source

		result = append(result, p)
	}

	// If we fail to fetch more than 1/3 of prices for this bulk
	// set, we error out
	if failed >= len(names) / 3 {
		return nil, fmt.Errorf("too many fetch failures")
	}

	return result, nil
}