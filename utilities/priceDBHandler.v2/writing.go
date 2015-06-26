package priceDB

import (
	"fmt"
	"time"

	"github.com/jackc/pgx"
)

const mtgpriceSource string = "mtgprice"
const mkmpriceSource string = "magiccardmarket"

var ConnError error = fmt.Errorf("db connection failed")
var SourceError error = fmt.Errorf("invalid price source")
var ScanError error = fmt.Errorf("failed to scan row")

type Prices []Price

type Price struct {
	Name, Set string
	Time      Timestamp

	Euro  int32 `json:",omitempty"`
	Price int32

	Source string
}

func SendPrices(pool *pgx.ConnPool, prices Prices) error {

	tx, err := pool.Begin()
	if err != nil {
		return ConnError
	}

	// We can exit anytime before tx.Commit is called
	// and avoid any changes to the db
	defer tx.Rollback()

	for _, p := range prices {

		err = insertPrice(tx, p)
		if err != nil {
			return err
		}

	}

	tx.Commit()

	return nil

}

// Inserts a price point into the db using a passed transaction
//
// Relies on the p.Source to determine which price table to insert into
func insertPrice(tx *pgx.Tx, p Price) (err error) {

	var statement string

	if p.Source == mtgpriceSource {
		statement = mtgpriceInsert
	} else if p.Source == mkmpriceSource {
		statement = mkmpriceInsert
	}

	if statement == "" {
		return SourceError
	}

	// Ignore the euro unless the price originates from a euro based source.
	if p.Source == mkmpriceSource {
		_, err = tx.Exec(statement, p.Name, p.Set,
			time.Time(p.Time), p.Euro, p.Price)
	} else if p.Source == mtgprice {
		_, err = tx.Exec(statement, p.Name, p.Set, time.Time(p.Time), p.Price)
	} else {
		return SourceError
	}

	return err

}
