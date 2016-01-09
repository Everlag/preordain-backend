package priceDB

import (
	"time"

	"github.com/jackc/pgx"
)

func GetCardLatestLowest(pool *pgx.ConnPool,
	name, source string) (Price, error) {

	if source == magiccardmarket {
		return getMKMLatestLowest(pool, name)
	} else if source == mtgprice {
		return getmtgpriceLatestLowest(pool, name)
	}

	return Price{}, SourceError

}

func getMKMLatestLowest(pool *pgx.ConnPool,
	name string) (Price, error) {

	var p Price
	var t time.Time
	err := pool.QueryRow(mkmPriceLatestLowest, name).Scan(
		&p.Name, &p.Set, &t, &p.Price)
	if err != nil {
		return p, ScanError
	}

	p.Time = Timestamp(t)

	p.Source = magiccardmarket

	return p, nil

}

func getmtgpriceLatestLowest(pool *pgx.ConnPool,
	name string) (Price, error) {

	var p Price
	var t time.Time
	err := pool.QueryRow(mtgPriceLatestLowest, name).Scan(
		&p.Name, &p.Set, &t, &p.Price)
	if err != nil {
		return p, ScanError
	}

	p.Time = Timestamp(t)

	p.Source = mtgprice

	return p, nil

}
