package priceSources

import(

	"time"

	"./mtgprice"
	"./magiccardmarket"

	"fmt"
	"log"

)

const RateExceeded string = "Attempted update beyond allowed rate"

type PriceMap struct{

	// Where the prices come from
	Source string

	// If the prices are euro, thus, if conversion is required
	HasEuro bool

	// When we acquired the prices
	Time int64

	// Map from set to cards to prices as cents in USD
	Prices map[string]map[string]int64
	// Maps from set to cards to prices as cents in EUR if the HasEuro flag
	// is set
	EURPrices map[string]map[string]int64

}

func GetMKMPrices(aLogger *log.Logger) (PriceMap, error) {
	
	keys, setList, now, err:= getResources(PriceSourceRequested{MKM:true})
	if err!=nil {
		if err.Error() == RateExceeded {
			return PriceMap{},
			err
		}

		aLogger.Println("Failed to acquire resources,  magiccardmarket ", err)
		return PriceMap{},
		fmt.Errorf("Failed to acquire resources, magiccardmarket ", err) 
	}

	aLogger.Println("Acquiring price data for magiccardmarket")

	// Get the price data as euros
	MKMResults, err:= magiccardmarket.GetCardPrices(keys.MKMConsumerKey,
		keys.MKMSecretKey,
		setList)
	if err!=nil {
		aLogger.Println("Failed to acquire magiccardmarket prices, ", err)
		return PriceMap{},
		fmt.Errorf("Failed to acquire prices for magiccardmarket, logged")
	}
	
	aLogger.Println("Acquired price data for magiccardmarket")
	aLogger.Println("Converting currency price data for magiccardmarket")

	// Convert the price data from EUR to USD using the latest exchange rate
	MKMResultsAsUSD, err:= fromEURtoUSD(keys.OpenexchangeratesKey,
		MKMResults)
	if err!=nil {
		aLogger.Println("Failed to convert magiccardmarket prices, ", err)
		return PriceMap{},
		fmt.Errorf("Failed to convert prices for magiccardmarket, logged")
	}

	aLogger.Println("Converted currency price data for magiccardmarket")

	pricingResults:= PriceMap{
		Source: magiccardmarket.PriceSource,
		Time: now,
		Prices: MKMResultsAsUSD,
		EURPrices: MKMResults,
		HasEuro: true, // Signal currency conversion is required.
	}

	// Store the new update time
	keys.MKMLastUpdate = now
	err = keys.updateOnDisk()
	if err!=nil {
		aLogger.Println("Failed to update on disk timestamp for source mtgprice, ", 
			err)
	}


	return pricingResults, nil


}

// Reads credentials and set list from disk then acquires a round of mtgprice
//
// Returns prices as an array of maps from sets to cards which map to prices
//
// Logs all events of failure to acquire prices
func GetmtgpricePrices(aLogger *log.Logger) (PriceMap, error) {
	keys, setList, now, err:= getResources(PriceSourceRequested{Mtgprice:true})
	if err!=nil {
		if err.Error() == RateExceeded {
			return PriceMap{},
			err
		}
		aLogger.Println("Failed to acquire resources, mtgprice ", err)
		return PriceMap{},
		fmt.Errorf("Failed to acquire resources, mtgprice ", err) 
	}

	aLogger.Println("Acquiring price data for mtgprice")

	mtgpriceResults, err:= mtgprice.GetCardPrices(keys.Mtgprice, setList, aLogger)
	if err!=nil {
		aLogger.Println("Failed to acquire mtgprice prices, ", err)
		return PriceMap{},
		fmt.Errorf("Failed to acquire prices for mtgprice, logged")
	}

	aLogger.Println("Acquired price data for mtgprice")

	pricingResults:= PriceMap{
		Source: mtgprice.PriceSource,
		Time: now,
		Prices: mtgpriceResults,
		HasEuro: false,
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

type PriceSourceRequested struct{

	MKM, Mtgprice bool

}

// Acquires all necessary generic resources for a price run.
//
// Provide it a bool for the specific source you wish to use
//
// Returns apikeys, set list, and the current time.
//
// Performs basic rate limiting sanity test.
func getResources(requestedSource PriceSourceRequested) (ApiKeys,
	[]string, int64, error) {

	// Sanity check to ensure user only wants one source for the keys
	if requestedSource.MKM && requestedSource.Mtgprice ||
	!(requestedSource.MKM || requestedSource.Mtgprice) {
		return ApiKeys{}, nil, 0, 
		fmt.Errorf("Choose one price source for keyset")
	}
	
	keys, err:= getApiKeys()
	if err!=nil {
		return ApiKeys{}, nil, 0, 
		fmt.Errorf("Failed to acquire apiKeys, ", err)
	}

	// Ensure we aren't sending too many price requests.
	now:= time.Now().UTC().Unix()
	if requestedSource.MKM {
		if now < keys.MKMPriceWaitTime + keys.MKMLastUpdate {
			return ApiKeys{}, nil, 0, fmt.Errorf(RateExceeded)
		}
	}else if requestedSource.Mtgprice {
		if now < keys.MtgpriceWaitTime + keys.MtgpriceLastUpdate {
			return ApiKeys{}, nil, 0, fmt.Errorf(RateExceeded)
		}
	}

	setList, err:= getSetList()
	if err!=nil {
		return ApiKeys{}, nil, 0,
		fmt.Errorf("Failed to acquire setList, ", err)
	}

	return keys, setList, now, nil

}
