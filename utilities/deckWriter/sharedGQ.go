package main

import(

	gq "github.com/PuerkitoBio/goquery"

	"strings"

)

// Selects all links on the page which contain strictly links
// to events or decks in an event
func selectEventAnchors(doc *gq.Document) *gq.Selection {
	return doc.Find("a").FilterFunction(func(i int, a *gq.Selection) bool{
		
		rel, ok:= a.Attr("href")
		
		if ok && strings.Contains(rel, "event?e=") {
			return true
		}
		return false
	})	
}