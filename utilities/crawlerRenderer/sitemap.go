package main

import(

	"os"
	"fmt"
	"log"

	"text/template"
	"path/filepath"
	"strings"

)

type PassableUrls struct{
	Urls []string
}

// Removes all urls ending in index
// or that will get google angry
func (passable *PassableUrls) clean() {
	var validUrls []string

	for _, url:= range passable.Urls{
		if !strings.Contains(url, "/index") &&
		   !strings.Contains(url, "&") {
			validUrls = append(validUrls, url)
		}
	}

	passable.Urls = validUrls

}

func renderSiteMap(someMeta meta, urls []string, aLogger *log.Logger) {
	
	usableUrls:= PassableUrls{
		Urls: urls,
	}

	usableUrls.clean()
	aLogger.Println("Sitemap has ", len(usableUrls.Urls), " urls")

	mold, err:= getSitemapTemplate(someMeta.SiteMapTemplate)
	if err!=nil {
		aLogger.Fatalf("Failed to acquire template for sitemap, ", err)
	}

	target:= filepath.Join(twitterRenderLoc, "sitemap.xml")

	err = fillSiteMap(mold, &usableUrls, target)
	if err!=nil {
		aLogger.Fatalf("Failed to fill sitemap template, ", err)
	}


}

func fillSiteMap(mold *template.Template,
	content *PassableUrls,
	target string) error {
	
	dump, err:= os.Create(target)
	if err!=nil {
		return fmt.Errorf("Failed to open target, ", err)
	}

	return mold.Execute(dump, *content)

}

// Takes a template location and returns a ready to run template
// assuming no errors
func getSitemapTemplate(loc string) (*template.Template, error) {
	
	return template.ParseFiles(loc)

}