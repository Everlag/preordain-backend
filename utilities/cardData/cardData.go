package main

import(

	"log"
	"os"

	"encoding/json"
	"io/ioutil"
	
	"./commanderDB"
	"./similarityDB"

)

func getAllCardData(aLogger *log.Logger) {
	//acquire mtgjson basic data we get for effectively free
	cardData:= buildBasicData(aLogger)
	stapleOnSetSpecificData(cardData, aLogger)

	cardData.addCommanderData()
	cardData.addSimilarityData()

	cardData.cleanSetNames(aLogger)

	cardData.dumpToDisk(aLogger)
}

// dumpToDisk commits each value of the card map and dumps it into
// the dataLoc folder under the name.json file
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

// Cleans the set names for the cards contained within.
// IE, removed set names we don't support and adds foil sets if available.
func (cardData *cardMap) cleanSetNames(aLogger *log.Logger) {
	
	// Grab the setlist
	setMap, err:= getSupportedSetList()
	if err!=nil {
		aLogger.Fatalln("Failed to acquire supported setlist, ", err)
	}

	for _, aCard:= range *cardData {
		
		properPrintings:= make([]string, 0)

		for _, aPrinting:= range aCard.Printings{

			foilName, ok:= setMap[aPrinting]
			if !ok {
				// Not ok means this is a set not in the setlist
				continue
			}

			// Add the non-foil name
			properPrintings = append(properPrintings, aPrinting)
			
			// And the foil if available
			if foilName != "" {
				properPrintings = append(properPrintings, foilName)	
			}

		}

		if len(properPrintings) == 0 {
			aLogger.Println("Failed to get any printings for ", aCard.Name)
		}

		aCard.Printings = properPrintings

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