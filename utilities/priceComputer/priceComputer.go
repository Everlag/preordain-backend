package main

import(

	"fmt"

	"io/ioutil"
	"encoding/json"

	"./../influxdbHandler"

)

// Where our influxdb client data is kept
const influxdbCredentials string = "influxdbCredentials.json"

func main() {

	aLogger:= GetLogger("priceWriter.log", "priceWriter")
	
	// Acquire the client we'll use to communicate with influxdb
	aClient, err:= getInfluxDBCLient()
	if err!=nil {
		aLogger.Fatalln("Failed to acquire client, ", err)
	}

	aLogger.Println("Clearing existing continuous queries")

	err = dropAllContinuousQueries(aClient)
	if err!=nil {
		aLogger.Fatalln(err)
	}

	aLogger.Println("Creating continuous queries")

	err = setupAllContinuousQueries(aClient)
	if err!=nil {
		aLogger.Fatalln(err)
	}

	aLogger.Println("Success")

}

func setupAllContinuousQueries(aClient *influxdbHandler.Client) error {

	// Acquire a list of series
	seriesList, err:= aClient.ListSeries()
	if err!=nil {
		return err
	}

	if len(seriesList)!=1 {
		return fmt.Errorf("Non-One amount of series points acquired")
	}

	fmt.Println("Creating continuous queries for ", len(seriesList[0].Points), " series")

	// Go through and scrape out the names for each series
	nameIndex:= seriesList[0].GetColumnIndex("name")
	names:= make([]string, len(seriesList[0].Points))

	for i, aPoint:= range seriesList[0].Points{

		names[i] = aPoint[nameIndex].(string)

	}

	// For each series, there may be multiple sets and sources.
	// We set up a continuous query for each of those.
	setsToSources:= make(map[string]string)
	for _, aSeriesName := range names{

		if aSeriesName == "Look at Me, I'm R&D" {
			continue
		}

		setsToSources = make(map[string]string)

		points, err:= aClient.SelectEntireSeries(aSeriesName)
		if err!=nil {
			fmt.Println("Failed to select entire series for '",
				aSeriesName, "'', ", err)
			continue
		}

		if len(points)!=1 {
			fmt.Println("Got a bad series in ", aSeriesName)
			continue
		}

		if len(points[0].Points)<1 {
			fmt.Println("Got a bad series in ", aSeriesName)
			continue
		}

		pointLength:= len(points[0].Points[0])

		sourceIndex:= points[0].GetColumnIndex("source")
		setIndex:= points[0].GetColumnIndex("set")
		// In the case of non-existent source or set columns,
		// this series is likely not something we want to mess with
		if sourceIndex == -1 || setIndex == -1 {
			continue
		}

		for _, aPoint:= range points[0].Points{
			if len(aPoint)!= pointLength {
				continue				
			}

			source:= aPoint[sourceIndex].(string)
			set:= aPoint[setIndex].(string)

			setsToSources[set] = source
		}

		// Now, for each source:set combination for this series,
		// we create nuke the query that may have existed before and create
		// a new continuous query
		for aSet, aSource:= range setsToSources{
			err = aClient.DropWeeksMedianContinuousQuerySeries(aSeriesName, aSet,
				aSource)
			if err!=nil {
				return fmt.Errorf("Failed to remove existing generated series",
					err)
			}

			err = aClient.AddWeeksMedianContinuousQuery(aSeriesName, aSet,
				aSource)
			if err!=nil {
				/*return fmt.Errorf("Failed to add query for '", aSeriesName,
					"', '", aSet, "', '", aSource, "', ", err)*/
				fmt.Println("Failed to add query for '", aSeriesName,
					"', '", aSet, "', '", aSource, "', ", err)
				
			}
		}

		fmt.Println(aSeriesName, " added successfully")

	}

	return nil

}

func dropAllContinuousQueries(aClient *influxdbHandler.Client) (error) {
	
	// Acquire all the queries for the database
	queryList, err:= aClient.ListContinuousQueries()
	if err!=nil {
		return err
	}

	// Sanity test to ensure we have >= 1 queries to delete
	if len(queryList)<=0 {
		return fmt.Errorf("No continuous queries to delete")
	}

	// We have a list of continuous queries now in influxdb point form,
	// find which column has the id number so we can nuke those queries
	idIndex:= queryList[0].GetColumnIndex("id")
	if idIndex == -1 {
		return fmt.Errorf("Failed to find id column")
	}


	// Grab the ids and issue delete requests for the queries they represent
	for _, aPoint:= range queryList[0].Points{

		// Continuous queries only have a single point
		anId:= int(aPoint[idIndex].(float64))

		err = aClient.DropContinuousQuery(int(anId))
		if err!=nil {
			return fmt.Errorf("Failed to drop Continuous query, id=", anId,
				" err=", err)
		}

	}

	return nil

}

func getInfluxDBCLient() (*influxdbHandler.Client, error) {

	creds, err:= getCredentials()
	if err!=nil {
		return &influxdbHandler.Client{}, err
	}

	aClient, err:= influxdbHandler.GetClient(creds.RemoteLocation, creds.DBName,
	 creds.User, creds.Pass,
	 creds.Read, creds.Write)
	if err!=nil {
		return &influxdbHandler.Client{}, err
	}

	return aClient, err
}

type credentials struct{
	RemoteLocation, DBName, User, Pass string

	Write, Read bool
}

func getCredentials() (credentials, error) {
	
	data, err:=ioutil.ReadFile(influxdbCredentials)
	if err!=nil {
		return credentials{}, fmt.Errorf("Failed to read influxdbCredentials, ", err)
	}

	var someCreds credentials
	err = json.Unmarshal(data, &someCreds)
	if err!=nil {
		return credentials{}, fmt.Errorf("Failed to unmarshal influxdbCredentials, ", err)
	}

	return someCreds, nil

}