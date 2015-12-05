package main

import(

	"log"
	"os"

	"encoding/json"
	"io/ioutil"

	"sort"
	
	"../../common/mtgjson"
	"../../common/setlist"

	"./commanderDB"
	"./deckDB"

)

func getAllCardData(aLogger *log.Logger) {
	// Base from mtgjson
	cardData:= buildBasicData(aLogger)
	
	// Set data such as reserved list status
	stapleOnSetSpecificData(cardData, aLogger)

	// Commander stats
	cardData.addCommanderData()

	// Modern deck data
	cardData.addDeckData()

	// Cleanup set names
	cardData.cleanSetNames(aLogger)

	// Send everything to disk
	cardData.dumpToDisk(aLogger)
}

// Commits each card in the cardMap to disk under name.json
func (cardData *cardMap) dumpToDisk(aLogger *log.Logger) {
	
	aLogger.Println("Commencing dump to disk of cardMap")

	var serialCard []byte
	var err error

	var cardPath string
	baseLoc:= dataLoc()

	for name, aCard:= range *cardData {

		// Serialize
		serialCard, err= json.Marshal(aCard)
		if err!=nil {
			aLogger.Println("Failed to marshal ", name)	
			continue
		}

		// Store
		cardPath = baseLoc + string(os.PathSeparator) + name + ".json"

		ioutil.WriteFile(cardPath, serialCard, 0666)

	}

	aLogger.Println("Dump complete")

}

// Apply modern deck data to the map
//
// Has the side effect of causing a long cache population
// if not cache is found
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


// Apply commander meta data to each card
//
// Has the side effect of causing a long
// cache population if the cache is unavailable.
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

	usagePath:= dataLoc() + string(os.PathSeparator) + topCommanderUsageLoc + ".json"

	ioutil.WriteFile(usagePath, serialUsage, 0666)

}

// Remove set names we don't support and add foil variants.
func (cardData *cardMap) cleanSetNames(aLogger *log.Logger) {
	
	// Grab the setlist
	setMap, err:= setlist.FoilMapping()
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

type cardMap map[string]*card

// Format of serialized card
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
	ModernPlay deckData.CardResult
}

// Acquire the baseline data for each card
//
// Requires a translation of the mtgjon structure
// to our extended format.
func buildBasicData(aLogger *log.Logger) cardMap {

	// Baseline
	foreignMap, err:= mtgjson.AllCardsX()
	if err!=nil {
		aLogger.Fatalf("", err)
	}

	// Translate to extended structure
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