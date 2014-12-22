package influxdbHandler
// A specialized package for interfacing with influxdb for our use case
// All actions are performed over the http api using BasicAuth

import(

	"net/http"
	"bytes"
	"encoding/json"

	"io/ioutil"

	"fmt"

	"strings"

)

var Columns = []string{"time", "price", "set", "source"}
var ColumnCount = len(Columns)

type Client struct{

	dbLoc, dbName string

	dataPostPath string

	userName, password string

	// Permissions are managed on the influxdb level but we might as well
	// save some processing time by respecting them here
	read, write bool

	httpClient *http.Client

}

func GetClient(dbLoc, dbName, userName, password string,
	canRead, canWrite bool) *Client {

	httpClient:= &http.Client{}

	dataPostPath:= "/db/" + dbName + "/series"

	aClient:= Client{
		dbLoc: dbLoc,
		dbName: dbName,
		dataPostPath: dataPostPath,
		userName: userName,
		password: password,
		read: canRead,
		write: canWrite,
		httpClient: httpClient,
	}

	return &aClient
	
}

func (aClient *Client) buildRequest(path, method string,
	payload []byte) (*http.Request, error) {

	// derive path with static settings
	fullPath:= aClient.dbLoc + path + "?time_precision=s"

	// setup the payload if necessary

	packagedPayload:= bytes.NewReader(payload)	
	
	// put it all together
	req, err:= http.NewRequest(method, fullPath, packagedPayload)
	if err!=nil {
		return nil, err
	}

	// stick authentication onto the request
	req.SetBasicAuth(aClient.userName, aClient.password)

	return req, nil

}

func (aClient *Client) sendRequest(req *http.Request) ([]byte, error) {
	
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

// Pings the influxdb endpoint to ensure it is alive
func (aClient *Client) Ping() error {
	req, err:= aClient.buildRequest("/ping", "GET", nil)
	if err!=nil {
		return fmt.Errorf("Failed to build request")
	}

	_, err = aClient.sendRequest(req)
	if err!=nil {
		return err
	}

	// We ping to see if they live, that's all
	return nil

}

// Sends the provided set of points to the db the client works with
func (aClient *Client) SendPoints(somePoints Points) error {
	
	if !aClient.write {
		return fmt.Errorf("Client does not have write permissions")
	}

	data, err:= json.Marshal(somePoints)
	if err!=nil {
		return fmt.Errorf("Failed to marshal provided points")
	}

	req, err:= aClient.buildRequest(aClient.dataPostPath, "POST", data)
	if err!=nil {
		return err
	}

	_, err = aClient.sendRequest(req)
	if err!=nil {
		return err
	}

	return nil

}


type Points []Point

type Point struct{

	Name string `json:"name"`
	Columns []string `json:"columns"`
	Points [][]PointData `json:"points"`
}

// We need to mix ints and strings so this gets us to where we need to be
type PointData interface{}

func BuildPoint(seriesName string, time int64, price int64,
	set, source string) Point {
	
	cleanedSetName:= normalizeSetName(set)

	data:= make([]PointData, ColumnCount)
	data[0] = PointData(time)
	data[1] = PointData(price)
	data[2] = PointData(cleanedSetName)
	data[3] = PointData(source)

	wrappedData:= make([][]PointData, 1)
	wrappedData[0] = data

	aPoint:= Point{
		Name: seriesName,
		Columns: []string(Columns),
		Points: wrappedData,
	}

	return aPoint

}

func BuildPointMultiplePrices(seriesName string, times []int64, prices []int64,
	set, source string) Point {
	
	cleanedSetName:= normalizeSetName(set)

	wrappedData:= make([][]PointData, len(times))

	var price int64
	for i, time:= range times{
		price = prices[i]

		data:= make([]PointData, ColumnCount)
		data[0] = PointData(time)
		data[1] = PointData(price)
		data[2] = PointData(cleanedSetName)
		data[3] = PointData(source)

		wrappedData[i] = data

	}

	aPoint:= Point{
		Name: seriesName,
		Columns: []string(Columns),
		Points: wrappedData,
	}

	return aPoint

}

// Normalize a provided set name into something not awful to query influxdb
// for.
func normalizeSetName(aName string) string {
	
	aName = strings.Replace(aName, "'", "", -1)
	aName = strings.Replace(aName, "\"", "", -1)
	aName = strings.Replace(aName, "-", "", -1)
	aName = strings.Replace(aName, ":", "", -1)

	return aName

}