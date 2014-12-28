package influxdbHandler

import(

	"strings"
	"bytes"

)

// Returns a queryable series name for a median(week's prices) of a given card.
//
// Resultant form is 'cardNameWithoutSpaces.sourceName.setName.WeeksMedian'
func GetMedianWeeksSeriesName(cardName, setName,
	sourceName string) string {
	
	setName = NormalizeName(setName)
	cardName = NormalizeName(cardName)

	var seriesNameBytes bytes.Buffer
	seriesNameBytes.WriteString(cardName)
	seriesNameBytes.WriteString(".")
	seriesNameBytes.WriteString(sourceName)
	seriesNameBytes.WriteString(".")
	seriesNameBytes.WriteString(setName)
	seriesNameBytes.WriteString(".WeeksMedian")

	return seriesNameBytes.String()

}

// Normalize a provided name into something not awful to query influxdb for.
func NormalizeName(aName string) string {
	
	aName = strings.Replace(aName, "'", "", -1)
	aName = strings.Replace(aName, "\"", "", -1)
	aName = strings.Replace(aName, "-", "", -1)
	aName = strings.Replace(aName, ":", "", -1)
	aName = strings.Replace(aName, " ", "", -1)
	aName = strings.Replace(aName, "(", "", -1)
	aName = strings.Replace(aName, ")", "", -1)
	aName = strings.Replace(aName, ",", "", -1)
	aName = strings.Replace(aName, "!", "", -1)
	aName = strings.Replace(aName, "?", "", -1)
	aName = strings.Replace(aName, "/", "", -1)
	aName = strings.Replace(aName, "&", "", -1)

	return aName

}