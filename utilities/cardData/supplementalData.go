package main

import(
	
	"log"
	"encoding/json"
	"io/ioutil"

)

type CompleteData map[string]SetData

type SetData struct {
	Name  string
	Cards []card
}


type SetList []setMeta

type setMeta struct{

	Name, Code string

}

func getSetNameToSetCodeTranslator(aLogger *log.Logger) map[string]string {
	data, err := ioutil.ReadFile("SetList.json")
	if err != nil {
		aLogger.Fatalf("Failed to open SetList.json, ", err)
	}

	var poorTranslator SetList

	err = json.Unmarshal(data, &poorTranslator)
	if err != nil {
		aLogger.Fatalf("Failed to unmarhsal SetList.json, ", err)
	}

	properTranslator:= make(map[string]string)

	for _, aSetMeta:= range poorTranslator{
		properTranslator[aSetMeta.Name] = aSetMeta.Code
	}

	return properTranslator
}

func getSetCodeToSetNameTranslator(aLogger *log.Logger) map[string]string {

	poorTranslator, err:= buildSetData()
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
	//thanks to AllSets using codes instead of names as keys, we get to
	//work around that!
	setNameTranslator:= getSetNameToSetCodeTranslator(aLogger)

	//get the actual data
	data, err := ioutil.ReadFile("AllSets-x.json")
	if err != nil {
		aLogger.Fatalf("Failed to open supplementary set data, ", err)
	}

	var supplementary CompleteData

	err = json.Unmarshal(data, &supplementary)
	if err != nil {
		aLogger.Fatalf("Failed to unmarhsal supplementary set data, ", err)
	}

	//go through and acquire the reserved list status of each card
	var someSetData SetData
	for _, aCard:= range aCardMap{

		//each card has at least one printing we can use as an index
		firstPrintingCode:= setNameTranslator[aCard.Printings[0]]
		someSetData = supplementary[firstPrintingCode]

		//find the occurrence of this card
		var someCard card
		for i := 0; i < len(someSetData.Cards); i++ {
			someCard = someSetData.Cards[i]

			if someCard.Name == aCard.Name {
				aCard.Reserved = someCard.Reserved
			}
		}

	}
}