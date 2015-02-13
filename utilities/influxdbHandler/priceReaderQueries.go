package influxdbHandler

import(

	"fmt"

	"strings"

)

// The queries we use
const ListSeriesQuery string = "list series;"
const ListContinuousQuery string = "list continuous queries;"

const SelectEntireSeriesTemplate string = "select * from \"seriesName\";"
const SelectWeeksMedianTemplate string =
"select median from \"seriesName\" where time > timeStart;"
const SelectFilteredSeriesTemplate string = 
"select price from \"seriesName\" where time > timeStart and set='setName' and source='sourceName';"
const SelectFilteredSeriesLatestTemplate string = 
"select price from \"seriesName\" where time > timeStart and set='setName' and source='sourceName' limit 1;"

// How many cards we can string together in one query for completing a full set
// query
//
// This takes into account the url limit of an effective 2048 characters.
const CardsPerSetBatch int = 17

func (aClient *Client) ListSeries() (Points, error) {

	if !aClient.read {
		return Points{}, fmt.Errorf("Client does not have read permissions")
	}
	
	seriesListBytes, err:= aClient.executeQuery(ListSeriesQuery)
	if err!=nil {
		return Points{}, err
	}

	seriesListPoints, err:= pointsFromBytes(seriesListBytes)
	if err!=nil {
		return Points{}, err
	}

	return seriesListPoints, nil

}


func (aClient *Client) ListContinuousQueries() (Points, error) {

	if !aClient.read {
		return Points{}, fmt.Errorf("Client does not have read permissions")
	}
	
	queryListBytes, err:= aClient.executeQuery(ListContinuousQuery)
	if err!=nil {
		return Points{}, err
	}

	queryListPoints, err:= pointsFromBytes(queryListBytes)
	if err!=nil {
		return Points{}, err
	}

	return queryListPoints, nil

}

func (aClient *Client) SelectEntireSeries(seriesName string) (Points, error) {
	
	if !aClient.read {
		return Points{}, fmt.Errorf("Client does not have read permissions")
	}

	fullQueryText:= strings.Replace(SelectEntireSeriesTemplate, "seriesName",
		seriesName, -1)

	seriesBytes, err:= aClient.executeQuery(fullQueryText)
	if err!=nil {
		return Points{}, err
	}

	return pointsFromBytes(seriesBytes)

}

// Returns the entire series with the available filters of source,
// set, and time applied to the points
//
// Takes a time as an int64 for the cutoff
func (aClient *Client) SelectFilteredSeries(cardName,
	setName, sourceName string, timeStart int64) (Points, error) {
	
	if !aClient.read {
		return nil, fmt.Errorf("Client does not have read permissions")
	}

	timeStartString:= TimestampToInfluxDBTime(timeStart)
	setName = aClient.NormalizeName(setName)

	fmt.Println(setName)

	fullQueryText:= strings.Replace(SelectFilteredSeriesTemplate, "seriesName",
		cardName, -1)
	fullQueryText = strings.Replace(fullQueryText, "setName",
		setName, -1)
	fullQueryText = strings.Replace(fullQueryText, "sourceName",
		sourceName, -1)
	fullQueryText = strings.Replace(fullQueryText, "timeStart",
		timeStartString, -1)

	seriesBytes, err:= aClient.executeQuery(fullQueryText)
	if err!=nil {
		return Points{}, err
	}

	return pointsFromBytes(seriesBytes)

}

// Returns the latest point of the series with the available filters of source,
// set, and time applied to the points
//
// Takes a time as an int64 for the cutoff
func (aClient *Client) SelectFilteredSeriesLatestPoint(cardName,
	setName, sourceName string, timeStart int64) (Points, error) {
	
	if !aClient.read {
		return nil, fmt.Errorf("Client does not have read permissions")
	}

	fullQueryText:= aClient.buildSelectFilteredSeriesLatestPoint(cardName,
		setName, sourceName, timeStart)

	seriesBytes, err:= aClient.executeQuery(fullQueryText)
	if err!=nil {
		return Points{}, err
	}

	return pointsFromBytes(seriesBytes)

}

// Builds the query for SelectFilteredSeriesLatestPoint.
//
// Exposing the query building allows multiple exposed query methods to
// take advantage of it.
func (aClient *Client) buildSelectFilteredSeriesLatestPoint(cardName,
	setName, sourceName string, timeStart int64) string {

	timeStartString:= TimestampToInfluxDBTime(timeStart)
	setName = aClient.NormalizeName(setName)

	var fullQueryText string

	fullQueryText = strings.Replace(SelectFilteredSeriesLatestTemplate,
		"seriesName",
		cardName, -1)
	fullQueryText = strings.Replace(fullQueryText, "setName",
		setName, -1)
	fullQueryText = strings.Replace(fullQueryText, "sourceName",
		sourceName, -1)
	fullQueryText = strings.Replace(fullQueryText, "timeStart",
		timeStartString, -1)

	return fullQueryText

}

// Selects the prices for an entire set. Uses reasonable batching to enchance
// performance and reduce the cost of round trip latency
func (aClient *Client) SelectSetsLatest(cardList []string,
	setName, sourceName string, timeStart int64) (Points, error) {
	
	
	acquiredPoints:= make(Points, 0)

	for _, aCardName:= range cardList {
		fullQueryText:= aClient.buildSelectFilteredSeriesLatestPoint(aCardName,
		setName, sourceName, timeStart)

		seriesBytes, err:= aClient.executeQuery(fullQueryText)
		if err!=nil {
			return Points{}, err
		}

		batchPoints, err:= pointsFromBytes(seriesBytes)
		if err!=nil {
			return Points{},
			fmt.Errorf("Failed to unmarshal points for set", err)
		}

		acquiredPoints = append(acquiredPoints, batchPoints...)

	}

	return acquiredPoints, nil

}



// Returns the WeeksMedian for a card with a given set and source without
// unmarshalling the result.
//
// Takes a time as an int64 for the cutoff
func (aClient *Client) SelectWeeksMedian(cardName,
	setName, sourceName string, timeStart int64) (Points, error) {

	if !aClient.read {
		return nil, fmt.Errorf("Client does not have read permissions")
	}

	setName = aClient.NormalizeName(setName)

	fmt.Println(setName)

	targetSeriesName:= aClient.GetMedianWeeksSeriesName(cardName,
		setName, sourceName)
	timeStartString:= TimestampToInfluxDBTime(timeStart)

	fullQueryText:= strings.Replace(SelectWeeksMedianTemplate, "seriesName",
		targetSeriesName, -1)
	fullQueryText = strings.Replace(fullQueryText, "timeStart",
		timeStartString, -1)

	seriesBytes, err:= aClient.executeQuery(fullQueryText)
	if err!=nil {
		return Points{}, err
	}

	return pointsFromBytes(seriesBytes)
}