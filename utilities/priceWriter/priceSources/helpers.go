package priceSources

import(

	"io/ioutil"
	"encoding/json"
	"strings"

	"io"
	"log"
	"os"
	"fmt"

)

const apiKeysLoc string = "apiKeys.json"
const setListLoc string = "setList.txt"

// The keys residing on disk alongside the traits we need to ensure we don't
// misuse them
type apiKeys struct{

	Mtgprice string
	MtgpriceLastUpdate int64
	MtgpriceWaitTime int64

}

// Acquires apikeys located at apiKeysLoc on disk
func getApiKeys() (apiKeys, error) {
	
	raw, err:= ioutil.ReadFile(apiKeysLoc)
	if err!=nil {
		return apiKeys{}, err
	}

	var keys apiKeys
	err = json.Unmarshal(raw, &keys)
	if err!=nil {
		return apiKeys{}, err
	}

	return keys, nil

}

// Updates the api keys status on disk. This is used when timestamps for each
// key are updated
func (keys *apiKeys) updateOnDisk() error {
	
	data, err:= json.Marshal(keys)
	if err!=nil {
		return err
	}

	err = ioutil.WriteFile(apiKeysLoc, data, 0666)
	if err!=nil {
		return err
	}

	return nil
}

func getSetList() ([]string, error) {

	sets, err:= ioutil.ReadFile(setListLoc)
	if err!=nil {
		return nil, err
	}

	return strings.Split(string(sets), "\n"), nil

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