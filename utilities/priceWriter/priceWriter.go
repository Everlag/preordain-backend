package main

import(

	"log"
	"time"

	"github.com/Everlag/preordain-backend/api/Prices/ApiServices/priceDBHandler.v2"

	"./priceSources"

	"github.com/jackc/pgx"

	"os"
	"fmt"
	"github.com/joho/godotenv"
)

func main() {

	// Populate config locations not explicitly set
	envError:= godotenv.Load("priceWriter.default.env")
	if envError!=nil {
		fmt.Println("failed to parse prices.default.env")
		os.Exit(1)
	}

	aLogger:= priceSources.GetLogger("priceWriter.log", "priceWriter")

	
	pool, err:= priceDB.Connect()
	if err!=nil {
		aLogger.Fatalln("Failed to ping remote server at client creation, ", err)
	}

	aLogger.Println("priceDB client active")

	RunPriceLoop(pool, aLogger)
	
}

// Attempts to run an update once an hour for each available price source.
// It is likely that rate limiting will be experienced so we log any errors
// received that aren't explicity rate limiting
func RunPriceLoop(pool *pgx.ConnPool, aLogger *log.Logger) {
	aLogger.Println("Starting price loop")

	for{

		magiccardmarket(pool, aLogger)
		mtgprice(pool, aLogger)

		time.Sleep(time.Hour * time.Duration(1))

	}
}

func mtgprice(pool *pgx.ConnPool, aLogger *log.Logger) {

	prices, err:= priceSources.GetmtgpricePrices(aLogger)
	if err!=nil {
		if err.Error()!=priceSources.RateExceeded {
			aLogger.Println(err)	
		}else{
			aLogger.Println("Sleeping for update, mtgprice")
		}
		return
	}

	uploadPriceResults([]priceSources.PriceMap{prices}, pool, aLogger)

}

func magiccardmarket(pool *pgx.ConnPool, aLogger *log.Logger) {
	
	prices, err:= priceSources.GetMKMPrices(aLogger)
	if err!=nil {
		if err.Error()!=priceSources.RateExceeded {
			aLogger.Println(err)	
		}else{
			aLogger.Println("Sleeping for update, magiccardmarket")
		}
		return
	}

	uploadPriceResults([]priceSources.PriceMap{prices}, pool, aLogger)

}

// Attempts to upload all price results for all provided sources
//
// Logs Failures to upload
func uploadPriceResults(pricingResults []priceSources.PriceMap,
	pool *pgx.ConnPool,
	aLogger *log.Logger) {

	for _, aPriceResult:= range pricingResults{

		aLogger.Println("Uploading data from ", aPriceResult.Source)

		err:= uploadSingleSourceResults(aPriceResult, pool)
		if err!=nil {
			// Write the data to a local file so we don't lose it
			storeFailedUpload(aPriceResult, err, aLogger)
			continue
		}

		aLogger.Println("Completed Upload of data from ", aPriceResult.Source)

	}

}