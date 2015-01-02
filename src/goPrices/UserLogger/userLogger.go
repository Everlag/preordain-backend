package UserLogger

import(

	"fmt"

	"time"

	"encoding/json"

	"./../../../utilities/influxdbHandler"

)

const CredentialsTagType string = "UserActions"
const CredentialsLoc string = "influxdbCredentials.userLogging.json"

// Provides write-only logging to a remote database of user actions.
type Logger struct{

	client *influxdbHandler.Client

}

// Returns a new logger with sanity checks in place to prevent reading.
func NewLogger() (*Logger, error) {
	// Grab the credentials first
	creds, err:= getCredentials(CredentialsLoc)
	if err!=nil {
		return &Logger{}, err
	}

	if !creds.Write || creds.Read || creds.Tag!=CredentialsTagType {
		return &Logger{},
		fmt.Errorf("Creds have inappropriate permissions")
	}

	aClient, err:= influxdbHandler.GetClient(creds.RemoteLocation, creds.DBName,
	 creds.User, creds.Pass,
	 creds.Read, creds.Write)
	if err!=nil {
		return &Logger{},
		fmt.Errorf("Failed to acquire client, ", err)
	}

	aLogger:= Logger{
		client: aClient,
	}

	return &aLogger, nil

}

// Writes a log entry for the provided user with their sessionKey, action,
// actionParameters, and contents of their request, if any, recorded.
//
// Action is the REST path they used. Parameters for an action are those
// filling in the variables in the REST path.
//
// Body contents can be nil.
func (aLogger *Logger) WriteAction(userName, action string,
	actionParameters map[string]string, bodyContents interface{}) error {

	parameterBytes, err:= json.Marshal(actionParameters)
	if err!=nil {
		return err
	}


	var bodyContentsBytes []byte
	if bodyContents!=nil {
		bodyContentsBytes, err = json.Marshal(bodyContents)
		if err!=nil {
			return err
		}	
	}
	
	return aLogger.client.SendUserActionPoint(userName, action,
		string(parameterBytes), string(bodyContentsBytes), time.Now().UTC().Unix())

}