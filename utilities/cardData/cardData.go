package main

import(

	"log"
	"os"

	"encoding/json"
	"io/ioutil"

	"sort"
	
	"../mtgjson"

	"./commanderDB"
	"./similarityDB"
	"./categoriesDB"
	"./deckDB"

)

func getAllCardData(aLogger *log.Logger) {
	//acquire mtgjson basic data we get for effectively free
	cardData:= buildBasicData(aLogger)
	stapleOnSetSpecificData(cardData, aLogger)

	
	cardData.addCommanderData()
	cardData.addSimilarityData()

	cardData.addCategoryData()
	cardData.addDeckData()

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
			continue
		}

		aCard.SimilarCards = similarityResults.Others
		aCard.SimilarCardConfidences = similarityResults.Confidences

	}

}

func (cardData *cardMap) addCategoryData() {
	
	categoryData:= categoryBuilder.GetQueryableCategoryData()

	//grab the value of each card
	for _, aCard:= range *cardData {
		
		categories:= categoryData.Query(aCard.Name)

		aCard.Categories = categories
	}

	// Pull down the categories, sort them by commander play,
	// and save them as well
	commanderData:= commanderData.GetQueryableCommanderData()
	completeCategories:= categoryData.GetCategories()
	for aCategory, cards:= range completeCategories{

		// Sorting by desirability is quite helpful
		commanderData.Sort(cards)

		serialCategory, err:= json.Marshal(cards)
		if err!=nil {
			log.Println("Failed to marshal category, ", aCategory , err)	
			return
		}

		usagePath:= dataLoc + string(os.PathSeparator) + aCategory +
		"." + categorySuffix + ".json"

		ioutil.WriteFile(usagePath, serialCategory, 0666)

	}

}

func (cardData *cardMap) addDeckData() error {
	deckData, err:= deckData.GetQueryableDeckData()
	if err!=nil {
		return err
	}

	for _, aCard:= range *cardData{

		result:= deckData.QueryCard(aCard.Name)

		aCard.ModernPlay = result

	}

	return nil
}

// We have a way to add to and sort the ratings of the most used cards
type cardUsageArray []cardUsagePoint

type cardUsagePoint struct{
	Name string
	CommanderUsage float64
}

func (someData cardUsageArray) Len() int {
	return len(someData)
}

// Swap is part of sort.Interface.
func (someData cardUsageArray) Swap(i, j int) {
	someData[i], someData[j] = someData[j], someData[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (someData cardUsageArray) Less(i, j int) bool {
	return someData[i].CommanderUsage > someData[j].CommanderUsage
}


// Initializes and queries the commanderData package for data regarding
// commander usage for each card
//
// Has the side 
func (cardData *cardMap) addCommanderData() {
	
	commanderData:= commanderData.GetQueryableCommanderData()
	
	completeUsage:= make(cardUsageArray, 0)
	var aUsagePoint cardUsagePoint

	// Grab the value of each card
	for _, aCard:= range *cardData {
		
		cardUsage, err:= commanderData.Query(aCard.Name)
		if err!=nil {
			//cards not being present is not at all unusual
			aCard.CommanderUsage = 0.0
			continue
		}

		aCard.CommanderUsage = cardUsage

		aUsagePoint = cardUsagePoint{aCard.Name, cardUsage}
		completeUsage = append(completeUsage, aUsagePoint)

	}

	
	sort.Sort(completeUsage)
	completeUsage = completeUsage[:topCommanderUsageCount]
	serialUsage, err:= json.Marshal(completeUsage)
	if err!=nil {
		log.Println("Failed to marshal commander usage data")	
		return
	}

	usagePath:= dataLoc + string(os.PathSeparator) + topCommanderUsageLoc + ".json"

	ioutil.WriteFile(usagePath, serialUsage, 0666)

}

// Cleans the set names for the cards contained within.
// IE, removed set names we don't support and adds foil sets if available.
func (cardData *cardMap) cleanSetNames(aLogger *log.Logger) {
	
	// Grab the setlist
	setMap, err:= getSupportedSetList()
	if err!=nil {
		aLogger.Fatalln("Failed to acquire supported setlist, ", err)
	}

	// We need to translate from set codes to names
	translator:= getSetCodeToSetNameTranslator(aLogger)
	if err!=nil {
		aLogger.Fatalln("Failed to acquire set code translator, ", err)
	}

	for _, aCard:= range *cardData {
		
		properPrintings:= make([]string, 0)

		for _, aPrinting:= range aCard.Printings{

			aPrinting = translator[aPrinting]

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

//the components of the card we use for exporting
type card struct{
	//What we get for free, or near free, from mtgjson
	Name string
	Text string
	ManaCost string
	Colors []string
	Power, Toughness, Type, ImageName string
	Printings, Types, SuperTypes, SubTypes []string
	Legalities []struct{
		Format string
		Legality string
	}
	Reserved bool
	Loyalty int

	//Extensions we add manually
	CommanderUsage float64
	SimilarCards []string
	SimilarCardConfidences []float64
	Categories []string
	ModernPlay deckData.CardResult
}

//acquires basic card data for our process
func buildBasicData(aLogger *log.Logger) cardMap {

	foreignMap, err:= mtgjson.AllCardsX()
	if err!=nil {
		aLogger.Fatalf("", err)
	}

	// Convert to our extended structure
	// Yes, I feel bad.
	aCardMap:= make(cardMap)
	for _, c:= range foreignMap{
		aCardMap[c.Name] = &card{
			Name: c.Name,
			Text: c.Text,
			ManaCost: c.ManaCost,
			Colors: c.Colors,
			Power: c.Power,
			Toughness: c.Toughness,
			ImageName: c.ImageName,
			Printings: c.Printings,
			Type: c.Type,
			Types: c.Types,
			SuperTypes: c.SuperTypes,
			SubTypes: c.SubTypes,
			Legalities: c.Legalities,
			Reserved: c.Reserved,
			Loyalty: c.Loyalty,
		}
	}

	return aCardMap

}