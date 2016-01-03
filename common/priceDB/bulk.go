package priceDB

import (
	"github.com/jackc/pgx"

	"fmt"
)

func GetBulkLatestLowest(pool *pgx.ConnPool,
	names []string, source string) (Prices, error) {

	result:= make(Prices, len(names))

	failed:= 0

	for i, c:= range names {
		p, err:= GetCardLatestLowest(pool, c, source)
		if err!=nil {
			// Record this as a failure
			failed++

			// Insert a placeholder
			p = Price{
				Name: c,
				Set: "Unknown",
				Price: -1,
				Source: source,
			}
		}

		result[i] = p
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

	result:= make(Prices, len(names))

	failed:= 0

	for i, c:= range names {
		p, err:= GetCardLatestHighest(pool, c, source)
		if err!=nil {
			// Record this as a failure
			failed++

			// Insert a placeholder
			p = Price{
				Name: c,
				Set: "Unknown",
				Price: -1,
				Source: source,
			}
		}

		result[i] = p
	}

	// If we fail to fetch more than 1/3 of prices for this bulk
	// set, we error out
	if failed >= len(names) / 3 {
		return nil, fmt.Errorf("too many fetch failures")
	}

	return result, nil
}