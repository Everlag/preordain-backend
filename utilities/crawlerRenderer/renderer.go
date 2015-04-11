package main

import(

	"fmt"
	"log"
	"os"

	"html/template"
	"path/filepath"

)


const twitterRenderLoc string = "twitterRenders/"

const index string = "index"

const metaName string = "meta.json"
const logName string = "renderLog.txt"

func main() {

	aLogger:= getLogger(logName)

	someMeta:= getMeta(aLogger)
	aLogger.Println(someMeta)

	urls:= renderTwitter(someMeta, aLogger)
	renderSiteMap(someMeta, urls, aLogger)
}

// Renders the template specified out to the location specified in meta.
//
// Returns a list of urls created.
func renderTwitter(someMeta meta, aLogger *log.Logger) []string {

	// Acquire our template which we shall be pouring into
	mold, err:= getHTMLTemplate(someMeta.TwitterCardTemplate)
	if err!=nil {
		aLogger.Fatalf("Failed to acquire template, ", err)
	}

	// Acquire the list of elements we shall deal with
	datums, err:= getDatumList(someMeta.LocalCardTextLoc)
	if err!=nil {
		aLogger.Fatalf("Failed to acquire list of datums, ", err)
	}

	urls:= make([]string, 0)

	var aCard card
	var someContent *PageContent
	var target string
	for _, aDatum := range datums{

		// Grab the card from disk
		aCard, err =  getCard(aDatum)
		if err!=nil {
			aLogger.Println("Failed to acquire ", aDatum)
		}

		// Apply an effectively blank printing to the card to ensure sanity.
		aCard.Printings = append(aCard.Printings, index)

		// Send the card to disk for each printing it's had
		for _, aPrinting:= range aCard.Printings{
			target = filepath.Join(someMeta.LocalCardRenders, aCard.Name, aPrinting)

			// Create all necessary directories
			basepath:= filepath.Dir(target)
			err = os.MkdirAll(basepath, 0777)
			if err!=nil {
				aLogger.Println("Failed to create directory, ", aCard.Name, aPrinting)
			}

			// Fill up a template form
			someContent = fillContent(&someMeta, &aCard, aPrinting)

			// Execute the template straight to disk
			err = fillTemplate(mold, someContent, target)
			if err!=nil {
				aLogger.Println("Failed to fill template, ", err)
				continue
			}

			// Add a successfully generated templated url to our page list
			urls = append(urls, someContent.Url)
		}

	}

	return urls
}

// Takes a template location and returns a ready to run template
// assuming no errors
func getHTMLTemplate(loc string) (*template.Template, error) {
	
	return template.ParseFiles(loc)

}

// Attempts to fill a given template with the provided content and deposit
// that inside the target file
func fillTemplate(mold *template.Template,
	content *PageContent,
	target string) (error) {
	
	dump, err:= os.Create(target)
	if err!=nil {
		return fmt.Errorf("Failed to open target, ", err)
	}

	return mold.Execute(dump, *content)

}