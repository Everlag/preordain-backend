package databaseHandling

import (
	"io"
	"os"

	"log"
	"fmt"

	"encoding/json"
)

//implements a simple and reliable copy of files
//src = https://gist.github.com/elazarl/5507969
func cp(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

func getLogger(fName, name string) (aLogger *log.Logger) {
	file, err:= os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err!=nil {
		fmt.Println("Starting logger failed, I have no mouth but must scream!")
		fmt.Println(err)
		os.Exit(0)
	}

	multi:= io.MultiWriter(file, os.Stdout)

	aLogger = log.New(multi, name + " ", log.Ldate|log.Ltime|log.Lshortfile)

	return
}

//converts the set data to json for disk purposes
func (someStorage *WrappedStorage) ToJson() ([]byte, error) {
	marshalledData, err := json.Marshal(someStorage)
	if err != nil {
		fmt.Println("Failed to marhsal storage data")
		return nil, fmt.Errorf("Failed to marhsal storage data")
	}

	return marshalledData, nil
}