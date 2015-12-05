package main

import(
	
	"log"

	"./../../common/mtgjson"

)

type SetList []setMeta

type setMeta struct{

	Name, Code string

}

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

//acquires the card data located only in our supplementary AllSets-x.json
func stapleOnSetSpecificData(aCardMap cardMap, aLogger *log.Logger) {

	//get the actual data
	supplementary, err:= mtgjson.AllSetsX()
	if err != nil {
		aLogger.Fatalf("Failed to open supplementary set data, ", err)
	}

	//go through and acquire the reserved list status of each card
	var set *mtgjson.Set
	var card mtgjson.Card
	for _, aCard:= range aCardMap{

		//each card has at least one printing we can use as an index
		firstPrintingCode:= aCard.Printings[0]
		set = supplementary[firstPrintingCode]

		//find the occurrence of this card
		for i := 0; i < len(set.Cards); i++ {
			card = set.Cards[i]

			if card.Name == aCard.Name {
				aCard.Reserved = card.Reserved
			}
		}

	}
}