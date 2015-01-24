package mtgprice
// Handles acquiring price data from mtgprice.

import(

	"log"
	"fmt"

	"net/http"
	"io/ioutil"

	"encoding/json"

	"time"

	"strings"
	"strconv"

)

// Which price source this is. It may seem redundant but this is important
// metadata
const PriceSource string = "mtgprice"

const baseurl string = "http://www.mtgprice.com/api?"

const badSetName string = "ERROR: set looks incorrect"
const BlankSetName string = "Blank Set Name"

// Returns a map of set to card to price for mtgprice.com
// using the provided setList and the apikey. A non-nil error is returned
// in the event we ran into an issue. Logging is done to a provided logger
func GetCardPrices(apiKey string, sets []string,
	priceLogger *log.Logger)(map[string]map[string]int64, error){

	setMap:= make(map[string]map[string]int64)

	var err error
	var persistentErr error
	for _, aSet:= range sets{

		setMap[aSet], err = getSetData(apiKey, aSet, priceLogger)
		if err!=nil && err.Error()!=BlankSetName {
			// Return a serious error only as we will otherwise end up
			// discarding a significant amount of price data on a malformed
			// set name.
			persistentErr = err
		}

	}

	return setMap, persistentErr

}

// Provided an apikey and a set name, this fetches all prices
// for the set and returns that in a map[cardName]cardPrice
func getSetData(apiKey, set string,
	priceLogger *log.Logger) (map[string]int64, error) {
	
	// normalize the set name to be mtgprice compatiable
	cleanedName := cleanSetName(set)
	if cleanedName == "" {
		return nil, fmt.Errorf(BlankSetName)
	}

	// derive the url we will be querying
	finalUrl := baseurl + "apiKey=" + apiKey + "&s=" + cleanedName

	priceData, err:= getRawSetData(finalUrl)
	if err!=nil {
		priceLogger.Println("mtgprice - ", err,
			"Set name : Cleaned Name",set," : ",cleanedName)
		return nil, err
	}

	priceMap:= make(map[string]int64)

	for _, aCard:= range priceData.Cards{

		priceMap[aCard.Name] = aCard.Price

	}


	return priceMap, nil

}

// Acquires set data for provided url and cleans to to our expectations
func getRawSetData(url string) (SetData, error){
	

	retrievedPriceData, err:= retrieveRawSetData(url)
	if err!=nil {
		return retrievedPriceData, err
	}

	// Note the timestamp for when this set is acquired
	retrievedPriceData.TimeRetrieved = time.Now().UTC().Unix()

	// Go through and convert from
	// mtgprice price shorthand(eg 1.8 instead of 1.80)

	retrievedPriceData.toStandardPrices()

	return retrievedPriceData, nil

}

// GETs and unmarshals price data for provided url
func retrieveRawSetData(url string) (SetData, error) {
	
	resp, err := http.Get(url)
	if err != nil {
		
		return SetData{}, fmt.Errorf("Failed to retrieve set data")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SetData{}, fmt.Errorf("Failed to read retrieved set data, ", err)
	}


	if string(body) == badSetName {
		return SetData{}, fmt.Errorf("Set parsing failed, ", err)
	}

	// The api needs some cards to be normalized before they can be accepted
	body = handleSpecialCardCases(body)

	var retrievedPriceData SetData
	err = json.Unmarshal(body, &retrievedPriceData)
	if err != nil {
		return SetData{}, fmt.Errorf("Failed to unmarshal data received")
	}

	return retrievedPriceData, nil

}

// Set data for a specific instance in time
type SetData struct {
	Name string

	Cards []cardData

	// Unix timestamp of set retrieval
	TimeRetrieved int64
}

// An extended form of what mtgprice provides us
type cardData struct {
	//data as returned by mtgprice
	MtgpriceID string
	Name       string
	FairPrice  string

	// Price as a number of cents
	Price int64
}

// Converts from mtgprice shorthand to int64 notation
func (aSet *SetData) toStandardPrices() {
	
	// We have a temporary store of cards as some of them may be culled
	// during the price parsing
	tempCards:= make([]cardData, 0)

	for _, aCard:= range aSet.Cards{

		// Remove the dollar sign
		aCard.FairPrice = strings.Replace(aCard.FairPrice, "$", "", -1)

		// Convert to dollars as a float64, not extremely clean but
		// prices are only provided as two decimals of accuracy
		dollars, err:= strconv.ParseFloat(aCard.FairPrice, 64)
		if err!=nil {
			continue
		}

		floatingCents:= dollars * 100
		cents:= int64(floatingCents)
		aCard.Price = cents

		tempCards = append(tempCards, aCard)

	}

	aSet.Cards = tempCards

}