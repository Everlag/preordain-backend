package nameNorm

import(

	"sort"
)

// Provide convenience wrappers to handle listing
// and testing the available decks

var names = []string{
	Jund, Junk,
	GRTron, UTron, UWTron,
	Affinity,
	Burn,
	URTwin, GrixisTwin, UWRTwin, TemurTwin,
	Scapeshift, TitanShift,
	GrixisControl,
	Infect, Bogles,
	LivingEnd, AdNauseam,
	Zoo, TribalZoo, BigZoo, SuicideZoo,
	Merfolk, AbzanCompany, CollectCallElves,
	UWRControl, UWControl,
	BWTokens, LanternControl,
	URDelver, GrixisDelver,
	FourCGifts, Stompy, SkredRed, Storm,
	Hatebears, SoulSisters, RUGAggro, Faeries,
	ToothAndNail, BloomTitan, PodlessPod,
	Grishoalbrand, TakingTurns,
}

// Keep a consistent ordering of names
func sortNames() {
	sort.Strings(names)
}

// Get a sorted list of names we have
func Names() (result []string) {
	result = make([]string, len(names))
	copy(result, names)

	return 
}

// Check if a provided name is something we have
func Valid(name string) bool {

	loc:= sort.SearchStrings(names, name)

	if loc == len(names) {
		return false
	}

	return name == names[loc]

}