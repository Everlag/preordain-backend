package nameNorm

import (
	"fmt"

	"./../deckData"

	// "strings"
)

// Map a name to any number of potential names we think
// the deck could have, the content filter will take
// care of that.
type nameFilter map[string]contentFilter

// Given a deck, determines what its name should be
// using the internal filters this nameFilter can access
func (f nameFilter) determine(d *deckData.Deck) error {

	// Normalize the incoming name
	clean:= Normalize(d.Name)

	content, ok:= f[clean]
	if !ok {
		return fmt.Errorf("no matches for name")
	}

	d.Name = content.determine(d)

	return nil
}

// Attempts to invert a provided clean name into the
// mtgtop8 deck names that may have spawned it.
//
// Returns invertedTuples due to some decks requiring
// the presence of a specific card  
func (f nameFilter) invert(clean string) []invertedTuple {
	
	// Intermediary layer before we deduplicate
	result:= make([]invertedTuple, 0)

	for dirty, second:= range f{
		// Invert and check if clean is a key
		mapped, ok:= second.invert(clean)
		if !ok {
			continue
		}

		// Fetch the card needed for the translation
		card:= mapped[clean]


		// Handle potential for exluding cards
		// from this mapped section due to having the
		// special case Default value
		cards:= make([]string, 0)
		excludeCards:= false
		if card == Default {

			excludeCards = true

			// We need to exlude every other card
			// that is used to signal another deck under this
			// mapping
			for _, exclude:= range mapped{
				if exclude == Default {
					continue
				}

				cards = append(cards, exclude)
			}
		}
		
		// The specific card is always included
		cards = append(cards, card)

		for _, c:= range cards{
			inverted:= invertedTuple{
				Name: dirty,
				Card: c,
			}
			if excludeCards && c != Default {
				inverted.Exclude = true
			}
			result = append(result, inverted)
		}

	}

	return result
}

// A tuple representing the name of a deck as well
// as a card that distinguishes a deck.
//
// Necessary for inversion as we would return Urzatron
// for both GR Tron and U tron, as an example
//
// Name as 'Default' is a flag to the deckDB package that
// no cards is explicitly required for this deck.
//
// If the exclude flag is set, any decks containing this card
// should be excluded rather than included.
type invertedTuple struct{
	Name, Card string
	Exclude bool
}

// Use the prescence of a card to indicate the distinct
// archetype.
//
// The Default key indicates the name if no
// specific match is made. A default key is required
type contentFilter map[string]string

// Given a deck, the filter checks if anything could
// fit or returns its default value.
//
// A content filter requires that a deck is already in
// its archetype.
func (f contentFilter) determine(d *deckData.Deck) string {

	// Single possibility means we just return default
	// and avoid iterating
	if len(f) == 1 {
		return f[Default]
	}

	for _, c:= range d.Maindeck{
		specific, ok:= f[c.Name]
		if !ok {
			continue
		}

		return specific
	}

	return f[Default]
}

// Inverts this content filter only if the provided
// clean name would be a key in that new filter
//
// This follows the 'value, ok' map handling
func (f contentFilter) invert(clean string) (map[string]string, bool) {

	// We invert it before checking, its just easier
	inv:= make(map[string]string)

	for k, v:= range f{
		inv[v] = k
	}

	_, ok:= inv[clean]

	return inv, ok

}

// A deck name
type name string
// A card name
type content string

// From here down are manually generated filters sets;
// be gentle, these are fragile!

// Populates the registered names filter
func populateTopLevel() {
	
	// Single filters that just map to a single
	// default value
	burnFilter:= contentFilter{Default: Burn}
	affinityFilter:= contentFilter{Default: Affinity}
	jundFilter:= contentFilter{Default: Jund}
	junkFilter:= contentFilter{Default: Junk}
	merfolkFilter:= contentFilter{Default: Merfolk}
	hatebearFilter:= contentFilter{Default: Hatebears}
	bogleFilter:= contentFilter{Default: Bogles}
	martyrlifeFilter:= contentFilter{Default: SoulSisters}
	rugaggroFilter:= contentFilter{Default: RUGAggro}
	uwFilter:= contentFilter{Default: UWControl}
	lanternFilter:= contentFilter{Default: LanternControl}
	valakutFilter:= contentFilter{Default: TitanShift}
	faeriesFilter:= contentFilter{Default: Faeries}
	toothandnailFilter:= contentFilter{Default: ToothAndNail}
	bloomTitanFilter:= contentFilter{Default: BloomTitan}
	infectFilter:= contentFilter{Default: Infect}
	podlesscocoFilter:= contentFilter{Default: PodlessPod}
	livingendFilter:= contentFilter{Default: LivingEnd}
	stormFilter:= contentFilter{Default: Storm}
	adnauseamFilter:= contentFilter{Default: AdNauseam}
	reanimatorFilter:= contentFilter{Default: Grishoalbrand}
	turnsFilter:= contentFilter{Default: TakingTurns}

	// Multiple choices based on card
	zooFilter:= contentFilter{
		Default: Zoo,
		"Tribal Flames": TribalZoo,
		"Geist of Saint Traft": BigZoo,
	}

	monogreenFilter:= contentFilter{
		Default: Stompy,
		"Heritage Druid": CollectCallElves,
	}

	uraggroFilter:= contentFilter{
		Default: URDelver,
		"Tasigur, the Golden Fang": GrixisDelver,
		"Gurmag Angler": GrixisDelver,
	}

	tronFilter:= contentFilter{
		Default: GRTron,
		"Gifts Ungiven": UWTron,
		"Treasure Mage": UTron,
	}

	grixiscontrolFilter:= contentFilter{
		Default: GrixisControl,
		"Delver of Secrets": GrixisDelver,
	}

	twinFilter:= contentFilter{
		Default: URTwin,
		"Kolaghan's Command": GrixisTwin,
		"Path to Exile": UWRTwin,
		"Tarmogoyf": TemurTwin,
	}

	scapeshiftFilter:= contentFilter{
		Default: Scapeshift,
		"Primeval Titan": TitanShift,
	}

	// Map from mtgtop8 archetype names to our filters
	f:= nameFilter{
		"Red Deck Wins": burnFilter,
		"Affinity": affinityFilter,
		"Jund": jundFilter,
		"Junk": junkFilter,
		"Zoo": zooFilter,
		"Merfolk": merfolkFilter,
		"Hatebear": hatebearFilter,
		"Aura Hexproof": bogleFilter,
		"4/5c Good Stuff": zooFilter,
		"Mono Green Aggro": monogreenFilter,
		"UR Aggro": uraggroFilter,
		"Martyr Life": martyrlifeFilter,
		"RUG Aggro": rugaggroFilter,
		"UrzaTron": tronFilter,
		"GR Tron": tronFilter,
		"Grixis Control": grixiscontrolFilter,
		"UWx Midrange": uwFilter,
		"UW Midrange": uwFilter,
		"UW Control": uwFilter,
		"Lantern Control": lanternFilter,
		"Valakut": valakutFilter,
		"Titan Valakut": valakutFilter,
		"Faeries": faeriesFilter,
		"Tooth and Nail": toothandnailFilter,
		"Evil Twin": twinFilter,
		"UWR Twin": twinFilter,
		"Jeskai Twin": twinFilter,
		"UR Twin": twinFilter,
		"Twin Exarch": twinFilter,
		"Bloom Titan": bloomTitanFilter,
		"Infect": infectFilter,
		"Scapeshift": scapeshiftFilter,
		"Melira Co Co": podlesscocoFilter,
		"Melira Company": podlesscocoFilter,
		"Podless Collected": podlesscocoFilter,
		"Living End": livingendFilter,
		"UR Storm": stormFilter,
		"Ad Nauseam": adnauseamFilter,
		"Instant Reanimator": reanimatorFilter,
		"Walks": turnsFilter,
	}

	// Register our filter at the top level
	// with the loose naming
	topLevel = make(nameFilter)
	for k, v:= range f{
		topLevel[Normalize(k)] = v
	}

}