package priceDB

import (
	"github.com/jackc/pgx"

	"fmt"

	"time"
)

func GetBulkLatestLowest(pool *pgx.ConnPool,
	names []string, source string) (Prices, error) {

	query:= mtgPriceLatestLowest
	if source == magiccardmarket {
		query = mkmPriceLatestLowest
	}

	// We use a stored function to handle this to avoid
	// significantly more code duplication.
	rows, err := pool.Query(bulkLatest, query, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	result:= make(Prices, 0)

	failed:= 0

	for rows.Next() {
		p := Price{}

		var t time.Time
		err = rows.Scan(&p.Name, &p.Set, &t, &p.Price)
		if err != nil {
			failed++
			// Indicate a failed price,
			// we can't possibly know the name of the card that
			// failed so the client can figure that out!
			p = Price{
				Name: "Unknown",
				Set: "Unknown",
				Price: -1,
			}
		}

		p.Time = Timestamp(t)
		p.Source = source

		result = append(result, p)
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

	query:= mtgPriceLatestHighest
	if source == magiccardmarket {
		query = mkmPriceLatestHighest
	}

	// We use a stored function to handle this to avoid
	// significantly more code duplication.
	rows, err := pool.Query(bulkLatest, query, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	result:= make(Prices, 0)

	failed:= 0

	for rows.Next() {
		p := Price{}

		var t time.Time
		err = rows.Scan(&p.Name, &p.Set, &t, &p.Price)
		if err != nil {
			failed++
			// Indicate a failed price,
			// we can't possibly know the name of the card that
			// failed so the client can figure that out!
			p = Price{
				Name: "Unknown",
				Set: "Unknown",
				Price: -1,
			}
		}

		p.Time = Timestamp(t)
		p.Source = source

		result = append(result, p)
	}

	// If we fail to fetch more than 1/3 of prices for this bulk
	// set, we error out
	if failed >= len(names) / 3 {
		return nil, fmt.Errorf("too many fetch failures")
	}

	return result, nil
}

type SummedWeeks []SummedWeek

// The sum of a number of cards' weekly medians
type SummedWeek struct {
	Time      Timestamp

	Price int32

	Source string
}

func GetBulkWeeklyLowest(pool *pgx.ConnPool,
	names []string, multipliers []int32,
	source string) (SummedWeeks, error) {

	if len(names) != len(multipliers) {
		return nil, fmt.Errorf("failed to match a multiplier to each card")
	}

	query:= mtgPriceWeeksLow
	if source == magiccardmarket {
		query = mkmPriceWeeksLow
	}

	// We use a stored function to handle this to avoid
	// significantly more code duplication.
	rows, err := pool.Query(bulkExtrema, query, names, multipliers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	result:= make(SummedWeeks, 0)

	for rows.Next() {
		w := SummedWeek{}

		var t time.Time
		err = rows.Scan(&t, &w.Price)
		if err != nil {
			return nil, ScanError
		}

		w.Time = Timestamp(t)
		w.Source = source

		result = append(result, w)
	}

	return result, nil
}

func GetBulkWeeklyHighest(pool *pgx.ConnPool,
	names []string, multipliers []int32,
	source string) (SummedWeeks, error) {

	if len(names) != len(multipliers) {
		return nil, fmt.Errorf("failed to match a multiplier to each card")
	}

	query:= mtgPriceWeeksHigh
	if source == magiccardmarket {
		query = mkmPriceWeeksHigh
	}

	// We use a stored function to handle this to avoid
	// significantly more code duplication.
	rows, err := pool.Query(bulkExtrema, query, names, multipliers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	result:= make(SummedWeeks, 0)

	for rows.Next() {
		w := SummedWeek{}

		var t time.Time
		err = rows.Scan(&t, &w.Price)
		if err != nil {
			return nil, ScanError
		}

		w.Time = Timestamp(t)
		w.Source = source

		result = append(result, w)
	}

	return result, nil
}