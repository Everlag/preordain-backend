package priceDB

import (
	"time"
	
	"github.com/jackc/pgx"
)

func GetCardMedianHistory(pool *pgx.ConnPool,
	name, set, source string) (Prices, error) {

	if source == magiccardmarket {
		return getMKMMedianHistory(pool, name, set)
	} else if source == mtgprice {
		return getmtgpriceMedianHistory(pool, name, set)
	}

	return nil, SourceError

}

func getMKMMedianHistory(pool *pgx.ConnPool,
	name, set string) (Prices, error) {

	rows, err := pool.Query(mtgpriceMedian, name, set)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices Prices
	for rows.Next() {
		p := Price{
			Name: name,
			Set: set,
			Source: magiccardmarket,
		}

		var t time.Time
		err = rows.Scan(&t, &p.Price)
		if err != nil {
			return nil, ScanError
		}

		p.Time = Timestamp(t)

		prices = append(prices, p)
	}

	return prices, nil

}

func getmtgpriceMedianHistory(pool *pgx.ConnPool,
	name, set string) (Prices, error) {

	rows, err := pool.Query(mtgpriceMedian, name, set)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices Prices
	for rows.Next() {
		p := Price{
			Name: name,
			Set: set,
			Source: mtgprice,
		}

		var t time.Time
		err = rows.Scan(&t, &p.Price)
		if err != nil {
			return nil, ScanError
		}

		p.Time = Timestamp(t)

		prices = append(prices, p)
	}

	return prices, nil

}
