package categoryBuilder

import(

	"log"

	"io/ioutil"
	"encoding/json"

	"strings"

)

// The raw, on disk representation
type cardMap map[string]card

type card struct{
	Name string
	Text string
}

// Our category definition schema
// Conventions:
// Colors are replaced with {C}
// Card names are transformed to ~
// Digits are transformed to 1 -> NOTE X is MUST also be considered a digit!
//
// Top level categories have sub categories which are
// a subset of the top level.
//
// Text items must appear all appear in some line of
// the card's text, they need not appear in all of them -> so look at all the text as one item!
//
// The text items need to all be either uppercased or lowercased
// for reasonable comparison

type categoriesMap map[string]topLevelCategory

type topLevelCategory struct{
	// Some top level categories have a map of sub-categories
	Categories map[string]category
	NoPrefix bool

	// Others don't fit into other categories and are delcared at the top level
	Text []string
}

type category struct{
	// Each chunk of text must be present on the card for it to be
	// classified as inside that category
	Text []string
}

// Our cache structure, very similar to QueryableCategoryData
type categoryCache struct{

	CardToCategory, CategoryToCard map[string][]string
	categories flatCategoryMap
}

// An easier to iterate on category map.
//
// Top level categories are compressed down to TopLevel - SubLevel
type flatCategoryMap map[string]category

func (aMap categoriesMap) flatten() flatCategoryMap {

	flat:= make(flatCategoryMap)

	var freshCategory category
	var freshName string
	for aTopLevelName, aTopLevelCategory:= range aMap{

		// This is a single level category
		if len(aTopLevelCategory.Text) > 0{
			freshCategory = category{}
			freshCategory.Text = make([]string, len(aTopLevelCategory.Text))
			for i, someText:= range aTopLevelCategory.Text{
				freshCategory.Text[i] = cleanCardText(someText, "noName")
			}

			flat[aTopLevelName] = freshCategory

		}else{

			for secondaryName, secondaryCategory := range aTopLevelCategory.Categories{
				freshCategory = category{}
				freshCategory.Text = make([]string, len(secondaryCategory.Text))

				for i, someText:= range secondaryCategory.Text{
					freshCategory.Text[i] = cleanCardText(someText, "noName")
				}

				if aTopLevelCategory.NoPrefix {
					freshName = secondaryName
				}else{
					freshName = aTopLevelName + " - " + secondaryName
				}

				flat[freshName] = freshCategory
			}
		}


	}

	return flat
}

func freshCache(aCategoriesMap categoriesMap) categoryCache {

	aCache:= categoryCache{}
	aCache.CardToCategory = make(map[string][]string)
	aCache.CategoryToCard = make(map[string][]string)

	// We process the categories first to make computation easier
	aCache.categories = aCategoriesMap.flatten()
	for aCategoryName:= range aCache.categories{
		aCache.CardToCategory[aCategoryName] = make([]string, 0)
	}

	return aCache
}

// Checks a card against the categories of the cache
// and stores it with its categories if it has one.
func (aCache *categoryCache) addCard(aCard card) {
	
	// Normalize the text
	cardText:= cleanCardText(aCard.Text, aCard.Name)
	name:= aCard.Name

	aCache.CardToCategory[name] = make([]string, 0)

	filterNotFound:= false
	for aCategory, filters:= range aCache.categories{

		filterNotFound = false
		for _, aFilter:= range filters.Text{
			if (!strings.Contains(cardText, aFilter)){
				filterNotFound = true
				break
			}
		}

		if !filterNotFound {
			aCache.CardToCategory[name] = append(aCache.CardToCategory[name],
				aCategory)
			aCache.CategoryToCard[aCategory] = append(aCache.CategoryToCard[aCategory],
				name)
		}

	}

}

func populateCache(aLogger *log.Logger) {
	
	// Grab the card data hosted on disk
	cardData, err:= ioutil.ReadFile("AllCards-x.json")
	if err!=nil {
		aLogger.Fatalf("Failed to read AllCards-x.json")
	}

	// Unmarshal it into a map of string to card with relevant data
	var aCardMap cardMap
	err = json.Unmarshal(cardData, &aCardMap)
	if err!=nil {
		aLogger.Fatalf("Failed to unmarshal card list, ", err)
	}

	// Grab the category data hosted on disk
	categoryData, err:= ioutil.ReadFile(categoryFile)
	if err!=nil {
		aLogger.Fatalf("Failed to read in category definitions, ", err)
	}

	// Unmarshal it into a category map
	var categoryMap categoriesMap
	err = json.Unmarshal(categoryData, &categoryMap)
	if err!=nil {
		aLogger.Fatalf("Failed to unmarshal category definitions, ", err)
	}

	// Go through and tag all the cards
	cache:= freshCache(categoryMap)
	for _, aCard:= range aCardMap{
		cache.addCard(aCard)
	}

	cacheData, err:= json.Marshal(cache)
	if err!=nil {
		aLogger.Fatalf("Failed to marshal cache data, ", err)
	}

	err = ioutil.WriteFile(cacheFile, cacheData, 0777)
	if err!=nil {
		aLogger.Fatalf("Failed to send cache to disk, ", err)
	}

}