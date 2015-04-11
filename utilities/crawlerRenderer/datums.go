package main

import(

	"log"

	"strings"
	"strconv"

	"encoding/json"
	"io/ioutil"
	
	"path/filepath"

)

// A commander fringe playable appears in at least 1% of sampled decks
const commanderFringeUsage float64  = 1.0

// A commander staple appears in at least 10% of sampled decks
const commanderStapleUsage float64 = 10.0

type PageContent struct{
	Creator, Domain, Title, Image string
	FirstHeader, FirstData, SecondHeader, SecondData, TwitterDescription string
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

// Fills a PageContent as well as possible with the provided metadata
func fillContent(someMeta *meta, aCard *card, printing string) *PageContent {
	
	var title string
	if printing == index {
		// If this is the special case index printing for sanity then
		// we just grab the first real printing available.
		title = strings.Join([]string{aCard.Name, "|", aCard.Printings[0]}, " ")
	}else{
		title = strings.Join([]string{aCard.Name, "|", printing}, " ")
	}

	metaItems:= make([]string, 0)
	if aCard.Reserved {
		metaItems = append(metaItems, "Reserved List")
	}

    if (aCard.CommanderUsage * 100) >= commanderStapleUsage{
		metaItems = append(metaItems, "Commander Staple")
    }else if (aCard.CommanderUsage * 100) >= commanderFringeUsage{
		metaItems = append(metaItems, "Commander Playable")
    }

    similarCardDatums:= make([]SimilarDatums, len(aCard.SimilarCards))
    for i, similar:= range aCard.SimilarCards{
    	similarCardDatums[i] = SimilarDatums{
    		Name: similar,
    		Link: getCanonicalLink(someMeta, similar, ""),
    	}
    }

	freshContent:= PageContent{
		// Meta
		Creator: someMeta.Creator,
		Domain: someMeta.Domain,
		Url: getCanonicalLink(someMeta, aCard.Name, printing),
		Image: strings.Join([]string{someMeta.RemoteImageLoc, aCard.ImageName + someMeta.RemoteImageExtension}, "/"),
		TwitterDescription: someMeta.TwitterDescription,

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

type meta struct{

	// Data for populating twitter fields
	Creator, Site, Domain, TwitterDescription string

	// Where card text is kept in the local filesystem
	LocalCardTextLoc string

	// Where our various templates our kept in the local
	TwitterCardTemplate, SiteMapTemplate string

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