package priceSources

import(

	"fmt"

	"./currencyConversion"

)


func fromEURtoUSD(conversionKey string,
	EURPrices map[string]map[string]int64) (map[string]map[string]int64,
		error) {

	// Acquire the conversion rate
	ratio, err:= currencyConversion.EURToUSD(conversionKey)
	if err!=nil {
		return map[string]map[string]int64{},
		fmt.Errorf("Failed to acquire conversion ratio, ", err)
	}

 	// Prepare a container for the results of the conversion
	resultsMap:= make(map[string]map[string]int64)

	// Do the full conversion, shouldn't take too long
	for aSetName, aSet:= range EURPrices{

		resultsMap[aSetName] = make(map[string]int64)

		for aCardName, aCardValue:= range aSet{

			asUSD:= int64(ratio * float64(aCardValue))
			resultsMap[aSetName][aCardName] = asUSD

		}

	}

	return resultsMap, nil

}