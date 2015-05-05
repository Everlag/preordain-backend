package priceReader

import(

	"fmt"

	"encoding/json"
	"io/ioutil"

)


type credentials struct{
	RemoteLocation, DBName, User, Pass string

	Tag string

	Write, Read bool
}

func getCredentials(credentialsLoc string) (credentials, error) {
	
	data, err:=ioutil.ReadFile(credentialsLoc)
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