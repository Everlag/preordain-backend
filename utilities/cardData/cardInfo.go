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

	"encoding/json"
	"io/ioutil"

	"./imageDB"
	"./commanderDB"
	"./similarityDB"

)

//the various locations our derived data goes
const dataLoc string = "cardText/"
const imageLoc string = "cardFulls/"
const cropLoc string =  "cardCrops/"
const symbolsLoc string = "cardSymbols/"

func main() {
	aLogger:= getLogger("core.log")

	//acquire mtgjson basic data we get for effectively free
	cardData:= buildBasicData(aLogger)
	stapleOnSetSpecificData(cardData, aLogger)

	cardData.addCommanderData()
	cardData.addSimilarityData()

	cardData.dumpToDisk(aLogger)

	imageScraper.ScrapeImages(imageLoc, cropLoc, symbolsLoc)

	fmt.Println(cardData["Chromatic Lantern"])
	fmt.Println(cardData["Dimir Signet"])

}

//dumpToDisk commits each value of the card map and dumps it into
//the dataLoc folder under the name.json file
func (cardData *cardMap) dumpToDisk(aLogger *log.Logger) {
	
	aLogger.Println("Commencing dump to disk of cardMap")

	var serialCard []byte
	var err error

	var cardPath string

	for name, aCard:= range *cardData {

		serialCard, err= json.Marshal(aCard)
		if err!=nil {
			aLogger.Println("Failed to marshal ", name)	
			continue
		}

		cardPath = dataLoc + string(os.PathSeparator) + name + ".json"

		ioutil.WriteFile(cardPath, serialCard, 0666)

	}

	aLogger.Println("Dump complete")

}

func (cardData *cardMap) addSimilarityData() {
	
	similarityData:= similarityBuilder.GetQueryableSimilarityData()

	//grab the value of each card
	for _, aCard:= range *cardData {
		
		similarityResults, err:= similarityData.Query(aCard.Name)
		if err!=nil {
			//cards not being present is not at all unusual
			aCard.CommanderUsage = 0.0
			continue
		}

		aCard.SimilarCards = similarityResults.Others
		aCard.SimilarCardConfidences = similarityResults.Confidences

	}

}

//initializes and queries the commanderData package for data regarding
//commander usage for each card
func (cardData *cardMap) addCommanderData() {
	
	commanderData:= commanderData.GetQueryableCommanderData()
	//grab the value of each card
	for _, aCard:= range *cardData {
		
		cardUsage, err:= commanderData.Query(aCard.Name)
		if err!=nil {
			//cards not being present is not at all unusual
			aCard.CommanderUsage = 0.0
			continue
		}

		aCard.CommanderUsage = cardUsage

	}

}

//we seed our initial card data off of AllCard-x.json
type cardMap map[string]*card

//the components of the card we use for similarity determination
type card struct{
	//What we get for free, or near free, from mtgjson
	Name string
	Text string
	ManaCost string
	Colors []string
	Power, Toughness, Type, ImageName string
	Printings, Types, SuperTypes, SubTypes []string
	Legalities map[string]string
	Reserved bool
	Loyalty int

	//Extensions we add manually
	CommanderUsage float64
	SimilarCards []string
	SimilarCardConfidences []float64
}

//acquires basic card data for our process
func buildBasicData(aLogger *log.Logger) cardMap {
	
	//grab the card data hosted on disk
	cardData, err:= ioutil.ReadFile("AllCards-x.json")
	if err!=nil {
		aLogger.Fatalf("Failed to read AllCards-x.json")
	}

	//unmarshal it into a map of string to card with relevant data
	var aCardMap cardMap
	err = json.Unmarshal(cardData, &aCardMap)
	if err!=nil {
		aLogger.Fatalf("Failed to unmarshal AllCards-x.json, ", err)
	}

	return aCardMap

}

func getLogger(fName string) (aLogger *log.Logger) {
	file, err:= os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err!=nil {
		fmt.Println("Starting logger failed, cannot write to logger to say logger failed. Oh god.")
		fmt.Println(err)
		os.Exit(0)
	}

	multi:= io.MultiWriter(file, os.Stdout)

	aLogger = log.New(multi, "User ", log.Ldate|log.Ltime|log.Lshortfile)

	return
}