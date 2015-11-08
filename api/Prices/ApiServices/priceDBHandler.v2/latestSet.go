package priceDB

import (
	"time"

	"github.com/jackc/pgx"
)

func GetSetLatest(pool *pgx.ConnPool, set, source string) (Prices, error) {

	if source == magiccardmarket {
		return getMKMSetLatest(pool, set)
	} else if source == mtgprice {
		return getmtgpriceSetLatest(pool, set)
	}

	return nil, SourceError

}

func getMKMSetLatest(pool *pgx.ConnPool, set string) (Prices, error) {

	rows, err := pool.Query(mkmPriceSetLatest, set)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices Prices
	for rows.Next() {
		p := Price{}

		var t time.Time
		err = rows.Scan(&p.Name, &p.Set, &t, &p.Price, &p.Euro)
		if err != nil {
			return nil, ScanError
		}

		p.Source = magiccardmarket
		p.Time = Timestamp(t)

		prices = append(prices, p)
	}

	return prices, nil

}

func getmtgpriceSetLatest(pool *pgx.ConnPool, set string) (Prices, error) {

	rows, err := pool.Query(mtgPriceSetLatest, set)
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

		p.Source = mtgprice
		p.Time = Timestamp(t)

		prices = append(prices, p)
	}

	return prices, nil

}
