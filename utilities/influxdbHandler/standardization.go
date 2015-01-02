package influxdbHandler

import(

	"strings"
	"bytes"

	"strconv"

)

// Normalizes a provided name into something not awful to query for.
//
// This is a more efficient version tied to the state of a client in order
// to be able to support a replacer object.
func (aClient *Client) NormalizeName(aName string) string {
	
	return aClient.replacer.Replace(aName)

}

// Returns a queryable series name for a median(week's prices) of a given card.
//
// Resultant form is 'cardNameWithoutSpaces.sourceName.setName.WeeksMedian'
func (aClient *Client) GetMedianWeeksSeriesName(cardName, setName,
	sourceName string) string {
	
	setName = aClient.NormalizeName(setName)
	cardName = aClient.NormalizeName(cardName)

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
//
// Inefficient but usable without a client.
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

func TimestampToInfluxDBTime(time int64) string {
	return strconv.FormatInt(time, 10) + "s"
}