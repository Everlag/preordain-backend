package deckData

import(

	"fmt"

	"encoding/json"
	"io/ioutil"

)

const baseLocation string = "http://mtgtop8.com/"
const archeTypeLocation string = baseLocation + "format?f=MO"

const archeTypePrefix string = "archetype"
const eventPrefix string = "event"
const specificDeckPrefix string = "d="
const deckPrefix string = ".mwDeck"

const generalLinkClass string = ".hover_tr"
const deckListClass string = ".Nav_link"

type CardResult struct{
	Decks []DeckResult
	// What percentage of archetypes have this card as at least
	// a one of.
	Popularity float64

}

type DeckResult struct{
	// The archetype
	Name string
	// How many copies appear on average in this archetype
	Average float64
	// If the card is more prevalent in sideboard or mainboard
	Sideboard bool
}

func (usableData *QueryableDeckData) QueryCard(name string) CardResult {
	
	results:= CardResult{}
	results.Decks = make([]DeckResult, 0)

	appeared:= 0.0
	total:= float64(len(usableData.PerArchetype))
	for archetype, stats:= range usableData.PerArchetype{
		deckResult, ok:= stats.QueryCard(name)
		if !ok {
			continue
		}

		deckResult.Name = archetype
		results.Decks = append(results.Decks, deckResult)

		appeared++
	}


	results.Popularity = appeared / total

	return results

}

type QueryableDeckData struct{

	// Statistics for mtg archetypes
	PerArchetype map[string]ArchetypeStats

}

func GetQueryableDeckData() (QueryableDeckData, error) {
	usableData:= QueryableDeckData{}
	err:= usableData.populate()

	return usableData, err
}

// Attempts to read some json encoded QueryableDeckData
// from local cache at cacheFile
func queryableDeckDataFromCache() (QueryableDeckData, error) {

	var usableData QueryableDeckData
	
	raw, err:= ioutil.ReadFile(cacheLoc())
	if err!=nil {
		return usableData, err
	}

	return usableData, json.Unmarshal(raw, &usableData)

}

func (usableData *QueryableDeckData) toDisk() error {
	raw, err:= json.Marshal(usableData)
	if err!=nil {
		return err
	}

	return ioutil.WriteFile(cacheLoc(), raw, 0777)
}

func (usableData *QueryableDeckData) populate() error {
	
	// Try to obtain from the cache
	data, err:= queryableDeckDataFromCache()
	if err!=nil {
		fmt.Println("Failed to acquire deck cache, acquiring remotely")
		return usableData.populateFromRemote()
	}

	fmt.Println("Acquired local deck cache")

	*usableData = data

	return nil

}

func (usableData *QueryableDeckData) populateFromRemote() error {
	
	fmt.Println("Getting archetype list")
	archetypes, names, err:= usableData.gatherArchetypes()
	if err!=nil {
		return err
	}

	// Keep all of our decklist links in one place, their location
	// maps to their name anyway!
	allLists:= make(map[string]string)

	totalArchetypes:= len(archetypes)
	viewedArchetypes:= 0

	for i, archetype:= range archetypes{

		// At this point we transform from mtgtop8 names to our names
		translatedName, err:= Translate(names[i])
		if err!=nil {
			fmt.Println("skipping", archetype, err)
			continue
		}

		decklists, err:= gatherArcheTypeMap(names[i], archetype)
		if err!=nil{
			return err
		}

		// Copy these lists to all of our lists
		for k, _:= range decklists{
			allLists[k] = translatedName
		}

		// Report progress
		viewedArchetypes++
		fmt.Println(viewedArchetypes, "/", totalArchetypes,
			"archetypes have deck list locations")

	}

	// Get the decklists
	archetypeToDecks:= make(map[string][]mwDeck)
	totalLists:= len(allLists)
	viewedLists:= 0
	for list, archetype:= range allLists{

		deck, err:= getDeck(baseLocation + list)
		if err!=nil {
			return err
		}

		decks, ok:= archetypeToDecks[archetype]
		if !ok {
			decks = make([]mwDeck, 0)
		}

		decks = append(decks, deck)
		archetypeToDecks[archetype] = decks

		viewedLists++
		fmt.Println(viewedLists, "/", totalLists,
			"lists acquired")

	}

	fmt.Println("Digesting Statistics")
	usableData.PerArchetype = make(map[string]ArchetypeStats)
	for archetype, decks:= range archetypeToDecks{
		usableData.PerArchetype[archetype] = analyzeArchetype(decks)


		fmt.Println(archetype, "\n",
			usableData.PerArchetype[archetype].AverageCopies)
	}

	// Save what we have obtained
	return usableData.toDisk()
	
}