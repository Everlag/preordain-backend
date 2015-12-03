// Package setlist provides standardized access to
// set lists found on disk.
//
// A setList is defined as a file named setList.txt which
// has each line as either a Magic set name or blank.
package setlist

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"strings"
)

// Suffix for a set defined as a foil version
// of another set.
const FoilSuffix string = " Foil"

// Standardized name for a setList on disk
const setListName string = "setList.txt"

// Acquires setList.txt from a location specified in
// the SETLIST environment variable.
//
// If SETLIST is blank, fetch a setlist in a local directory.
func Get() ([]string, error) {
	// Fetch optionally specified set list
	// root loc from environment
	loc := os.Getenv("SETLIST")
	if len(loc) == 0 {
		loc = "./"
	}

	setListLoc := filepath.Join(loc, setListName)

	setsRaw, err := ioutil.ReadFile(setListLoc)
	if err != nil {
		return nil, err
	}

	// Trim excess whitespace from set names
	sets := strings.Split(string(setsRaw), "\n")
	for i, set := range sets {
		set = strings.TrimSpace(set)

		sets[i] = set
	}

	return sets, nil
}

// Acquire a map[set]foilVariant, follows the same handling
// of environment config as Get.
//
// Sets without a foil variant map to an empty string.
//
// Useful for mapping from a single set list to multiple
func FoilMapping() (map[string]string, error) {

	// Acquire complete list of supported sets
	setList, err := Get()
	if err != nil {
		return nil, err
	}

	// map[set]foilVariant
	setMap := make(map[string]string)

	for _, aSet := range setList {

		// Check for foil set
		if strings.Contains(aSet, FoilSuffix) {

			// Add if foil to mapping
			normalName := strings.Replace(aSet, FoilSuffix, "", -1)
			setMap[normalName] = aSet

		} else {
			// Non-foil set means we check if its present already
			// to avoid overriding
			_, ok := setMap[aSet]
			if ok {
				// Already present
				continue
			}

			// Add blank foil variant
			setMap[aSet] = ""

		}

	}

	return setMap, nil
}
