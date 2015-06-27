package priceDB

import (
	"time"

	"github.com/jackc/pgx"

	"fmt"
)

func GetCardLatest(pool *pgx.ConnPool,
	name, set, source string) (Price, error) {

	if source == magiccardmarket {
		return getMKMLatest(pool, name, set)
	} else if source == mtgprice {
		return getmtgpriceLatest(pool, name, set)
	}

	return Price{}, SourceError

}

func getMKMLatest(pool *pgx.ConnPool,
	name, set string) (Price, error) {

	var p Price
	var t time.Time
	err := pool.QueryRow(mkmPriceLatest, name, set).Scan(
		&p.Name, &p.Set, &t, &p.Price, &p.Euro)
	if err != nil {
		return p, ScanError
	}

	p.Time = Timestamp(t)

	p.Source = magiccardmarket

	return p, nil

}

func getmtgpriceLatest(pool *pgx.ConnPool,
	name, set string) (Price, error) {

	var p Price
	var t time.Time
	err := pool.QueryRow(mtgPriceLatest, name, set).Scan(
		&p.Name, &p.Set, &t, &p.Price)
	if err != nil {
		return p, ScanError
	}

	p.Time = Timestamp(t)

	fmt.Println(t, p.Time)

	p.Source = mtgprice

	return p, nil

}
