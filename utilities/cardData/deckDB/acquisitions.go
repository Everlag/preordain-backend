package deckData

import(

	gq "github.com/PuerkitoBio/goquery"

	"strings"

	"fmt"

	"time"

)

// Returns the url of each archetype for the first page of that
// archetype. The first page will usually have some high profile
// events.
func (usableData *QueryableDeckData) gatherArchetypes() ([]string,
	[]string, error) {

	return findMtgTop8URLs(archeTypeLocation,
		generalLinkClass, archeTypePrefix)
}

// Acquires a map[decklistLoc]archeType for an archetype
// residing at loc
func gatherArcheTypeMap(name, loc string) (map[string]string, error) {
	rawLists, _, err:= findMtgTop8URLs(baseLocation + loc,
		generalLinkClass, eventPrefix, specificDeckPrefix)
	if err!=nil {
		return nil, err
	}

	results:= make(map[string]string)

	for _, url:= range rawLists{

		locations, _, err:= findMtgTop8URLs(baseLocation + url,
			deckListClass, deckPrefix)
		
		if err!=nil {
			// Skip this as something internal broke
			continue
		}
		if len(locations) != 1 {
			// Return an error because that means there was either no
			// export returned or more than one. Which is real bad.
			return nil,
			fmt.Errorf("more or less than one export url found for", url)
		}

		results[locations[0]] = name
	}

	return results, nil
}

// Given a location and a string that must appear in any link
// returned, we parse the page for matching anchor elements
// and return the link as well as the text content of the anchor
func findMtgTop8URLs(loc, class string, chunks ...string) ([]string, []string, error) {
	var doc *gq.Document
	var e error

	fmt.Print("\tfetching ", loc)

	start:= time.Now()

	doc, e = gq.NewDocument(loc)
	if e!= nil{
		return nil, nil, e
	}

	end:= time.Now()

	fmt.Println(" ", end.Sub(start))

	contents:= make([]string, 0)
	URLs:= make([]string, 0)

	// Find the archetypes
	doc.Find(class).Find("a").Each(func(i int, s *gq.Selection){

		// Grab the href
		href, exists:= s.Attr("href")
		if exists==false{
			return
		}

		for _, chunk:= range chunks{
			if !strings.Contains(href, chunk) {
				return
			}
		}

		URLs = append(URLs, href)
		contents = append(contents, s.Text())

	})

	return URLs, contents, nil
}