package commanderData
//scrapes a list of decks from the mtg commander database.
//
//proceeds to grab the deck at each url provided by the list of decks

import(

	"os"
	"fmt"
	"log"

	gq "github.com/PuerkitoBio/goquery"

	"strings"

	"html"

	"encoding/json"
	
	"io/ioutil"
)

//the granularity of appearances as an integer mapping to a percentage between
//0 and 100%
const granularity float64 = 10000

//a place to unmarshal the acquired json to
type salvationDeck struct{
	Deck []Card
}

type Card struct{
	CardName string
}

//where we keep the high level commander db
const commanderDBLocation string = "http://www.mtgsalvation.com/forums/the-game/commander-edh/204408-commander-decklist-database"

//the location where the json encoded data for each deck is stored.
const deckDataAttribute string = "data-card-list"
const deckClass string = "deck"
const deckElementType string = "table"

//returns an array of deck urls as found at commanderDBLocation
func getDeckUrls() ([]string, error) {
	var doc *gq.Document
	var e error

	doc, e = gq.NewDocument(commanderDBLocation)
	if e!= nil{
		return nil, e
	}

	deckUrls := make([]string, 0)

	//in each spoiler section, find links.
	doc.Find(".spoiler").Find("a").Each(func(i int, s *gq.Selection){

		//grab the class to make sure we aren't trying to work with card pages
		class, _:= s.Attr("class")

		if strings.Contains(class, "card-link"){
			return
		}

		//grab the href otherwise and that is a deck we'll deal with
		deckHref, exists:= s.Attr("href")
		if exists==false{
			return
		}

		deckUrls = append(deckUrls, deckHref)

	})

	return deckUrls, nil

}

//adds to the target map the cards found here
//
//if it fails, it will log the failure and return an empty array
func getDeckList(deckLoc string, target map[string]int, aLogger *log.Logger){
	
	var doc *gq.Document
	var e error

	doc, e = gq.NewDocument(deckLoc)
	if e!= nil{
		aLogger.Println("Failed to open ", deckLoc)
		return
	}

	doc.Find("table").EachWithBreak(func(i int, s *gq.Selection) bool{

		//grab json encoded list of cards
		deckEncoded, exists:= s.Attr(deckDataAttribute)
		if exists==false{
			fmt.Println("Candidate deck failed for ", deckLoc)
			aLogger.Println("Candidate deck failed for ", deckLoc)
			return true
		}

		deckDecoded:= html.UnescapeString(deckEncoded)

		//attempt to unmarshal it
		var aDeck salvationDeck
		err:= json.Unmarshal([]byte(deckDecoded), &aDeck)
		if err!=nil {
			fmt.Println("Failed to unmarshal deck for ", deckLoc)
			aLogger.Println("Failed to unmarshal deck for ", deckLoc)
			return true
		}
		

		for _, aCard := range aDeck.Deck{
			//sets the card name to lower case to normalize across user
			//capitalization errors
			effectiveName := normalizeCardName(aCard.CardName)

			_, exists:= target[effectiveName]

			if !exists {
				target[effectiveName] = 1
			}else{
				target[effectiveName]++
			}

		}

		//signal the end of this iteration
		return false

	})

	return

}

func populateRawCache(aLogger *log.Logger) {
	
	deckUrls, err:= getDeckUrls()
	if err!=nil {
		fmt.Println("Failed to get deck urls")
		os.Exit(1)
	}


	count:= make(map[string]int)

	for i, aUrl:= range deckUrls{
		getDeckList(aUrl, count, aLogger)
		fmt.Println("Acquired ", aUrl, " ", i+1, " of ", len(deckUrls))
	}


	serialCount, err:= json.MarshalIndent(count, "", "    ")
	if err!=nil {
		fmt.Println("Failed to marshal count")
		aLogger.Println("Failed to marshal count")
	}

	ioutil.WriteFile(cacheLoc(), serialCount, 0666)

}

//the actual storage of cards.
//appearance is an integer out of 10k with 0 being 0% relative to the most
//prominent card and 50% being equated to 5k
//
//this provides two decimals of accuracy.
type cardData struct{
	Name string
	Appearance int
}

type cardDataCollection []cardData

func (someData cardDataCollection) Len() int {
	return len(someData)
}

// Swap is part of sort.Interface.
func (someData cardDataCollection) Swap(i, j int) {
	someData[i], someData[j] = someData[j], someData[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (someData cardDataCollection) Less(i, j int) bool {
	return someData[i].Appearance < someData[j].Appearance
}