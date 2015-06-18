package deckData

type ArchetypeStats struct{
	// A map[cardName]appearance rate
	// with the fact that they simply appear in a list
	//
	// Exludes sideboard
	Appearances map[string]float64

	// The appearances map but strictly for sideboard cards 
	SideboardAppearances map[string]float64

	// How many copies a deck plays of a card divided across its sideboard
	// and mainboard averaged across the archetype
	AverageCopies map[string]float64

}

// Returns basic statistics for the deck if the card is present.
//
// If not present, returns false as the second result.
//
// Use the map val, ok access method for a key possibly not present
func (stats *ArchetypeStats) QueryCard(name string) (DeckResult, bool) {

	result:= DeckResult{}

	average, ok:= stats.AverageCopies[name]
	if !ok {
		return result, false
	}
	
	result.Average = average

	sideboard, ok:= stats.SideboardAppearances[name]
	if !ok {
		result.Sideboard = false
		return result, true
	}
	mainboard, ok:= stats.Appearances[name]
	if !ok {
		result.Sideboard = true
		return result, true
	}

	if mainboard < sideboard {
		result.Sideboard = true
		return result, true
	}

	return result, true

}

// Digests the results of a run of analyzeArchetype into an ArchetypeStats
func digestStats(listCount float64, totalCount,
	mainboardExistence, sideboardExistence map[string]int) ArchetypeStats {
	
	stats:= ArchetypeStats{}
	stats.Appearances = make(map[string]float64)
	stats.SideboardAppearances = make(map[string]float64)
	stats.AverageCopies = make(map[string]float64)

	for name, count:= range totalCount{
		average:= float64(count) / listCount
		// Cap the average
		if average > 4.0 {
			average = 4.0
		}

		mainboardAppearances:= float64(mainboardExistence[name]) / listCount
		sideboardAppearances:= float64(sideboardExistence[name]) / listCount

		stats.Appearances[name] = mainboardAppearances
		stats.SideboardAppearances[name] = sideboardAppearances
		stats.AverageCopies[name] = average
	}

	return stats

}

// Analyzes an archetype for all traits found in ArchetypeStats
func analyzeArchetype(decks []mwDeck) ArchetypeStats {

	// A chunk of decks we can mess with
	malleable:= make([]mwDeck, len(decks))
	copy(malleable, decks)

	// How many total decks we deal with, this is important
	totalDecks:= len(decks)
	
	// The binary fact that they appear in the deck
	mainboardExistence:= make(map[string]int)
	sideboardExistence:= make(map[string]int)

	// The sum of all copies of the card in the deck
	totalCopies:= make(map[string]int)

	mainboard:= 0
	sideboard:= 0
	name:= ""
	for _, deck:= range decks{

		for _, card:= range deck{

			name = card.Name

			mainboard = deck.Mainboard(name)
			sideboard = deck.Sideboard(name)

			totalCopies[name]+= mainboard + sideboard
			if mainboard > 0 {
				mainboardExistence[name]++	
			}
			if sideboard > 0 {
				sideboardExistence[name]++	
			}

		}

	}

	return digestStats(float64(totalDecks), totalCopies,
		mainboardExistence, sideboardExistence)

}