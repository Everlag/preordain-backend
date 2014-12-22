package main

import(

	"log"

	"./influxdbHandler"

	"./priceSources"

	"os"
	"strconv"
	"io/ioutil"
	"encoding/json"

	"time"
)

const serverLoc string = "http://192.168.56.101:8086"
const dbName string = "goPrices"
const user string = "writer"
const pass string = "U6QQ5gsvy9NxuJFn9kPN"

// How many sources of prices we deal with
const priceSourceQuantity int = 1

const uploadFailureStorage string = "uploadFailures"

func main() {
	aClient:= influxdbHandler.GetClient(serverLoc, dbName,
	 user, pass,
	 true, true)
	
	aLogger:= priceSources.GetLogger("priceWriter.log", "priceWriter")

	RunPriceLoop(aClient, aLogger)
	
}

// Attempts to run an update once an hour for each available price source.
// It is likely that rate limiting will be experienced so we log any errors
// received that aren't explicity rate limiting
func RunPriceLoop(aClient *influxdbHandler.Client, aLogger *log.Logger) {
	aLogger.Println("Starting price loop")

	for{

		mtgprice(aClient, aLogger)

		time.Sleep(time.Hour * time.Duration(1))

	}
}

func mtgprice(aClient *influxdbHandler.Client, aLogger *log.Logger) {

	prices, err:= priceSources.GetmtgpricePrices(aLogger)
	if err!=nil {
		if err.Error()!=priceSources.RateExceeded {
			aLogger.Println(err)	
		}
		return
	}

	uploadPriceResults([]priceSources.PriceMap{prices}, aClient, aLogger)

}

// Attempts to upload all price results for all provided sources
//
// Logs Failures to upload
func uploadPriceResults(pricingResults []priceSources.PriceMap,
	aClient *influxdbHandler.Client,
	aLogger *log.Logger) {

	for _, aPriceResult:= range pricingResults{

		aLogger.Println("Uploading data from ", aPriceResult.Source)

		err:= uploadSingleSourceResults(aPriceResult, aClient)
		if err!=nil {
			// Write the data to a local file so we don't lose it
			storeFailedUpload(aPriceResult, err, aLogger)
			continue
		}

		aLogger.Println("Completed Upload of data from ", aPriceResult.Source)

	}

}

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

func uploadSingleSourceResults(aPriceResult priceSources.PriceMap,
	aClient *influxdbHandler.Client) error {

	// Construct the points to send
	points:= make([]influxdbHandler.Point, 0)

	for aSetName, cardMap:= range aPriceResult.Prices{

		for aCardName, aPrice:= range cardMap{

			aPoint:= influxdbHandler.BuildPoint(aCardName,
				aPriceResult.Time, aPrice, aSetName, aPriceResult.Source)

			points = append(points, aPoint)

		}
	}

	// Send the points to the db
	err:= aClient.SendPoints(points)

	return err

}