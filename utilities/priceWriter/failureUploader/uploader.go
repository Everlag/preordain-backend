package main
// Implements a basic uploader for price data that has been stored
// upon failure to upload immediately following collection.
//
// Usage is:
//	failureUploader -source 1419967891mtgprice -creds influxdbCredentials.json

import(

	"fmt"

	"flag"

	"io/ioutil"
	"encoding/json"

	"./../../influxdbHandler"
	"./../priceSources"



)


func main() {
	aLogger:= priceSources.GetLogger("failureUploader.log", "failureUploader")

	// Read in the flags
	var uploadSource string
	flag.StringVar(&uploadSource, "source",
		"1419967891mtgprice", "The location of the failed upload")
	var influxdbCredentials string
	flag.StringVar(&influxdbCredentials, "creds",
		"influxdbCredentials.json", "The location of the credentials")
	flag.Parse()

	// Acquire the client
	creds, err:= getCredentials(influxdbCredentials)
	if err!=nil {
		aLogger.Fatalln(err)
	}

	aClient, err:= influxdbHandler.GetClient(creds.RemoteLocation, creds.DBName,
	 creds.User, creds.Pass,
	 creds.Read, creds.Write)
	if err!=nil {
		aLogger.Fatalln("Failed to ping remote server at client creation, ", err)
	}

	aLogger.Println("Acquire client, acquiring upload")

	upload, err:= getFailedUpload(uploadSource)
	if err!=nil {
		aLogger.Fatalln("Failed to acquire upload, ",
			err)	
	}

	aLogger.Println("Acquired upload, starting upload")

	err = uploadSingleSourceResults(upload, aClient)
	if err!=nil {
		aLogger.Fatalln("Failed to upload prices, ",
			err)
	}

	aLogger.Println("Upload Successful, feel free to delete failed data.")

}

func getFailedUpload(uploadSource string) (priceSources.PriceMap, error) {

	data, err:=ioutil.ReadFile(uploadSource)
	if err!=nil {
		return priceSources.PriceMap{},
		fmt.Errorf("Failed to read failed upload, ", err)
	}


	var upload priceSources.PriceMap
	err = json.Unmarshal(data, &upload)
	if err!=nil {
		return priceSources.PriceMap{},
		fmt.Errorf("Failed to unmarshal failed upload, ", err)
	}

	return upload, nil
}

func uploadSingleSourceResults(aPriceResult priceSources.PriceMap,
	aClient *influxdbHandler.Client) error {

	// Construct the points to send
	points:= make([]influxdbHandler.Point, 0)

	for aSetName, cardMap:= range aPriceResult.Prices{

		for aCardName, aPrice:= range cardMap{

			// Deal with the fact that some price sources may have multiple
			// currencies that were massaged into USD
			var aPoint influxdbHandler.Point
			if aPriceResult.HasEuro {
				// An original price in euros is recorded alongside the USD
				// conversion
				euroPrice:= aPriceResult.EURPrices[aSetName][aCardName]
				
				aPoint = influxdbHandler.BuildPointWithEuro(aCardName,
					aPriceResult.Time, aPrice, euroPrice,
					aSetName, aPriceResult.Source)
			
			}else{
			
				aPoint = influxdbHandler.BuildPoint(aCardName,
					aPriceResult.Time, aPrice, aSetName, aPriceResult.Source)
			
			}

			points = append(points, aPoint)

		}
	}

	// Send the points to the db
	err:= aClient.SendPoints(points)

	return err

}

type credentials struct{
	RemoteLocation, DBName, User, Pass string

	Write, Read bool
}

func getCredentials(influxdbCredentials string) (credentials, error) {
	
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