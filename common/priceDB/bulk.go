package priceDB

import (
	"github.com/jackc/pgx"
)

func GetBulkLatestLowest(pool *pgx.ConnPool,
	names []string, source string) (Prices, error) {

	result:= make(Prices, len(names))

	for i, c:= range names {
		p, err:= GetCardLatestLowest(pool, c, source)
		if err!=nil {
			return nil, err
		}

		result[i] = p
	}

	return result, nil
}

func GetBulkLatestHighest(pool *pgx.ConnPool,
	names []string, source string) (Prices, error) {

	result:= make(Prices, len(names))

	for i, c:= range names {
		p, err:= GetCardLatestHighest(pool, c, source)
		if err!=nil {
			return nil, err
		}

		result[i] = p
	}

	return result, nil
}