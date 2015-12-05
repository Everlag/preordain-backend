package main

import(

	"log"

	"time"

)

// YEAR-MONTH-DAY
const ReleaseDataFormat string = "2006-01-02"

type setMap map[string]*set

type set struct{

	// For free
	Name string
	ReleaseDate string
	Code string

	// Custom additions
	Timestamp int64

}

// Sets the epoch timestamp from set's string ReleaseDate
func (aSet *set) setTimestamp(aLogger *log.Logger) {
	
	timestamp, err:= time.Parse(ReleaseDataFormat, aSet.ReleaseDate)
	if err!=nil {
		aLogger.Println("Failed to convert time for ", aSet.Name, err)
		return
	}

	aSet.Timestamp = timestamp.UTC().Unix()

}