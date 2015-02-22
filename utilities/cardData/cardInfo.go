package main
//Hooks together the following datasources and outputs multiple directories
//containing everything needed for static mtg finance analysis except prices.
//
//Sources used are:
//
//mtgjson for inherent card info
//mtgsalvation for commander usage
//mtgimage for card images UNTESTED
//ngrams for card similarity

import(

	"log"
	"os"
	"fmt"
	"io"

	"./imageDB"
)

// The various locations our derived data goes
const dataLoc string = "cardText/"
const imageLoc string = "cardFulls/"
const cropLoc string =  "cardCrops/"
const symbolsLoc string = "cardSymbols/"

// The location and top count of the commander data we release
const topCommanderUsageLoc string = "commanderUsage"
const topCommanderUsageCount int = 1000

const categorySuffix string = "category"

func main() {
	aLogger:= getLogger("core.log")

	// Dumps into dataLoc the data for each card
	getAllCardData(aLogger)

	
	// Dumps into dataLoc the set data for each set
	getAllSetData(aLogger)

	imageScraper.ScrapeImages(imageLoc, cropLoc, symbolsLoc)
	
}

func getLogger(fName string) (aLogger *log.Logger) {
	file, err:= os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err!=nil {
		fmt.Println("Starting logger failed, cannot write to logger to say logger failed. Uh oh.")
		fmt.Println(err)
		os.Exit(0)
	}

	multi:= io.MultiWriter(file, os.Stdout)

	aLogger = log.New(multi, "User ", log.Ldate|log.Ltime|log.Lshortfile)

	return
}