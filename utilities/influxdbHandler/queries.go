package influxdbHandler

import(

	"net/http"
	"net/url"

	"fmt"

	"io/ioutil"
)

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

	// Copy the request body into a byte array which is less than efficient
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