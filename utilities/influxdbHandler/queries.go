package influxdbHandler

import(

	"net/http"
	"net/url"

	"fmt"
	"io/ioutil"

	"strings"
	"strconv"

)

const ListSeriesQuery string = "list series"
const ListContinuousQuery string = "list continuous queries;"
const SelectEntireSeriesTemplate string = "select * from \"seriesName\""
const DropSeriesQueryTemplate string = "drop series \"seriesName\""
const DropContinuousQueryTemplate string = "drop continuous query queryID"
const AddWeeksMedianTemplate string = 
"select MEDIAN(price) from \"seriesName\" where set='setName' and source='sourceName' group by time(7d) into containerSeriesName"
const DropWeeksMedianTemplate string = 
"drop series \"seriesName\""

func (aClient *Client) buildQuery(query string)(*http.Request, error){

	properQuery:= url.QueryEscape(query)

	// derive path with static settings
	fullPath:= aClient.dbLoc + aClient.dataPostPath + 
	"?time_precision=s" + "&q=" + properQuery
	
	// put it all together
	req, err:= http.NewRequest("GET", fullPath, nil)
	if err!=nil {
		return nil, err
	}

	// stick authentication onto the request
	req.SetBasicAuth(aClient.userName, aClient.password)

	return req, nil

}

func (aClient *Client) sendQuery(req *http.Request) ([]byte, error) {
	
	response, err:= aClient.httpClient.Do(req)
	if err!=nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Request not explicit success")
	}

	// copy the request body into a byte array which is less than efficient
	// but frees up the httpclient to do other work
	responseBody, err:= ioutil.ReadAll(response.Body)
	if err!=nil {
		return nil, fmt.Errorf("Failed to read request body")
	}

	return responseBody, nil

}

// Build and sends the provided query.
func (aClient *Client) executeQuery(someQueryString string) ([]byte, error){

	aQuery, err:= aClient.buildQuery(someQueryString)
	if err!=nil {
		return nil, err
	}

	seriesListBytes, err:= aClient.sendQuery(aQuery)

	return seriesListBytes, err

}

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
	
	fullQueryText:= strings.Replace(SelectEntireSeriesTemplate, "seriesName",
		seriesName, -1)

	seriesBytes, err:= aClient.executeQuery(fullQueryText)
	if err!=nil {
		return Points{}, err
	}

	return pointsFromBytes(seriesBytes)

}

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

	targetSeriesName:= GetMedianWeeksSeriesName(cardName, setName, sourceName)
	setName = NormalizeName(setName)

	fullQueryText:= strings.Replace(AddWeeksMedianTemplate,
		"seriesName", cardName, -1)
	fullQueryText = strings.Replace(fullQueryText, "setName", setName, -1)
	fullQueryText = strings.Replace(fullQueryText, "sourceName", sourceName, -1)
	fullQueryText = strings.Replace(fullQueryText, "containerSeriesName",
		targetSeriesName, -1)

	_, err:= aClient.executeQuery(fullQueryText)
	
	//fmt.Println(fullQueryText)

	return err

}

func (aClient *Client) DropWeeksMedianContinuousQuery(cardName,
	setName, sourceName string) error {
	
	if !aClient.write {
		return fmt.Errorf("Client does not have write permissions")
	}

	targetSeriesName:= GetMedianWeeksSeriesName(cardName, setName, sourceName)

	fullQueryText:= strings.Replace(DropWeeksMedianTemplate,
		"seriesName", targetSeriesName, -1)

	_, err:= aClient.executeQuery(fullQueryText)
	
	return err
}