package influxdbHandler

import(

	"fmt"

	"strings"
	"strconv"

)

const DropSeriesQueryTemplate string = "drop series \"seriesName\""
const DropContinuousQueryTemplate string = "drop continuous query queryID"
const AddWeeksMedianTemplate string = 
"select MEDIAN(price) from \"seriesName\" where set='setName' and source='sourceName' group by time(7d) into containerSeriesName"
const DropWeeksMedianTemplate string = 
"drop series \"seriesName\""

// Drops a series with name of seriesName. The series name is enclosed with quotes
// so support for a variety of series is available.
//
// Returns an error upon an operational failure
func (aClient *Client) DropSeries(seriesName string) (error) {
	
	if !aClient.write {
		return fmt.Errorf("Client does not have write permissions")
	}

	// Build the full query out of the template replacing the proper name
	fullQueryText:= strings.Replace(DropSeriesQueryTemplate,
		"seriesName", seriesName, -1)

	_, err:= aClient.executeQuery(fullQueryText)
	
	return err

}

func (aClient *Client) DropContinuousQuery(id int) (error) {
	
	if !aClient.write {
		return fmt.Errorf("Client does not have write permissions")
	}

	queryID:= strconv.Itoa(id)

	fullQueryText:= strings.Replace(DropContinuousQueryTemplate, "queryID",
		queryID, -1)

	_, err:= aClient.executeQuery(fullQueryText)

	return err

}

func (aClient *Client) AddWeeksMedianContinuousQuery(cardName,
	setName, sourceName string) error {
	
	if !aClient.write {
		return fmt.Errorf("Client does not have write permissions")
	}

	targetSeriesName:= aClient.GetMedianWeeksSeriesName(cardName,
		setName, sourceName)
	setName = aClient.NormalizeName(setName)

	fullQueryText:= strings.Replace(AddWeeksMedianTemplate,
		"seriesName", cardName, -1)
	fullQueryText = strings.Replace(fullQueryText, "setName", setName, -1)
	fullQueryText = strings.Replace(fullQueryText, "sourceName", sourceName, -1)
	fullQueryText = strings.Replace(fullQueryText, "containerSeriesName",
		targetSeriesName, -1)

	_, err:= aClient.executeQuery(fullQueryText)

	return err

}

func (aClient *Client) DropWeeksMedianContinuousQuerySeries(cardName,
	setName, sourceName string) error {
	
	if !aClient.write {
		return fmt.Errorf("Client does not have write permissions")
	}

	targetSeriesName:= aClient.GetMedianWeeksSeriesName(cardName,
		setName, sourceName)

	fullQueryText:= strings.Replace(DropWeeksMedianTemplate,
		"seriesName", targetSeriesName, -1)

	_, err:= aClient.executeQuery(fullQueryText)
	
	return err
}