package ApiServices

import(

	"encoding/json"
	"io/ioutil"

)

// A global map used to ensure we only search for valid sets
var sets = make(map[string]bool)
// To ensure we only search for and allow into trades cards
// that are actual magic cards. These can be cards that aren't in the list of
// valid sets but we whitelist input to be within the domain of all cards
// ever printed
var cards = make(map[string]bool)

var cardsToSets = make(map[string]map[string]bool)

// Populates the setToCardMap and the cards map
func populateCardMaps() error {
	
	cardErr, setErr:= populateCardsMap(), populateSets()
	if cardErr!=nil {
		return cardErr
	}
	if setErr!=nil {
		return setErr
	}

	return nil

}

func populateSets() error {

	// Acquire the list of valid sets we'll deal with
	setList, err:= getSetList()
	if err!=nil {
		return err
	}

	for _, aSet:= range setList{
		sets[aSet] = true
	}

	return nil

}

func populateCardsMap() error {

	// Acquire the map of card names
	cardData, err:= ioutil.ReadFile("AllCards-x.json")
	if err!=nil {
		return err
	}

	var aCardList cardMap
	err = json.Unmarshal(cardData, &aCardList)
	if err!=nil {
		return err
	}

	for aCardName, aCard:= range aCardList{
		cards[aCardName] = true

		cardsToSets[aCardName] = make(map[string]bool)
		for _, aSet:= range aCard.Printings{
			cardsToSets[aCardName][aSet] = true
		}

	}

	return nil

}

type cardMap map[string]card

type card struct{
	Name string
	Printings []string
}