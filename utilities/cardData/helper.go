package main

import(

	"strings"

	"io/ioutil"

)

const setListLoc string = "setList.txt"

func getSupportedSetListFlat() ([]string, error) {

	// Get the raw, newline delimited, list from disk
	sets, err:= ioutil.ReadFile(setListLoc)
	if err!=nil {
		return nil, err
	}

	// Clean it up to be an array
	setList:= strings.Split(string(sets), "\n")

	return setList, nil

}

// Returns a map of set names located in setList.txt to their foil equivalent.
// 
// If no equivalent exists, a blank string is in place
func getSupportedSetList() (map[string]string, error) {

	// Clean it up to be an array
	setList, err:= getSupportedSetListFlat()
	if err!=nil {
		return nil, err
	}

	// A place to map from setName:foilName
	setMap:= make(map[string]string)

	for _, aSet:= range setList{

		// Check if this is a foil set
		if strings.Contains(aSet, " Foil"){

			// Add it to the map if it is
			normalName:= strings.Replace(aSet, " Foil", "", -1)
			setMap[normalName] = aSet

		}else{
			// A non foil set requires us to check if the foil version is
			// already present

			_, ok:= setMap[aSet]
			if ok {
				// Already present
				continue
			}

			// Add it in with a blank foil name
			setMap[aSet] = ""

		}

	}

	return setMap, nil

}