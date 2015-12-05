package commanderData

import(

	"os"
	"fmt"
	"log"

	"strings"

	"path/filepath"

)

//Location of cache so we don't have to hit remote often
const cacheFile string = "commanderData.cache.json"

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

// Returns the location of the cache file
// as specified by the CACHE environment variable.
//
// An empty CACHE variable directs output to the working directory.
func cacheLoc() string {

	// Fetch optionally specified cache location
	// root loc from environment
	loc:= os.Getenv("CACHE")
	if len(loc) == 0 {
		loc = "./"
	}

	return filepath.Join(loc, cacheFile)
}

// Simple wrapper for allowing QueryableCommanderData the ability to sort cards
// based on their usage.
type cardItems []cardItem
type cardItem struct{
	Name string
	QueryableData *QueryableCommanderData
}

func (someData cardItems) Len() int {
	return len(someData)
}

// Swap is part of sort.Interface.
func (someData cardItems) Swap(i, j int) {
	someData[i], someData[j] = someData[j], someData[i]
}

// Less is part of sort.Interface.
// It is implemented by calling the "by" closure in the sorter.
func (someData cardItems) Less(i, j int) bool {
	// The only possible error is that of a card not found, whose priority
	// is set to zero
	valueI, _:= someData[i].QueryableData.Query(someData[i].Name)
	valueJ, _:= someData[j].QueryableData.Query(someData[j].Name)

	return valueI < valueJ
}