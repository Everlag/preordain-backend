package main

import(

	"io/ioutil"
	"encoding/json"

	"time"


	"fmt"
	"io"
	"os"
	"log"

)

const StateLoc string = "deckWriter.state.json"

// Fetch the last time we performed a full fetch
//
// This is present next to the binary
func GetState() (time.Time, error) {
	
	d, err:= ioutil.ReadFile(StateLoc)
	if err!=nil {
		return time.Time{}, err
	}

	var last time.Time
	return last, json.Unmarshal(d, &last)

}

// Set the current time as when we last performed a full fetch
func SetState() error {
	now:= time.Now()

	d, err:= json.Marshal(now)
	if err!=nil {
		return err
	}

	return ioutil.WriteFile(StateLoc, d, 0666)
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