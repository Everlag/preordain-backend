package main

import(

	"log"

	"encoding/json"
	"io/ioutil"
	"os"

	"time"

	"strings"

)

// YEAR-MONTH-DAY
const ReleaseDataFormat string = "2006-01-02"

const TimestampMapLoc string = "timestamps"

type setMap map[string]*set

type set struct{

	Name string
	ReleaseDate string
	Type string
	Code string
	Booster interface{}


	Timestamp int64

}

// Converts release data to a timestamp and sets that as a int64
func (aSet *set) setTimestamp(aLogger *log.Logger) {
	
	timestamp, err:= time.Parse(ReleaseDataFormat, aSet.ReleaseDate)
	if err!=nil {
		aLogger.Println("Failed to convert time for ", aSet.Name, err)
		return
	}

	aSet.Timestamp = timestamp.UTC().Unix()

}

// dumpToDisk commits each value of the card map and dumps it into
// the dataLoc folder under the name.json file
func (setData *setMap) dumpToDisk(aLogger *log.Logger) {
	
	aLogger.Println("Commencing dump to disk of setMap")

	var serialSet []byte
	var err error

	var setPath string

	for _, aSet:= range *setData {

		// Skip mtgo only sets
		if aSet.Type == "masters" {
			continue
		}

		serialSet, err= json.Marshal(aSet)
		if err!=nil {
			aLogger.Println("Failed to marshal ",  aSet.Name)	
			continue
		}

		// Make the set name not explode for invalid characters on some
		// file systems
		cleanedSetName:= strings.Replace(aSet.Name, ":", "", -1)
		cleanedSetName = strings.Replace(cleanedSetName, "–", "", -1)
		cleanedSetName = strings.Replace(cleanedSetName, "\"", "", -1)

		setPath = dataLoc() + string(os.PathSeparator) + cleanedSetName + ".json"

		err:= ioutil.WriteFile(setPath, serialSet, 0666)
		if err!=nil {
			aLogger.Println("Failed to save set ", cleanedSetName, err)
		}else{
			aLogger.Println("Saved ", cleanedSetName)
		}

	}

	aLogger.Println("Dump complete")

}