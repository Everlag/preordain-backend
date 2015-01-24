package magiccardmarket
// Provides access to magiccardmarket.eu pricing data.
// Note: all returned prices are in EURO.

import(

	"fmt"

	"net/http"

)

const PriceSource string = "magiccardmarket"

// Note: MKM is not completely rfc compliant because %encoded signature
// is a no-go. Thus, cannot use standard oauth clients.
const expansionPath string = 
"https://www.mkmapi.eu/ws/v1.1/output.json/expansion/1"
const productPath string = 
"https://www.mkmapi.eu/ws/v1.1/output.json/product/"

// The specific language to request for a product
const englishID string = "1"

const oauthVersion string = "1.0"
const oauthSignatureMethod string = "HMAC-SHA1"

// How many concurrent workers we have.
// Due to the fact that a 150ms latency alone will result in 1.75 hours of
// fetching, concurrency is absolutely required if this is to finish
// in a sane amount of time.
var WorkerCount int = 20


// Returns a map from set to card to price in EURO.
//
// Handles foil sets.
//
// Architected as follows:
// Acquires a map of clean set names to MKM compatiable.
// Runs as many as WorkerCount price workers that each scrape sets.
func GetCardPrices(consumerKey, consumerSecret string,
	setList []string) (map[string]map[string]int64, error) {
	
	// Acquire an http client to use for the duration.
	aClient:= &http.Client{}

	// Build a map from our sets to MKM sets
	cleanToMKM, err:= GetSetMap(consumerKey, consumerSecret,
		setList, aClient)
	if err!=nil {
		return map[string]map[string]int64{},
		fmt.Errorf("Failed to get set map, ", err)
	}

	priceMap:= make(map[string]map[string]int64)


	err = runWorkers(setList, priceMap,
		cleanToMKM,
		consumerKey, consumerSecret)
	if err!=nil {
		return map[string]map[string]int64{},
		fmt.Errorf("Price worker encountered ", err)
	}

	return priceMap, nil


}