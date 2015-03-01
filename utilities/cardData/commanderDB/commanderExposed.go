package commanderData

import(

	"fmt"

	"encoding/json"
	
	"io/ioutil"

	"sort"

)

//we have places where we put our data
const cacheFile string = "commanderData.cache.json"

//an exposable type that allows users to query for data regarding
//our commander data.
type QueryableCommanderData struct{
	data map[string]int
	Accuracy int
	ComparisonCard string
}

func GetQueryableCommanderData() QueryableCommanderData {
	usableData:= QueryableCommanderData{}
	usableData.populate()

	return usableData
}

//attempts to query the usable commander data for card appearances
func (usableData *QueryableCommanderData) Query(name string) (float64, error) {
	usableName:= normalizeCardName(name)
	appearance, ok:= usableData.data[usableName]

	if !ok {
		return 0, fmt.Errorf("Card not present in data")
	}

	percentValue:= float64(appearance) / float64(usableData.Accuracy)

	return percentValue, nil
}

// Attempts to sort cards based on their commander usage. Cards with no
// recorded use are considered to have 0.0
func (usableData *QueryableCommanderData) Sort(names []string) []string {
	
	// Build the card items for sorting
	items:= make(cardItems, len(names))
	for i, aName:= range names{

		items[i] = cardItem{
			Name: aName,
			QueryableData: usableData,
		}

	}

	sort.Stable(sort.Reverse(items))

	// Convert back to strings
	for i, anItem:= range items{
		names[i] = anItem.Name
	}

	return names
	
}

//populates the QueryableCommanderData with mtgsalvation data.
//
//a cache file is kept at cacheFile
func (usableData *QueryableCommanderData) populate() {
	
	aLogger:= getLogger("deckScraper.log", "deckScraper")

	//see if a cache exists and we can read from it
	cacheData, err := ioutil.ReadFile(cacheFile)	
	if err!=nil {
		//in the event we can't read from the cache, populate it
		populateRawCache(aLogger)
		cacheData, err = ioutil.ReadFile(cacheFile)
		if err!=nil {
			//if we can't read the cache right after populating it, we quit
			fmt.Println("Failed to read cache after populating it")
			aLogger.Fatalf("Failed to read cache after populating it")
		}
	}

	//unmarshal the cache
	var count map[string]int
	err = json.Unmarshal(cacheData, &count)
	if err != nil {
		aLogger.Fatalf("Failed to unmarshal cache")
	}
	
	//populate a non-normalized list of card datas with counts
	rawSortedCount:= make(cardDataCollection, len(count))
	i:= 0
	for cardName, count:= range count{
		rawSortedCount[i] = cardData{
			Name:cardName,
			Appearance: count,
		}
		i++
	}

	//sort the non-normalized list
	sort.Sort(rawSortedCount)

	//convert the list into a normalized entity
	mostUsedCard:= rawSortedCount[len(rawSortedCount) - 1]

	normalizedSortedCount:= make(cardDataCollection, len(count))
	for i := 0; i < len(count); i++ {
		normalizedSortedCount[i] = rawSortedCount[i]
		normalizedSortedCount[i].Appearance =
		int(
			float64(rawSortedCount[i].Appearance) /
			float64(mostUsedCard.Appearance) *
			granularity)
	}

	//then we convert this normalized count into a map for easy querying
	normalizedCountMap:= make(map[string]int)

	for i := 0; i < len(count); i++ {
		normalizedCountMap[normalizedSortedCount[i].Name] = normalizedSortedCount[i].Appearance
	}

	usableData.data = normalizedCountMap
	usableData.Accuracy = int(granularity)
	usableData.ComparisonCard = mostUsedCard.Name

}