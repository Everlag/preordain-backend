package commanderData

import(

	"os"
	"fmt"
	"log"

	"strings"

)

func getLogger(fName, name string) (aLogger *log.Logger) {
	file, err:= os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err!=nil {
		fmt.Println("Starting logger failed, I have no mouth but must scream!")
		fmt.Println(err)
		os.Exit(0)
	}
	defer file.Close()

	aLogger = log.New(file, name + " ", log.Ldate|log.Ltime|log.Lshortfile)

	return
}

//normalize card names to a standard form which should ignore most trivial typing errors
func normalizeCardName(cardName string) string {
	properName:= strings.ToLower(cardName)

	return properName 
}