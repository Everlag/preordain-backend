package main

import(

	"fmt"
	"log"
	"os"
	"io"
	"strings"

)

func getCanonicalLink(someMeta *meta, cardName string, printing string) string {
	return strings.Join([]string{someMeta.RemoteCardLoc,
		cardName, printing}, "/")
}

func getLogger(fName string) (aLogger *log.Logger) {
	file, err:= os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err!=nil {
		fmt.Println("Starting logger failed, cannot write to logger to say logger failed. Uh oh.")
		fmt.Println(err)
		os.Exit(0)
	}

	multi:= io.MultiWriter(file, os.Stdout)

	aLogger = log.New(multi, "User ", log.Ldate|log.Ltime|log.Lshortfile)

	return
}