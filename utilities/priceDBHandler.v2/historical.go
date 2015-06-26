package priceDB

import (
	"time"

	"github.com/jackc/pgx"
)

func GetCardHistory(pool *pgx.ConnPool,
	name, set, source string) (Prices, error) {

	if source == magiccardmarket {
		return getMKMHistory(pool, name, set)
	} else if source == mtgprice {
		return getmtgpriceHistory(pool, name, set)
	}

	return nil, SourceError

}

func getMKMHistory(pool *pgx.ConnPool,
	name, set string) (Prices, error) {

	rows, err := pool.Query(mtgpriceHistory, name, set)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices Prices
	for rows.Next() {
		p := Price{}

		var t time.Time
		err = rows.Scan(&p.Name, &p.Set, &t, &p.Price)
		if err != nil {
			return nil, ScanError
		}

		p.Time = Timestamp(t)

		prices = append(prices, p)
	}

	return prices, nil

}

func getmtgpriceHistory(pool *pgx.ConnPool,
	name, set string) (Prices, error) {

	rows, err := pool.Query(mkmpriceHistory, name, set)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices Prices
	for rows.Next() {
		p := Price{}

		var t time.Time
		err = rows.Scan(&p.Name, &p.Set, &t, &p.Price, &p.Source)
		if err != nil {
			return nil, ScanError
		}

		p.Time = Timestamp(t)

		prices = append(prices, p)
	}

	return prices, nil

}
