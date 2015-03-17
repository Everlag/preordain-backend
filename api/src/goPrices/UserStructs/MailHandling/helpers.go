package MailHandling

import(

	"fmt"

	"io/ioutil"
	"encoding/json"

)

type credentials struct{

	SendingAddress string
	PrivateKey, PublicKey, Domain string

}

func getMailgunCredentials() (credentials, error) {
	
	data, err:=ioutil.ReadFile(credentialsLoc)
	if err!=nil {
		return credentials{},
		fmt.Errorf("Failed to read mailgun credentials, ", err)
	}

	var someCreds credentials
	err = json.Unmarshal(data, &someCreds)
	if err!=nil {
		return credentials{},
		fmt.Errorf("Failed to unmarshal mailgun credentials, ", err)
	}

	return someCreds, nil

}