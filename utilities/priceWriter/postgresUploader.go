package main

import(
	"time"
	
	"./priceSources"

	"github.com/Everlag/preordain-backend/api/Prices/ApiServices/priceDBHandler.v2"
	"github.com/jackc/pgx"

)

func uploadSingleSourceResults(aPriceResult priceSources.PriceMap,
	pool *pgx.ConnPool) error {

	prices:= make(priceDB.Prices, 0)

	var p priceDB.Price
	var t time.Time
	for aSetName, cardMap:= range aPriceResult.Prices{

		for aCardName, aPrice:= range cardMap{

			var euroPrice int64
			if aPriceResult.HasEuro {
				euroPrice = aPriceResult.EURPrices[aSetName][aCardName]
			}

			t = time.Unix(aPriceResult.Time, 0)

			p = priceDB.Price{
					Name: aCardName,
					Set: aSetName,
					Time: priceDB.Timestamp(t),
					Price: int32(aPrice),
					Euro: int32(euroPrice),
					Source: aPriceResult.Source,
				}

			prices = append(prices, p)

		}

	}

	return priceDB.SendPrices(pool, prices)

}