package currencyConversion

import(

	"fmt"

	"net/http"
	"encoding/json"
	"io/ioutil"

)

// Where to GET our currency data from
const ratioPath string =
"http://openexchangerates.org/api/latest.json?app_id="

// What the base ratio that we want to deal with is
const desiredBase string = "USD"

// A typical response from openexchangerates.org
type conversionResponse struct {
        Base  string `json:"base"`
        Rates struct {
                EUR float64 `json:"EUR"`
        } `json:"rates"`
        Timestamp int `json:"timestamp"`
}

// Returns a conversion ratio to go from EUR TO USD
func EURToUSD(apiKey string) (float64, error) {
	
	ratioResponse, err:= getConversionRatio(apiKey)
	if err!=nil {
		return 0,
		fmt.Errorf("Failed to get conversion data, ", err )
	}

	if ratioResponse.Base != desiredBase {
		return 0,
		fmt.Errorf("Got undesired base from openexchangerates, got: ", 
			ratioResponse.Base)
	}

	if ratioResponse.Rates.EUR <= 0 {
		return 0,
		fmt.Errorf("Got bad conversion from openexchangerates, got: ", 
			ratioResponse.Rates.EUR)		
	}

	ratio:= 1 / ratioResponse.Rates.EUR

	return ratio, nil

}

// Acquires a conversionResponse from openexchangerates
func getConversionRatio(apiKey string) (conversionResponse, error) {
	fullPath:= ratioPath + apiKey

	resp, err:= http.Get(fullPath)
	if err!=nil {
		return conversionResponse{},
		fmt.Errorf("Failed to GET conversion data, ", err)
	}
	defer resp.Body.Close()
	
	data, err:= ioutil.ReadAll(resp.Body)
	if err!=nil {
		return conversionResponse{},
		fmt.Errorf("Failed to read conversion data, ", err)
	}

	var aResponse conversionResponse
	err = json.Unmarshal(data, &aResponse)
	if err!=nil {
		return conversionResponse{},
		fmt.Errorf("Failed to Unmarshal conversion data, ", err)
	}

	return aResponse, nil

}