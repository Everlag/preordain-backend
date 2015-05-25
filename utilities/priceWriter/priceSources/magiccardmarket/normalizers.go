package magiccardmarket

// Provides translation between MKM set names and our names.
var dirtyToCleanSetNames = map[string]string{

	"Duel Decks: Phyrexia vs. The Coalition":"Duel Decks: Phyrexia vs. the Coalition",
	"Premium Deck Series: Fire & Lightning": "Premium Deck Series: Fire and Lightning",
	"Beatdown": "Beatdown Box Set",
	"Battle Royale": "Battle Royale Box Set",

	"Alpha": "Limited Edition Alpha",
	"Beta": "Limited Edition Beta",

	"Player Rewards Promos" : "Player Rewards",
	"Prerelease Promos": "Prerelease Events",
	"Friday Night Magic Promos": "Friday Night Magic",
	"DCI Promos": "Grand Prix",
	"Release Promos": "Launch Parties",
	"Game Day Promos": "Game Day",
	"Happy Holidays Promos": "Happy Holidays",
	"Judge Rewards Promos": "Judge Gift Program",
	"Harper Prism Promos": "Media Inserts",
	"Duels of the Planeswalkers Promos": "Duels of the Planeswalkers",

	"Conspiracy":"Magic: The Gathering—Conspiracy",

	"Guru Lands":"Guru",

	"Commander 2013":"Commander 2013 Edition",
	"Commander": "Magic: The Gathering-Commander",
	"Planechase 2012": "Planechase 2012 Edition",
	"Modern Masters 2015": "Modern Masters 2015 Edition",

	"Magic 2015":"Magic 2015 Core Set",
	"Magic 2014":"Magic 2014 Core Set",

	"Sixth Edition": "Classic Sixth Edition",

	"Revised": "Revised Edition",
	"Unlimited": "Unlimited Edition",
}

// Set names that we ignore because MKM doesn't support them in a way friendly
// friendly to us.
var ignoredSetNames = map[string]bool{
	"Pro Tour":true,
	"Time Spiral \"Timeshifted\"":true,
}

func isIgnoredSetName(name string) bool {

	return ignoredSetNames[name] ||
	name == ""

}