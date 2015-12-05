package deckData

import(

	"strings"

	"fmt"

	"os"
	"path/filepath"

)

const cacheFile string = "deckData.cache.json"

const sampleDeck string = `// Deck file for Magic Workstation created with mtgtop8.com
// NAME : UR Aggro
// CREATOR : ThomasH 
// FORMAT : 
        1 [RTR] Blood Crypt
        1 [UNH] Mountain
        1 [ISD] Sulfur Falls
        1 [UNH] Swamp
        1 [GTC] Watery Grave
        2 [RTR] Steam Vents
        3 [UNH] Island
        4 [KTK] Polluted Delta
        4 [ZEN] Scalding Tarn
        1 [FRF] Gurmag Angler
        2 [FRF] Tasigur, the Golden Fang
        3 [ISD] Snapcaster Mage
        3 [M14] Young Pyromancer
        4 [ISD] Delver of Secrets
        1 [ROE] Deprive
        1 [CMD] Electrolyze
        2 [DTK] Kolaghan's Command
        2 [M12] Mana Leak
        2 [RAV] Remand
        2 [DIS] Spell Snare
        2 [CMD] Terminate
        4 [M11] Lightning Bolt
        4 [DKA] Thought Scour
        2 [ROE] Inquisition of Kozilek
        3 [NPH] Gitaxian Probe
        4 [FD] Serum Visions
SB:  1 [RTR] Vandalblast
SB:  2 [NPH] Spellskite
SB:  1 [DTK] Rending Volley
SB:  1 [RTR] Rakdos Charm
SB:  1 [DTK] Negate
SB:  1 [JOU] Magma Spray
SB:  1 [RTR] Izzet Staticaster
SB:  1 [FD] Engineered Explosives
SB:  3 [M12] Dragon's Claw
SB:  1 [RTR] Dispel
SB:  2 [9E] Blood Moon`

// Removes all text between an opener and ender string for all instances
// of the opener-ender pairs
func ripAllBetween(text, opener, ender string) string {
	cleanedText:= text

	openerIndex := strings.Index(cleanedText, opener)
	closerIndex := strings.Index(cleanedText, ender)

	for openerIndex!=-1 &&
		closerIndex!=-1{

		cleanedText = strings.Replace(cleanedText, cleanedText[openerIndex:closerIndex+1], "", 1)

		openerIndex = strings.Index(cleanedText, opener)
		closerIndex = strings.Index(cleanedText, ender)
	}

	return cleanedText
}

var archetypeTranslation = map[string]string{
	"Red Deck Wins": "Burn",
	"UR Aggro": "Delver",
	"Affinity": "Affinity",
	"Mono Green Aggro": "Collect-Call Elves",
	"Junk": "Abzan",
	"Jund": "Jund",
	"Merfolk": "Merfolk",
	"Zoo": "Zoo",
	"Aura Hexproof": "Slippery Bogles",
	"4/5c Good Stuff": "Domain Zoo",
	"Dredgevine": "Dredgevine",
	"Hatebear": "Hatebears",
	"UrzaTron": "RG Tron",
	"UWx Midrange": "UWx Midrange",
	"UW Control": "UWr Control",
	"Martyr Life": "Soul Sisters",
	"The Rock": "The Rock",
	"Faeries": "UB Faeries",
	"Snow Red": "Skred Red",
	"Twin Exarch": "Splinter Twin",
	"Bloom Titan": "Amulet Bloom",
	"Infect": "Infect",
	"Birthing Pod": "Podless Collected",
	"UR Storm": "UR Storm",
	"Scapeshift": "Scapeshift",
	"Ad Nauseam": "Ad Nauseam",
	"Living End": "Living End",
}

// Translates a deck name from mtgtop8 to our vernacular.
//
// Deck names not present are ignored.
func Translate(name string) (string, error) {
	
	clean, ok:= archetypeTranslation[name]
	if !ok {
		return "", fmt.Errorf("failed to translate")
	}

	return clean, nil

}

// Returns the location of the cache file
// as specified by the CACHE environment variable.
//
// An empty CACHE variable directs output to the working directory.
func cacheLoc() string {

	// Fetch optionally specified cache location
	// root loc from environment
	loc:= os.Getenv("CACHE")
	if len(loc) == 0 {
		loc = "./"
	}

	return filepath.Join(loc, cacheFile)
}
