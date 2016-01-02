package main

import (
	"fmt"

	gq "github.com/PuerkitoBio/goquery"

	"./../../common/deckDB/deckData"

	"strings"

	"time"

	"net/url"

	"regexp"

	"golang.org/x/text/encoding/charmap"

)

// Match dd/mm/yy
var timePattern = regexp.MustCompile(`\b[0-9]{2}\b\/\b[0-9]{2}\b\/\b[0-9]{2}\b`)

// Fetches a specific event and returns an equivalent Event
// holding everything we could want to know about it
//
// TODO: check db if each event or deck is already present
//       to avoid refetching them unnecessarily
func FetchEvent(id string) (*deckData.Event, error) {

	loc := fmt.Sprintf("http://mtgtop8.com/event?e=%s", id)

	// Actual event
	doc, err:= gq.NewDocument(loc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event", err)
	}

	// Event name
	name, err:= eventName(doc)
	if err!=nil {
		return nil, fmt.Errorf("failed to read title", err)
	}

	happened, err:= eventTime(doc)
	if err!=nil {
		return nil, fmt.Errorf("failed to read time", err)
	}

	// Deck ids for easier fetching
	ids, err:= eventDecks(doc)
	if err!=nil {
		return nil, fmt.Errorf("failed to read decks", err)
	}


	// Fetch each decklist
	decks:= make([]*deckData.Deck, 0)

	for _, dID:= range ids{
		d, err:= FetchDeck(dID)
		if err!=nil {
			continue
		}

		d.DeckID = dID

		decks = append(decks, d)
	}


	// More funky encoding shenanigans!
	decoder1215:= charmap.Windows1252.NewDecoder()

	name, err = decoder1215.String(name)
	if err!=nil {
		return nil, err
	}

	return &deckData.Event{
		Name: name, EventID: id,
		Happened: deckData.Timestamp(happened),
		Decks: decks}, nil
}

// Finds name of an event given its document
func eventName(doc *gq.Document) (string, error) {

	sel:= doc.Find(".w_title").Find("tr")
	header:= sel.Text()
	
	// Nothing or no space means we messed up
	if len(header) == 0 ||
	strings.Index(header, "\n") == -1 {
		return "", fmt.Errorf("invalid or nonexistent name header")
	}
	title:= strings.Split(header, "\n")[0]

	return strings.TrimSpace(title), nil
}

// Finds time a specific event happened given its document
//
// Time is formatted dd/mm/yy and sits in the document
// as 'Format [-] TIME\nOPTIONAL'
//
// Optional appears only on some events, they're a hassle
func eventTime(doc *gq.Document) (time.Time, error) {

	sel:= doc.Find("td.S14")
	header:= strings.TrimSpace(sel.Text())

	// Find a string matching our date format in the header.
	//
	// Much easier with a regexp
	match := timePattern.FindStringSubmatch(header)
	if len(match) == 0 {
		return time.Time{}, fmt.Errorf("invalid or nonexistent date header")
	}

	return time.Parse("02/01/06", match[0])
}

// Finds all decks in a specific event
//
// Returns ids and corrresponding mtgtop8 deck names
func eventDecks(doc *gq.Document) ([]string, error) {

	ids:= make([]string, 0)

	sel:= selectEventAnchors(doc)

	sel.Each(func(i int, a *gq.Selection){
		
		// No need to check, already filtered
		rel, _:= a.Attr("href")

		// Parse query values
		u, err:= url.Parse(rel)
		if err !=nil {
			return
		}

		id:= u.Query().Get("d")
		ids = append(ids, id)
	})

	// Optional decks can fail but are nice to have
	optIds, err:= optionalDecks(doc)
	if err == nil {
		ids = append(ids, optIds...)
	}

	return ids, nil
}

// Finds all decks which appear for large but not small events
//
// Also returns deck names
func optionalDecks(doc *gq.Document) ([]string, error) {

	ids:= make([]string, 0)

	sel:= doc.Find("[name=sel_deck]").Find("optgroup").Find("option")
	if sel.Length() == 0 {
		return nil, fmt.Errorf("no optional decks")
	}

	sel.Each(func(i int, o *gq.Selection){
		// No need to check, already filtered
		id, ok:= o.Attr("value")
		if !ok {
			return
		}

		ids = append(ids, id)
	})

	return ids, nil
}