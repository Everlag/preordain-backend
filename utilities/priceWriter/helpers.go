package main

import(

	"log"

	"./priceSources"

	"os"
	"strconv"
	"io/ioutil"
	"encoding/json"
	
)

// Where we store our failed uploads
const uploadFailureStorage string = "uploadFailures"

// Stores prices in the uploadFailureStorage directory under the name
// timestamp + sourceName which allows us to recover this data at a later time
func storeFailedUpload(aPriceResult priceSources.PriceMap, err error,
	aLogger *log.Logger) {

	// Derive the location of this data
	failureLocation:= uploadFailureStorage + string(os.PathSeparator) +
	strconv.FormatInt(aPriceResult.Time, 10) + aPriceResult.Source

	// Report that we are handling the failure
	aLogger.Println("UPLOAD FAILURE -", aPriceResult.Source,
	"  - ", err, " Dumping to ", failureLocation)

	resultData, err:= json.Marshal(aPriceResult)
	if err!=nil {
		aLogger.Println("STORAGE FAILURE -", aPriceResult.Source,
		"  - ", err)
		return
	}

	ioutil.WriteFile(failureLocation, resultData, 0666)
}