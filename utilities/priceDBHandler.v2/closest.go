package priceDB

import (
	"time"

	"github.com/jackc/pgx"
)

func GetCardClosest(pool *pgx.ConnPool,
	name, set string, when Timestamp, source string) (Price, error) {

	if source == magiccardmarket {
		return getMKMClosest(pool, name, set, time.Time(when))
	} else if source == mtgprice {
		return getmtgpriceClosest(pool, name, set, time.Time(when))
	}

	return Price{}, SourceError

}

func getMKMClosest(pool *pgx.ConnPool,
	name, set string, when time.Time,) (Price, error) {

	var p Price
	var t time.Time
	err := pool.QueryRow(mkmPriceClosest, name, set, when).Scan(
		&p.Name, &p.Set, &t, &p.Price, &p.Euro)
	if err != nil {
		return p, ScanError
	}

	p.Time = Timestamp(t)

	p.Source = magiccardmarket

	return p, nil

}

func getmtgpriceClosest(pool *pgx.ConnPool,
	name, set string, when time.Time) (Price, error) {

	var p Price
	var t time.Time
	err := pool.QueryRow(mtgPriceClosest, name, set, when).Scan(
		&p.Name, &p.Set, &t, &p.Price)
	if err != nil {
		return p, ScanError
	}

	p.Time = Timestamp(t)

	p.Source = mtgprice

	return p, nil

}
