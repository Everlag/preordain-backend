package priceReader

import(

	"fmt"

	"./../../../utilities/influxdbHandler"

)

// We can optionally tag a client for a specific library to prevent common
// human errors when dealing with api keys
const CredentialsTagType string = "Prices"

const CredentialsLoc string = "influxdbCredentials.priceReader.json"

// Some useful time constants for performing queries
const Day int64 = 3600 * 24
const Month int64 = 30 * Day // A month, in our eyes, has 30 days
const TwoMonths int64 = 2 * Month 
const ThreeMonths int64 = 3 * Month
const Year int64 = 12 * Month
const Decade int64 = 10 * Year

// Returns an influxdb client capable of reading price data from the remote
// server. Credentials must be located in influxdbCredentials.json beside the
// binary.
//
// Performs sanity tests to ensure that we can read but that we can't write
// using the provided client
func AcquireReader() (*influxdbHandler.Client, error) {
	
	// Grab the credentials first
	creds, err:= getCredentials(CredentialsLoc)
	if err!=nil {
		return &influxdbHandler.Client{}, err
	}

	if !creds.Read || creds.Write || creds.Tag!=CredentialsTagType {
		return &influxdbHandler.Client{},
		fmt.Errorf("Creds have inappropriate permissions")
	}

	aClient, err:= influxdbHandler.GetClient(creds.RemoteLocation, creds.DBName,
	 creds.User, creds.Pass,
	 creds.Read, creds.Write)
	if err!=nil {
		return &influxdbHandler.Client{},
		fmt.Errorf("Failed to acquire client, ", err)
	}

	return aClient, nil

}