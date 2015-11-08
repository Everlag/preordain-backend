package ApiServices

import(

	"./../../../common/mtgjson"

	"strings"
	"sort"

)

// A global map used to ensure we only search for valid sets
var sets = make(map[string]bool)

// To ensure we only search for and allow into trades cards
// that are actual magic cards. These can be cards that aren't in the list of
// valid sets but we whitelist input to be within the domain of all cards
// ever printed
var cards = make(map[string]bool)

var cardsToSets = make(map[string]map[string]bool)

// This is specifically geared towards being capable of providing per-set
// data
var setsToCardsAndRarity = make(SetsToCards)

// Populates the setToCardMap and the cards map
//
// Pass a influxdbClient
func populateCardMaps() error {
	
	var setErr, cardErr, cardRarityErr error
	sets, setErr = populateSets()
	cards, cardsToSets, cardErr = populateCardsTranslationMap(sets)
	setsToCardsAndRarity, cardRarityErr = populateCardsRarityMap(sets)
	if cardErr!=nil {
		return cardErr
	}
	if setErr!=nil {
		return setErr
	}
	if cardRarityErr!=nil {
		return cardRarityErr
	}

	return nil

}

func populateSets() (map[string]bool, error) {

	sets:= make(map[string]bool)

	// Acquire the list of valid sets we'll deal with
	setList, err:= getSetList()
	if err!=nil {
		return sets, err
	}

	// Adds names of sets we use
	for _, aSet:= range setList{
		sets[aSet] = true
	}

	return sets, nil

}

func populateCardsTranslationMap(validSets map[string]bool) (map[string]bool,
	map[string]map[string]bool, error) {

	cards:= make(map[string]bool)
	cardsToSets:= make(map[string]map[string]bool)

	// Acquire the map of card names
	aCardList, err:= mtgjson.AllCardsX()
	if err!=nil {
		return nil, nil, err
	}

	// Also acquire the sets we support and sort them so
	// we can easily search
	setList, err:= getSetList()
	if err!=nil {
		return cards, cardsToSets, err
	}

	sort.Strings(setList)

	for aCardName, aCard:= range aCardList{
		cards[aCardName] = true

		cardsToSets[aCardName] = make(map[string]bool)
		for _, aPrinting:= range aCard.Printings{
			
			_, ok:= validSets[aPrinting]
			if ok{
				cardsToSets[aCardName][aPrinting] = true

				// Make a check to add the foil as well if such
				// a printing exists
				foilCandidate:= aPrinting + " Foil"
				if sort.SearchStrings(setList, foilCandidate)!=-1{
					cardsToSets[aCardName][foilCandidate] = true
				}
			}

		}

	}

	return cards, cardsToSets, nil

}

type cardMap map[string]card

type card struct{
	Name string
	Printings []string
}

func populateCardsRarityMap(validSets map[string]bool) (SetsToCards, error) {
	
	aSetToCardsMap:= make(SetsToCards)

	setMap, err:= mtgjson.AllSetsX()
	if err!=nil {
		return nil, err
	}

	for _, aSet:= range setMap{
		_, ok:= validSets[aSet.Name]
		if !ok {
			continue
		}

		for _, aCard:= range aSet.Cards{
			aSetToCardsMap.addCardToSet(aSet.Name, setSpecficCard{
				Name: aCard.Name,
				Rarity: aCard.Rarity,
				})
		}
	}

	return aSetToCardsMap, nil

}

type setMap map[string]set
type set struct{
	Name string
	Cards []setSpecficCard
}

// A wrapped map to make life ever so easier.
type SetsToCards map[string][]setSpecficCard
type setSpecficCard struct{
	Name, Rarity string
}

// Returns a list of cards contained within that set.
//
// An empty list is returned if provided an invalid set name
func (aSetToCardsMap SetsToCards) getCardName(aSet string) []string {

	cards, ok:= aSetToCardsMap[aSet]
	if !ok {
		return make([]string, 0)
	}

	cardNames:= make([]string, len(cards))
	for i, aCard:= range cards{
		cardNames[i] = aCard.Name
	}

	return cardNames
}

// Returns a list of cards contained within that set with the provided rarity
//
// An empty list is returned if provided an invalid set name
func (aSetToCardsMap SetsToCards) getCardsWithRarity(aSet,
	rarity string) []string {

	aSet = strings.Replace(aSet, " Foil", "", -1)

	cards, ok:= aSetToCardsMap[aSet]
	if !ok {
		return make([]string, 0)
	}

	cardNames:= make([]string, 0)
	for _, aCard:= range cards{
		if aCard.Rarity == rarity{
			cardNames = append(cardNames, aCard.Name)
		}
	}

	return cardNames
}

func (aSetToCardsMap SetsToCards) addCardToSet(aSet string,
	aCard setSpecficCard) {

	_, ok:= aSetToCardsMap[aSet]
	if !ok {
		aSetToCardsMap[aSet] = make([]setSpecficCard, 0)
	}

	aSetToCardsMap[aSet] = append(aSetToCardsMap[aSet], aCard)

}