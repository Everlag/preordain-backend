package main

import(

	"log"
	"fmt"

	"./../influxdbHandler"

	"./priceSources"

	"os"
	"strconv"
	"io/ioutil"
	"encoding/json"

	"time"
)

// How many sources of prices we deal with
const priceSourceQuantity int = 1

// Where we store our failed uploads
const uploadFailureStorage string = "uploadFailures"

// Where our influxdb client data is kept
const influxdbCredentials string = "influxdbCredentials.json"

func main() {

	aLogger:= priceSources.GetLogger("priceWriter.log", "priceWriter")
	
	creds, err:= getCredentials()
	if err!=nil {
		aLogger.Fatalln(err)
	}

	aLogger.Println("Credentials read")

	
	aClient, err:= influxdbHandler.GetClient(creds.RemoteLocation, creds.DBName,
	 creds.User, creds.Pass,
	 creds.Read, creds.Write)
	if err!=nil {
		aLogger.Fatalln("Failed to ping remote server at client creation, ", err)
	}

	aLogger.Println("Influxdb client active")

	// Imports from a standard price source mtgprice seeds
	//Import(aClient)
	//os.Exit(0)

	RunPriceLoop(aClient, aLogger)
	
}

// Attempts to run an update once an hour for each available price source.
// It is likely that rate limiting will be experienced so we log any errors
// received that aren't explicity rate limiting
func RunPriceLoop(aClient *influxdbHandler.Client, aLogger *log.Logger) {
	aLogger.Println("Starting price loop")

	for{

		magiccardmarket(aClient, aLogger)
		mtgprice(aClient, aLogger)

		time.Sleep(time.Hour * time.Duration(1))

	}
}

func mtgprice(aClient *influxdbHandler.Client, aLogger *log.Logger) {

	prices, err:= priceSources.GetmtgpricePrices(aLogger)
	if err!=nil {
		if err.Error()!=priceSources.RateExceeded {
			aLogger.Println(err)	
		}else{
			aLogger.Println("Sleeping for update, mtgprice")
		}
		return
	}

	uploadPriceResults([]priceSources.PriceMap{prices}, aClient, aLogger)

}

func magiccardmarket(aClient *influxdbHandler.Client, aLogger *log.Logger) {
	
	prices, err:= priceSources.GetMKMPrices(aLogger)
	if err!=nil {
		if err.Error()!=priceSources.RateExceeded {
			aLogger.Println(err)	
		}else{
			aLogger.Println("Sleeping for update, magiccardmarket")
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

			// Deal with the fact that some price sources may have multiple
			// currencies that were massaged into USD
			var aPoint influxdbHandler.Point
			if aPriceResult.HasEuro {
				// An original price in euros is recorded alongside the USD
				// conversion
				euroPrice:= aPriceResult.EURPrices[aSetName][aCardName]
				
				aPoint = influxdbHandler.BuildPointWithEuro(aCardName,
					aPriceResult.Time, aPrice, euroPrice,
					aSetName, aPriceResult.Source)
			
			}else{
			
				aPoint = influxdbHandler.BuildPoint(aCardName,
					aPriceResult.Time, aPrice, aSetName, aPriceResult.Source)
			
			}

			points = append(points, aPoint)

		}
	}

	// Send the points to the db
	err:= aClient.SendPoints(points)

	return err

}

type credentials struct{
	RemoteLocation, DBName, User, Pass string

	Write, Read bool
}

func getCredentials() (credentials, error) {
	
	data, err:=ioutil.ReadFile(influxdbCredentials)
	if err!=nil {
		return credentials{}, fmt.Errorf("Failed to read influxdbCredentials, ", err)
	}

	var someCreds credentials
	err = json.Unmarshal(data, &someCreds)
	if err!=nil {
		return credentials{}, fmt.Errorf("Failed to unmarshal influxdbCredentials, ", err)
	}

	return someCreds, nil

}