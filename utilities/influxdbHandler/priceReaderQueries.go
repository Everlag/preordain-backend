package influxdbHandler

import(

	"fmt"

	"strings"

)

const ListSeriesQuery string = "list series"
const ListContinuousQuery string = "list continuous queries;"

const SelectEntireSeriesTemplate string = "select * from \"seriesName\""
const SelectWeeksMedianTemplate string =
"select median from \"seriesName\" where time > timeStart"
const SelectFilteredSeriesTemplate string = 
"select price from \"seriesName\" where time > timeStart and set='setName' and source='sourceName'"
const SelectFilteredSeriesLatestTemplate string = 
"select price from \"seriesName\" where time > timeStart and set='setName' and source='sourceName' limit 1"

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

	timeStartString:= TimestampToInfluxDBTime(timeStart)
	setName = aClient.NormalizeName(setName)

	fmt.Println(setName)

	fullQueryText:= strings.Replace(SelectFilteredSeriesLatestTemplate, "seriesName",
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

// Returns the WeeksMedian for a card with a given set and source without
// unmarshalling the result.
//
// Takes a time as an int64 for the cutoff
func (aClient *Client) SelectWeeksMedian(cardName,
	setName, sourceName string, timeStart int64) (Points, error) {

	if !aClient.read {
		return nil, fmt.Errorf("Client does not have read permissions")
	}

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