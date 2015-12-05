package main
//Hooks together the following datasources and outputs multiple directories
//containing everything needed for static mtg finance analysis except prices.
//
//Sources used are:
//
//mtgjson for inherent card info
//mtgsalvation for commander usage
//mtgimage for card images UNTESTED
//ngrams for card similarity

import(

	"log"
	"os"
	"fmt"
	"io"

	"github.com/joho/godotenv"

	"path/filepath"
)

// The various locations our derived data goes
const dataDir string = "cardText/"
const typeAheadDir string = "typeAhead/"

// The location and top count of the commander data we release
const topCommanderUsageLoc string = "commanderUsage"
const topCommanderUsageCount int = 1000

func main() {
	aLogger:= getLogger("core.log")

	// Populate config locations not explicitly set
	envError:= godotenv.Load("cardData.default.env")
	if envError!=nil {
		fmt.Println("failed to parse cardData.default.env")
		os.Exit(1)
	}

	// Notify intent to user
	fmt.Printf(`
Output directories:
	general:   %v
	typeAhead: %v
`, dataLoc(), typeAheadLoc())

	// Dumps data for each card into dataLoc 
	getAllCardData(aLogger)

	// Dumps typeAhead content into typeAheadLoc 
	getAllTypeAheadData(aLogger)

}

// Returns the complete path to our general output directory
func dataLoc() string {
	return filepath.Join(outputLoc(), dataDir)
}

// Returns the complete path to our typeahead output directory
func typeAheadLoc() string {
	return filepath.Join(outputLoc(), typeAheadDir)
}

// Returns the location of our general output directory
// as specified by the OUTPUT environment variable.
//
// An empty OUTPUT variable directs output to the working directory.
func outputLoc() string {

	// Fetch optionally specified cache location
	// root loc from environment
	loc:= os.Getenv("OUTPUT")
	if len(loc) == 0 {
		loc = "./"
	}

	return loc
}

// Acquires a logger a standardized logger.
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