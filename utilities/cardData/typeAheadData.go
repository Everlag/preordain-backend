package main

import(

	"log"

	"strings"
	"sort"

	"os"
	"io/ioutil"
	"encoding/json"

	"./commanderDB"

)

// Generates and outputs typeAhead data to typeAheadLoc()
func getAllTypeAheadData(aLogger *log.Logger) {
	
	// Generate
	aTypeAhead:= buildTypeAheadCardData(aLogger)
	// Output
	aTypeAhead.dumpToDisk(aLogger)
}


func buildTypeAheadCardData(aLogger *log.Logger) (typeAhead) {
	
	aTypeAhead:= make(typeAhead)

	// List Cards
	cardList:= getRawCardNames(aLogger)

	// Add the cards
	aTypeAhead.addList(cardList)

	// Sort by commander use
	commanderData:= commanderData.GetQueryableCommanderData()
	aTypeAhead.sortByCommanderUsage(&commanderData)

	return aTypeAhead
}

func getRawCardNames(aLogger *log.Logger) ([]string) {

	// Acquire the map of card names
	cardsMap:= buildBasicData(aLogger)

	cardList:= make([]string, 0)
	for aCardName:= range cardsMap{
		cardList = append(cardList, aCardName)
	}

	return cardList
}

// A map[text]options.
type typeAhead map[string][]string

// Adds a list of strings to the typeahead.
func (aTypeAhead *typeAhead) addList(names []string) {
	// Allows us to index
	valueTypeAhead:= *aTypeAhead

	var key string
	for _, aName:= range names{

		// Replace the special case of AEther cards
		aName = strings.Replace(aName, "Æ", "AE", -1)
		aLowerName:= strings.ToLower(aName)

		// Develop subarrays for each depth of key
		for keyIndexEnd := 1; keyIndexEnd < len(aName) + 1; keyIndexEnd++ {
			
			if keyIndexEnd > len(aName) {
				break
			}
			key = aLowerName[0:keyIndexEnd]

			_, ok:= valueTypeAhead[key]
			if !ok {
				valueTypeAhead[key] = make([]string, 0)
			}

			valueTypeAhead[key] = append(valueTypeAhead[key], aName)

		}

	}
}

// Sorts all fields of the typeAhead based on commander usage
//
// Additionally, each field is pre-sorted alphabetically so cards
// without significant commander usage can have some order.
//
// This assumes that commanderUsage uses a STABLE sort.
func (aTypeAhead *typeAhead) sortByCommanderUsage(commanderUsage *commanderData.QueryableCommanderData) {
	
	for aKey, names:= range *aTypeAhead{

		sort.Strings(names)

		(*aTypeAhead)[aKey] = commanderUsage.Sort(names)

	}

}

// Dumps each stored typeahead query to typeAheadLoc() as
// $QUERY.json 
func (aTypeAhead *typeAhead) dumpToDisk(aLogger *log.Logger) {

	var serialChoices []byte
	var err error

	var path string

	for aKey, names:= range *aTypeAhead {

		serialChoices, err= json.Marshal(names)
		if err!=nil {
			aLogger.Println("Failed to marshal ", aKey)	
			continue
		}

		path = typeAheadLoc() + string(os.PathSeparator) + aKey + ".json"

		err = ioutil.WriteFile(path, serialChoices, 0666)
		if err!=nil {
			aLogger.Println("Failed to write choices, ", err)
		}

	}

}