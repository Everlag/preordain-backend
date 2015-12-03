package priceSources

import(

	"io/ioutil"
	"encoding/json"

	"io"
	"log"
	"os"
	"fmt"

	"path/filepath"
)

const apiKeysName string = "apiKeys.json"

// The keys residing on disk alongside the traits we need to ensure we don't
// misuse them
type ApiKeys struct{

	Mtgprice string
	MtgpriceLastUpdate int64
	MtgpriceWaitTime int64

	MKMConsumerKey string
	MKMSecretKey string
	MKMLastUpdate int64
	MKMPriceWaitTime int64

	OpenexchangeratesKey string

}

// Acquires apikeys located at apiKeysLoc on disk
func getApiKeys() (ApiKeys, error) {

	apiKeysLoc:= getApiKeysLoc()

	raw, err:= ioutil.ReadFile(apiKeysLoc)
	if err!=nil {
		return ApiKeys{}, err
	}

	var keys ApiKeys
	err = json.Unmarshal(raw, &keys)
	if err!=nil {
		return ApiKeys{}, err
	}

	return keys, nil

}

// Updates the api keys status on disk. This is used when timestamps for each
// key are updated
func (keys *ApiKeys) updateOnDisk() error {
	
	data, err:= json.Marshal(keys)
	if err!=nil {
		return err
	}

	apiKeysLoc:= getApiKeysLoc()

	err = ioutil.WriteFile(apiKeysLoc, data, 0666)
	if err!=nil {
		return err
	}

	return nil
}

// Intelligently fetches location of api keys
func getApiKeysLoc() string {
	// Fetch optionally specified keys
	// root loc from environment
	loc:= os.Getenv("APIKEYS")
	if len(loc) == 0 {
		loc = "./"
	}

	return filepath.Join(loc, apiKeysName)
}

func GetLogger(fName, name string) (aLogger *log.Logger) {
	file, err:= os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err!=nil {
		fmt.Println("Starting logger failed, cannot write to logger to say logger failed. Oh god.")
		fmt.Println(err)
		os.Exit(0)
	}

	multi:= io.MultiWriter(file, os.Stdout)

	aLogger = log.New(multi, name, log.Ldate|log.Ltime|log.Lshortfile)

	return
}