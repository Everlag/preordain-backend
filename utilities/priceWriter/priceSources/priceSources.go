package priceSources

import(

	"time"

	"./mtgprice"

	"fmt"
	"log"

)

const RateExceeded string = "Attempted update beyond allowed rate"

type PriceMap struct{

	Source string

	Time int64

	Prices map[string]map[string]int64

}

// Reads credentials and set list from disk then acquires a round of mtgprice
//
// Returns prices as an array of maps from sets to cards which map to prices
//
// Logs all events of failure to acquire prices
func GetmtgpricePrices(aLogger *log.Logger) (PriceMap, error) {
	keys, err:= getApiKeys()
	if err!=nil {
		aLogger.Println("Failed to acquire apiKeys, ", err)
		return PriceMap{}, err
	}

	// Make a check to ensure we aren't sending unnecessary traffic
	now:= time.Now().UTC().Unix()
	if now < keys.MtgpriceWaitTime + keys.MtgpriceLastUpdate {
		return PriceMap{}, fmt.Errorf(RateExceeded)
	}

	setList, err:= getSetList()
	if err!=nil {
		aLogger.Println("Failed to acquire setList, ", err)
		return PriceMap{}, err
	}

	aLogger.Println("Acquiring price data for mtgprice")

	mtgpriceResults, err:= mtgprice.GetCardPrices(keys.Mtgprice, setList, aLogger)
	if err!=nil {
		aLogger.Println("Failed to acquire mtgprice prices, ", err)
		return PriceMap{}, fmt.Errorf("Failed to acquire prices for mtgprice, logged")
	}

	aLogger.Println("Acquired price data for mtgprice")

	pricingResults:= PriceMap{
		Source: mtgprice.PriceSource,
		Time: now,
		Prices: mtgpriceResults,
	}

	// Store the new update time
	keys.MtgpriceLastUpdate = now
	err = keys.updateOnDisk()
	if err!=nil {
		aLogger.Println("Failed to update on disk timestamp for source mtgprice, ", 
			err)
	}


	return pricingResults, nil

}