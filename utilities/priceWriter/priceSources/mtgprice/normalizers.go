package mtgprice

import(

	"strings"

)

//there are sets we need to map to mtgprice's naming scheme beyond simply cleaning it
//also, there are sets that we simply don't want. those map to blanks
var dirtySetNames = map[string]string{
	//product that isn't supported by mtgprice
	"Anthologies": "",
	"Duel Decks: Heroes vs. Monsters": "Duel_Decks_Heroes_vs_Monsters",
	"Duel Decks: Jace vs. Vraska": "Duel_Decks_Jace_vs_Vraska",
	"Introductory Two-Player Set":"",
	"Masters Edition": "",
	"Masters Edition II": "",
	"Masters Edition III": "",
	"Masters Edition IV": "",
	"Time Spiral \"Timeshifted\"":"Timespiral_Timeshifted",
	"Vanguard":"",
	"Promo set for Gatherer":"",
	"Rivals Quick Start Set":"",
	"Modern Event Deck 2014":"",
	"Vintage Masters": "",
	"Battle for Zendikar Foil":"",
	"Battle for Zendikar":"",
	"Zendikar Expedition":"",


	//now supported product!
	//"Duels of the Planeswalkers": "",

	//specific supplementary product
	"Modern Masters 2015 Edition": "Modern Masters 2015",
	"Magic: The Gathering-Commander":"Commander",
	"Commander 2013 Edition": "Commander_2013",
	"Deckmasters": "Deckmasters_Box_Set",
	"Planechase 2012 Edition": "Planechase 2012",
	"Duel Decks: Phyrexia vs. the Coalition": "Duel_Decks_Phyrexia_vs_The_Coalition",

		//deal wizards, screw you for THAT GODDAMN LONG HYPHEN.
		//IT BROKE SO MANY THINGS.
		//sincerly, whoever has to maintain this.
	"Magic: The Gathering—Conspiracy": "Conspiracy",

	//expert level sets with names modified
	"Ravnica: City of Guilds": "Ravnica",
	"Journey into Nyx": "Journey_Into_Nyx",

	//core-or core equivalent- sets
	"Revised Edition": "Revised",
	"Magic 2016 Core Set": "M16",
	"Magic 2015 Core Set": "M15",
	"Magic 2014 Core Set": "M14",
	"Magic 2013": "M13",
	"Magic 2012": "M12",
	"Magic 2011": "M11",
	"Magic 2010": "M10",
	"Tenth Edition": "10th Edition",
	"Ninth Edition": "9th Edition",
	"Eighth Edition": "8th Edition",
	"Seventh Edition": "7th Edition",
	"Classic Sixth Edition": "6th Edition",
	"Fifth Edition": "5th Edition",
	"Fourth Edition": "4th Edition",
	"Unlimited Edition": "Unlimited",
	"Limited Edition Alpha": "Alpha",
	"Limited Edition Beta": "Beta",
	

}

//cleans a set's name to be accepted by mtgprice's servers
func cleanSetName(set string) string {
	//deal with special cases
	properName, ok:= dirtySetNames[set]
	if ok{
		//we found a set we need to change
		set = properName
		if properName == "" {
			//we have a set that just doesn't work...
			return ""
		}
	}

	//there exists the special case for foil sets.
	if strings.Index(set, " Foil")!=-1{
		//we need to get it without the foil section
		//to pass it to the cleaner
		set = strings.Replace(set, " Foil", "", -1)

		properName, ok:= dirtySetNames[set]
		if ok{
			//we found a set we need to change
			set = properName
			if properName == "" {
				//we have a set that just doesn't work...
				return ""
			}
		}

		//now we add the foil tag back on
		set = set + "_(Foil)"
	}
	

	//remove spaces, apostrophes, and other characters stripped
	//from mtgprice's names
	set = strings.Replace(set, " ", "_", -1)
	set = strings.Replace(set, "'", "", -1)
	set = strings.Replace(set, ":", "", -1)
	set = strings.Replace(set, ".", "", -1)


	return set
}

//there are cards whose unescaped names, its a fairly new and raw
//api- nothing to fault them for, WILL cause a failure during
//unmarshalling
var dirtyCardNames = map[string]string{
	"Kongming, \"Sleeping Dragon\"":"Kongming, Sleeping Dragon",
	"Pang Tong, \"Young Phoenix\"": "Pang Tong, Young Phoenix",
	"\"Ach! Hans, Run!\"": "Ach! Hans, Run!",
}

//sometimes there are cards which are improperly escaped. we handle that!
func handleSpecialCardCases(setData []byte) []byte {
	stringed:= string(setData)

	//inefficient but general and fairly quick.
	//not something that will scale, but it shouldn't have to.
	for raw, fixed:= range dirtyCardNames{
		stringed = strings.Replace(stringed, raw, fixed, -1)
	}

	return []byte(stringed)
}
