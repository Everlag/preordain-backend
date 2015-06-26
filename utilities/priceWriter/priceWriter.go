package main

import(

	"log"

	"./../influxdbHandler"

	"./priceSources"

	"time"
)

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