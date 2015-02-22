package categoryBuilder

import(

	"fmt"

	"io/ioutil"
	"encoding/json"

)

// Where we store the intermediate cache for quick, low cost startup.
//
// The non-existence of a cache results in the cache being rebuilt
const cacheFile string = "categoryData.cache.json"
const categoryFile string = "categorySchema.json"

type QueryableCategoryData struct{
	cardsToCategories map[string][]string
	categoriesToCards map[string][]string
}

func GetQueryableCategoryData() QueryableCategoryData {
	usableData:= QueryableCategoryData{}
	usableData.populate()

	return usableData
}

func (usableData *QueryableCategoryData) populate() {

	aLogger:= getLogger("categoryBuilder.log", "categories")

	// See if a cache exists and we can read from it
	cacheData, err := ioutil.ReadFile(cacheFile)	
	if err!=nil {
		// In the event we can't read from the cache, populate it
		populateCache(aLogger)
		cacheData, err = ioutil.ReadFile(cacheFile)
		if err!=nil {
			// If we can't read the cache right after populating it, we quit
			fmt.Println("Failed to read cache after populating it")
			aLogger.Fatalf("Failed to read cache after populating it")
		}
	}

	var cache categoryCache
	err = json.Unmarshal(cacheData, &cache)
	if err!=nil {
		aLogger.Fatalf("Failed to unmarshal cache, ", err)
	}

	usableData.cardsToCategories = cache.CardToCategory
	usableData.categoriesToCards = cache.CategoryToCard

}

// Returns an array of categories corresponding to a provided card name.
//
// An empty array is returned if no categories are found for the card.
func (usableData *QueryableCategoryData) Query(name string) ([]string) {
	categories, ok:= usableData.cardsToCategories[name]
	if !ok {
		categories = make([]string, 0)
	}

	return categories
}

// Returns a map[category name]cards in category. 
func (usableData *QueryableCategoryData) GetCategories() map[string][]string {
	
	return usableData.categoriesToCards

}