package main

import(

	"fmt"
	"log"
	"os"
	"io"
	"strings"
	"strconv"

	"encoding/json"
	"io/ioutil"

	"html/template"
	"path/filepath"

)


const twitterRenderLoc string = "twitterRenders/"

const metaName string = "meta.json"
const logName string = "renderLog.txt"

func main() {

	aLogger:= getLogger(logName)

	metaData:= getMeta(aLogger)
	aLogger.Println(metaData)

	renderTwitter(metaData, aLogger)
}

func renderTwitter(someMeta meta, aLogger *log.Logger) {

	// Acquire our template which we shall be pouring into
	mold, err:= getTemplate(someMeta.TwitterCardTemplate)
	if err!=nil {
		aLogger.Fatalf("Failed to acquire template, ", err)
	}

	// Acquire the list of elements we shall deal with
	datums, err:= getDatumList(someMeta.LocalCardTextLoc)
	if err!=nil {
		aLogger.Fatalf("Failed to acquire list of datums, ", err)
	}

	var aCard card
	var someContent *PageContent
	var target string
	for _, aDatum := range datums{

		// Grab the card from disk
		aCard, err =  getCard(aDatum)
		if err!=nil {
			aLogger.Println("Failed to acquire ", aDatum)
		}

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
			}
		}

	}

}

type meta struct{

	// Data for populating twitter fields
	Creator, Site, Domain string

	// Where card text is kept in the local filesystem
	LocalCardTextLoc string

	// Where our various templates our kept in the local
	TwitterCardTemplate string

	// To be able to associate a page and image with a card
	RemoteImageLoc, RemoteImageExtension string

	// The path our client uses for cards
	RemoteCardLoc string

	// The path in which we store our renders.
	// NOTE: must match structurally with the production server
	LocalCardRenders string
}

// Returns the metadata we have set.
//
// If failure occurs, simply Fatalfs while logging cause
func getMeta(aLogger *log.Logger) (meta) {
	
	metaData, err:= ioutil.ReadFile(metaName)
	if err!=nil {
		aLogger.Fatalf("Failed to read", metaName)
	}

	var metaParsed meta
	err = json.Unmarshal(metaData, &metaParsed)
	if err!=nil {
		aLogger.Fatalf("Failed to unmarshal ", metaName, ", ", err)
	}

	return metaParsed

}

// Takes a template location and returns a ready to run template
// assuming no errors
func getTemplate(loc string) (*template.Template, error) {
	
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

type PageContent struct{
	Creator, Domain, Title, Image string
	FirstHeader, FirstData, SecondHeader, SecondData string
	Url string
	Description, MetaItems string
	SimilarItems []SimilarDatums
}

type SimilarDatums struct{
	Name, Link string
}

func getDatumList(loc string) ([]string, error) {
	
	files, err:= ioutil.ReadDir(loc)
	if err!=nil {
		return nil, err
	}

	datumList:= make([]string, 0)
	datumLoc:= ""

	for _, aFile:= range files{

		if !aFile.IsDir() {
			datumLoc = filepath.Join(loc, aFile.Name())
			datumList = append(datumList, datumLoc)	
		}

	}

	return datumList, nil

}

// The components of the card we can expose
type card struct{
	//What we get for free, or near free, from mtgjson
	Name string
	Text string
	ManaCost string
	Colors []string
	Power, Toughness, Type, ImageName string
	Printings, Types, SuperTypes, SubTypes []string
	Legalities map[string]string
	Reserved bool
	Loyalty int

	//Extensions we add manually
	CommanderUsage float64
	SimilarCards []string
	SimilarCardConfidences []float64
	Categories []string
}

// Attempt to grab a card from disk given its location
func getCard(loc string) (card, error) {
	
	cardRaw, err:= ioutil.ReadFile(loc)
	if err!=nil {
		return card{}, err
	}

	var cardData card
	err = json.Unmarshal(cardRaw, &cardData)
	if err!=nil {
		return card{}, err
	}

	return cardData, nil

}

// A commander fringe playable appears in at least 1% of sampled decks
const commanderFringeUsage float64  = 1.0

// A commander staple appears in at least 10% of sampled decks
const commanderStapleUsage float64 = 10.0

// Fills a PageContent as well as possible with the provided metadata
func fillContent(someMeta *meta, aCard *card, printing string) *PageContent {
	title:= strings.Join([]string{aCard.Name, "|", printing}, " ")

	metaItems:= make([]string, 0)
	if aCard.Reserved {
		metaItems = append(metaItems, "Reserved List")
	}

    if (aCard.CommanderUsage * 100) >= commanderStapleUsage{
		metaItems = append(metaItems, "Commander Staple")
    }else if (aCard.CommanderUsage * 100) >= commanderFringeUsage{
		metaItems = append(metaItems, "Playable")
    }

    similarCardDatums:= make([]SimilarDatums, len(aCard.SimilarCards))
    for i, similar:= range aCard.SimilarCards{
    	similarCardDatums[i] = SimilarDatums{
    		Name: similar,
    		Link: getCanonicalLink(someMeta, similar, "somePrinting"),
    	}
    }

	freshContent:= PageContent{
		// Meta
		Creator: someMeta.Creator,
		Domain: someMeta.Domain,
		Url: getCanonicalLink(someMeta, aCard.Name, printing),
		Image: strings.Join([]string{someMeta.RemoteImageLoc, aCard.ImageName + someMeta.RemoteImageExtension}, "/"),

		// Card
		Title: title,
		Description: aCard.Text,
		SimilarItems: similarCardDatums,
		MetaItems: strings.Join(metaItems, ", "),

		FirstHeader: "Text",
		FirstData: aCard.Text,
		SecondHeader: "Commander Play",
		SecondData: strconv.FormatFloat(aCard.CommanderUsage * 100, byte('f'), -1, 64) + "%",
	}

	return &freshContent
}
/* TO FINISH
type PageContent struct{
	FirstHeader, FirstData, SecondHeader, SecondData string
}
*/

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