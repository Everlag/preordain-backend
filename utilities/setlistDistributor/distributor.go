package main
// Implements a small tool for managing setList.txt files.

import(

	"path/filepath"
	"os"

	"fmt"

	"flag"

	"strings"

	"io/ioutil"

)

var ignoredPathItems = []string{"cardCrops", "cardFulls",
"cardText", "cardSymbols"}

const masterName string = "setList.master.txt"

// Allows us to maintain non-global state while walking directories
type setListReplacer struct{
	ignoredPathItems []string

	setListLocations []string

	contents []byte
}

func (aReplacer *setListReplacer) loadContents() error {

	setList, err:= ioutil.ReadFile(masterName)
	if err!=nil {
		return err
	}

	aReplacer.contents = setList

	return nil

}

// Returns whether the provided path is one we care about
func (aReplacer *setListReplacer) validPath(path string) bool {

	for _, aSection:= range ignoredPathItems{
		if strings.Contains(path, aSection) {
			return false
		}
	}

	return true
}

func (aReplacer *setListReplacer) performReplacements() error {
	
	var err error
	for _, aPath:= range aReplacer.setListLocations{
		err = ioutil.WriteFile(aPath, aReplacer.contents, 0777)
		if err!=nil {
			return err
		}
	}

	return nil

}

func (aReplacer *setListReplacer) walkPath(path string,
	target os.FileInfo, err error) error {


	if !aReplacer.validPath(path) {
		return nil	
	}

	if target.Name() == "setList.txt" {
		aReplacer.setListLocations = append(aReplacer.setListLocations, path)
	}

	return nil

}

func main() {
	
	flag.Parse()
	root:= flag.Arg(0)
	if root == "" {
		fmt.Println("Please provide a path")
		os.Exit(1)
	}

	// Build a place to maintain state
	aSetListReplacer:= setListReplacer{ignoredPathItems:ignoredPathItems}

	// Acquire the contents of the master setList
	fmt.Println("Loading contents of master list")
	err:= aSetListReplacer.loadContents()
	if err!=nil {
		fmt.Println("Could not find ", masterName)
		os.Exit(1)
	}

	// Populate list of paths containing setList.txt
	fmt.Println("Starting to walk path")

	filepath.Walk(root, aSetListReplacer.walkPath)

	// Query the user if they want to replace these paths
	fmt.Println("Path walked\nThese locations were found:")

	for _, aLoc:= range aSetListReplacer.setListLocations{
		fmt.Println(" -> ", aLoc)
	}

	// Default to no
	fmt.Println("Proceed with replacement? [y/N]")
	var answer string
	fmt.Scanf("%s\n", &answer)
	if answer != "y" && answer != "Y" {
		fmt.Println("Aborting replacement, no changes made")
		os.Exit(0)
	}

	// Perform the replacement
	fmt.Println("Performing replacements")

	err = aSetListReplacer.performReplacements()
	if err!=nil{
		fmt.Println("Encountered error doing replacements, setLists may be left in inconsistent state. ",
			err)
		os.Exit(1)
	}

	fmt.Println("Replacements completed")

}