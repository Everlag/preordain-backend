package categoryBuilder

import(

	"os"
	"fmt"
	"log"
	"io"

	"strings"

)

func getLogger(fName, name string) (aLogger *log.Logger) {
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

//characters that we remove in card text to normalize data
var badCharacters = [...]string{".", ",", "\"", "{", "}", "[", "]"}
var numbers = [...]string{"1","2","3","4","5","6","7","8","9","0", " N", " X"}

func cleanCardText(someText, name string) string {
		//we need to clean the text of the string for hard brackets
	cleanedText:= someText

	//first, switch the name out for a token
	cleanedText = strings.Replace(cleanedText, name, "~", -1)

	hintOpenerIndex := strings.Index(cleanedText, "(")
	hintCloserIndex := strings.Index(cleanedText, ")")

	for hintOpenerIndex!=-1 &&
		hintCloserIndex!=-1{

		cleanedText = strings.Replace(cleanedText,
			cleanedText[hintOpenerIndex:hintCloserIndex+1], "", 1)

		hintOpenerIndex = strings.Index(cleanedText, "(")
		hintCloserIndex = strings.Index(cleanedText, ")")

	}

	manaOpenerIndex := strings.Index(cleanedText, "{")
	manaCloserIndex := strings.Index(cleanedText, "}")

	for manaOpenerIndex!=-1 &&
		manaCloserIndex!=-1{

		cleanedText = strings.Replace(cleanedText,
			cleanedText[manaOpenerIndex:manaCloserIndex+1], " mtgSymbol ", 1)

		manaOpenerIndex = strings.Index(cleanedText, "{")
		manaCloserIndex = strings.Index(cleanedText, "}")

	}

	cleanedText = strings.ToLower(cleanedText)

	for _, aBadChar := range badCharacters{
		cleanedText = strings.Replace(cleanedText, aBadChar, "", -1)
	}
	for _, aNumber := range badCharacters{
		cleanedText = strings.Replace(cleanedText, aNumber, "aNumber", -1)
	}

	cleanedText = strings.TrimSpace(cleanedText)

	return cleanedText
}


//normalize card names to a standard form which should ignore most trivial typing errors
func normalizeCardName(cardName string) string {
	properName:= strings.ToLower(cardName)

	return properName 
}