package main

import (
	"fmt"

	gq "github.com/PuerkitoBio/goquery"

	// "strings"
	"net/http"

	"net/url"
)

// Acquires a list of events from mtgtop8
func FetchEventList() ([]string, error) {
	
	// Meta to fetch for
	meta, err:= eventMeta()
	if err!=nil {
		return nil, err
	}

	// List of events to fetch
	lastLength:= -1
	events:= make([]string, 0)

	// Current page, 1 indexed
	page:= 1

	// Go until we haven't increased
	// the number of events
	for lastLength < len(events) {
		
		// Save length before fetch
		lastLength = len(events)

		// Fetch event page
		doc, err:= eventPaging(page, meta)
		if err!=nil{
			return nil, err
		}

		// Parse out the events
		fresh, err:= eventListFromDoc(doc)
		if err!=nil {
			return nil, err
		}
		// Add onto event list and dedup
		events = append(events, fresh...)
		events = dedupIDs(events)

		// Next page for next run
		page++
	}

	return events, nil
}

// Gets the nth page of listed events, 1 indexed.
func eventPaging(n int, meta string) (*gq.Document, error) {
	loc:= fmt.Sprintf("http://mtgtop8.com/format?f=MO&meta=%s", meta)
	page:= fmt.Sprintf("%d", n)

	resp, err:= http.PostForm(loc, url.Values{"cp": {page}})
	if err!=nil {
		return nil, err
	}

	return gq.NewDocumentFromResponse(resp) 
}

// Acquires the meta to fetch from
func eventMeta() (string, error) {
	// Fetch base page for modern format
	doc, err:= gq.NewDocument("http://mtgtop8.com/format?f=MO")
	if err!=nil {
		return "", fmt.Errorf("failed to fetch meta page", err)
	}

	// Default meta always the selected option of the meta selector
	sel:= doc.Find("select[name=meta]").Find("option[selected]")

	if sel.Length() != 1{
		fmt.Println(sel.Length())
		return "", fmt.Errorf("invalid number of default metas")
	}

	rel:= ""

	// No need to check, already filtered
	sel.Each(func(i int, o *gq.Selection){
		// No need to check, already filtered
		metaRel, ok:= o.Attr("value")
		if !ok {
			return
		}

		rel = metaRel
	})

	if len(rel) == 0 {
		return "", fmt.Errorf("no metas parsed")
	}

	// Parse query values
	u, err:= url.Parse(rel)
	if err !=nil {
		return "", fmt.Errorf("failed to parse relative meta %s", rel)
	}
	
	return u.Query().Get("meta"), nil
}

// Deduplicates a series of provided ids
//
// Requires ~2*len(ids) space
func dedupIDs(ids []string) []string {

	deduper:= make(map[string]struct{})

	for _, id:= range ids{
		deduper[id] = struct{}{}
	}

	cleaned:= make([]string, len(deduper))
	i:= 0
	for id:= range deduper{
		cleaned[i] = id
		i++
	}

	return cleaned
}

// Reads all events from a document,
// deduplication by caller is required for sanity.
func eventListFromDoc(doc *gq.Document) ([]string, error) {
	
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

		id:= u.Query().Get("e")
		ids = append(ids, id)
	})

	return ids, nil
}