package magiccardmarket

import(

	"fmt"

	"net/url"
	"strings"
	"strconv"

	"net/http"
	"encoding/json"

)

// Acquires the prices for an expansion and returns a map[cardName]EUR-Cents
func getExpansionPrices(aSet string, foil bool,
	cleanToMKM map[string]string,
	consumerKey, consumerSecret string,
	aClient *http.Client) (map[string]int64, error) {
	
	// Ensure we have a map to stick these prices in
	priceHolster:= make(map[string]int64)

	// Acquire set ids
	setIds, err:= GetExpansion(consumerKey, consumerSecret, aSet,
		cleanToMKM, aClient)
	if err!=nil {
		return map[string]int64{},
		fmt.Errorf("Failed to get ", aSet, err)
	}

	// For each id, go through and acquire their name and price
	for _, anId:= range setIds{
		stringedId:= strconv.Itoa(anId)

		name, price, err:= GetPrice(consumerKey, consumerSecret,
			stringedId, aClient, foil)
		if err!=nil {
			return map[string]int64{},
			fmt.Errorf("Failed to get ", aSet, anId, err)
		}

		// Record the price
		priceHolster[name] = price

	}

	return priceHolster, nil

}

// Acquires the low EURO price for a product id alongside its name.
//
// MKM has various extras for each set but that isn't a terribly important
// issue.
//
// In the event a foil is requested, the foil low is provided.
// Otherwise, the LOW is returned.
func GetPrice(consumerKey, consumerSecret, productID string,
	aClient *http.Client, foil bool) (name string, price int64, err error) {
	
	productRequest:= productPath + productID

	result, err:= getResource(productRequest,
		consumerKey, consumerSecret, aClient)
	if err!=nil {
		err = fmt.Errorf("Failed to get expansion,", err)
		return
	}

	var productHolster productResponse
	err = json.Unmarshal(result, &productHolster)
	if err!=nil {
		err = fmt.Errorf("Failed to Unmarhsal product,", err)
		return
	}

	var floatingPrice float64
	if foil {
		floatingPrice = productHolster.Product.PriceGuide.LOWFOIL
	}else{
		floatingPrice = productHolster.Product.PriceGuide.LOW
	}

	// Convert the price to Euro-cents
	price = int64(floatingPrice * 100)
	name =  productHolster.Product.Name.English.ProductName

	return

}

// Returns an array of all ids belonging to the set under our namespace.
//
// Set names need to get tweaked as they are sent in so make sure to
// provide a proper cleanToMKM map
func GetExpansion(consumerKey, consumerSecret, set string,
	cleanToMKM map[string]string,
	aClient *http.Client) ([]int, error) {
	
	MKMset, ok:= cleanToMKM[set]
	if !ok{
		return nil,
		fmt.Errorf("Failed to convert expansion to MKM ")
	}
	cleanSet:= strings.Replace(url.QueryEscape(MKMset), "+", "%20", -1)
	fullExpansion:= expansionPath + "/" + cleanSet
	
	result, err:= getResource(fullExpansion,
		consumerKey, consumerSecret, aClient)
	if err!=nil {
		return nil,
		fmt.Errorf("Failed to get expansion,", err)
	}

	var expansionHolster expansionResponse
	err = json.Unmarshal(result, &expansionHolster)
	if err!=nil {
		return nil,
		fmt.Errorf("Failed to Unmarhsal expansion,", err)
	}

	ids:= make([]int, 0)
	for _, aProduct:= range expansionHolster.Card{
		ids = append(ids, aProduct.IdProduct)
	}

	return ids, nil

}

// Acquires a map of our sets to their MKM equivalent
//
// Translates the slightly different MKM sets to mtgjson compatiable			
func GetSetMap(consumerKey, consumerSecret string,
	setList []string,
	aClient *http.Client) (map[string]string, error) {

	result, err:= getResource(expansionPath,
		consumerKey, consumerSecret, aClient)
	if err!=nil {
		return map[string]string{},
		fmt.Errorf("Failed to get expansions,", err)
	}

	var setHolster setListResponse
	err = json.Unmarshal(result, &setHolster)
	if err!=nil {
		return map[string]string{},
		fmt.Errorf("Failed to Unmarhsal expansions,", err)
	}

	// We use a map for easy querying to ensure we have the sets
	// we need
	basicSetMap:= make(map[string]string)
	for _, aSetIdentity:= range setHolster.Expansion{
		// Ensure we are dealing with names localized to our system
		aSetIdentity.setProperName()
		basicSetMap[aSetIdentity.CleanedName] = aSetIdentity.Name
	}

	desiredSetMap:= make(map[string]string)
	for _, aDesiredSet:= range setList{

		// Remove the foil portion as necessary
		aDesiredSet = strings.Replace(aDesiredSet, " Foil", "", -1)

		if isIgnoredSetName(aDesiredSet){
			continue
		}

		MKMName, ok:= basicSetMap[aDesiredSet]
		if !ok {
			return map[string]string{},
			fmt.Errorf("Failed to find set in MKM, ", aDesiredSet)
		}

		desiredSetMap[aDesiredSet] = MKMName
	}

	return desiredSetMap, nil

}