package main

import(

	"fmt"
	"log"

	"encoding/json"
	"io/ioutil"
	"os"

	"time"

)

// YEAR-MONTH-DAY
const ReleaseDataFormat string = "2006-01-02"

const TimestampMapLoc string = "timestamps"

func getAllSetData(aLogger *log.Logger) {

	setData, err:= buildSetData()
	if err!=nil {
		aLogger.Fatalln("Failed to acquire basic set data ", err)
	}

	// Create timestamps for set data and get a map of name to data
	timeStampMap:= setData.setTimestamps(aLogger)

	dumpTimeStamps(timeStampMap, aLogger)

	setData.dumpToDisk(aLogger)

}

func dumpTimeStamps(timeStampMap map[string]int64, aLogger *log.Logger) {
	// Save that map
	serialTimestamps, err:= json.Marshal(timeStampMap)
	if err!=nil {
		aLogger.Fatalln("Failed to marshal timestamps, ", err)
	}

	setPath:= dataLoc + string(os.PathSeparator) + TimestampMapLoc + ".json"

	err = ioutil.WriteFile(setPath, serialTimestamps, 0666)
	if err!=nil {
		aLogger.Fatalln("Failed to save timestamps, ", err)
	}
}

func buildSetData() (setMap, error) {
	
	setData, err:= ioutil.ReadFile("AllSets-x.json")
	if err!=nil {
		return setMap{}, fmt.Errorf("Failed to read AllSets-x.json, ", err)
	}

	//unmarshal it into a map of string to card with image name
	var aSetMap setMap
	err = json.Unmarshal(setData, &aSetMap)
	if err!=nil {
		return setMap{}, fmt.Errorf("Failed to unmarshal set map, ", err)
	}

	return aSetMap, nil

}


type setMap map[string]*set

// Interprets and sets the unix timestamp for every set in the map
//
// Returns a map of set names to their timestamp
func (setData *setMap) setTimestamps(aLogger *log.Logger) map[string]int64 {
	
	timestampMap:= make(map[string]int64)

	for _, aSet:= range *setData{

		aSet.setTimestamp(aLogger)

		timestampMap[aSet.Name] = aSet.Timestamp

	}

	return timestampMap

}

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

		serialSet, err= json.Marshal(aSet)
		if err!=nil {
			aLogger.Println("Failed to marshal ",  aSet.Name)	
			continue
		}

		setPath = dataLoc + string(os.PathSeparator) + aSet.Name + ".json"

		ioutil.WriteFile(setPath, serialSet, 0666)

	}

	aLogger.Println("Dump complete")

}