package main

import(
	
	"log"

	"./../../common/mtgjson"

)

type SetList []setMeta

type setMeta struct{

	Name, Code string

}

// Acquire a map[code]Name translation layer
func getSetCodeToSetNameTranslator(aLogger *log.Logger) map[string]string {

	poorTranslator, err:= mtgjson.AllSetsX()
	if err != nil {
		aLogger.Fatalf("Failed to unmarhsal AllSets-x ", err)
	}

	properTranslator:= make(map[string]string)

	for _, aSetMeta:= range poorTranslator{
		properTranslator[aSetMeta.Code] = aSetMeta.Name
	}

	return properTranslator
}

// Attach reserved list status to cards
func stapleOnSetSpecificData(aCardMap cardMap, aLogger *log.Logger) {

	// Base
	supplementary, err:= mtgjson.AllSetsX()
	if err != nil {
		aLogger.Fatalf("Failed to open supplementary set data, ", err)
	}

	// Check each card
	var set *mtgjson.Set
	var card mtgjson.Card
	for _, aCard:= range aCardMap{

		// Each printing has one valid
		firstPrintingCode:= aCard.Printings[0]
		set = supplementary[firstPrintingCode]

		// Find this card in the set
		for i := 0; i < len(set.Cards); i++ {
			card = set.Cards[i]

			if card.Name == aCard.Name {
				aCard.Reserved = card.Reserved
			}
		}

	}
}