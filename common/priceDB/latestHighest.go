package priceDB

import (
	"time"

	"github.com/jackc/pgx"
)

func GetCardLatestHighest(pool *pgx.ConnPool,
	name, source string) (Price, error) {

	if source == magiccardmarket {
		return getMKMLatestHighest(pool, name)
	} else if source == mtgprice {
		return getmtgpriceLatestHighest(pool, name)
	}

	return Price{}, SourceError

}

func getMKMLatestHighest(pool *pgx.ConnPool,
	name string) (Price, error) {

	var p Price
	var t time.Time
	err := pool.QueryRow(mkmPriceLatestHighest, name).Scan(
		&p.Name, &p.Set, &t, &p.Price, &p.Euro)
	if err != nil {
		return p, ScanError
	}

	p.Time = Timestamp(t)

	p.Source = magiccardmarket

	return p, nil

}

func getmtgpriceLatestHighest(pool *pgx.ConnPool,
	name string) (Price, error) {

	var p Price
	var t time.Time
	err := pool.QueryRow(mtgPriceLatestHighest, name).Scan(
		&p.Name, &p.Set, &t, &p.Price)
	if err != nil {
		return p, ScanError
	}

	p.Time = Timestamp(t)

	p.Source = mtgprice

	return p, nil

}
