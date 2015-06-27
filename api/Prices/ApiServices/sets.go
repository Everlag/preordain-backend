package ApiServices

import(

	"fmt"

	"net/http"
	"github.com/emicklei/go-restful"

	"./../../../utilities/priceDBHandler.v2"

)

// Constants used for calculating the EV from a box for a set
const Common string = "Common"
const Uncommon string = "Uncommon"
const Rare string = "Rare"
const Mythic string = "Mythic Rare"

const CommonsPerBox float64 = 396
const UncommonsPerBox float64 = 108
const RaresPerBox float64 = 31.5
const MythicsPerBox float64 = 4.5

const MythicMinImpact float64 = 0.0
const RareMinImpact float64 = 0.0
const UncommonMinImpact float64 = 100.0
const CommonMinImpact float64 = 100.0

// Register price data returning data concerning
// full sets.
func (aService *PriceService) registerSets() {
	
	priceService:= aService.Service

	priceService.Route(priceService.
		GET("/Set/{setName}/Latest").To(aService.getSetLatestPrices).
		// Docs
		Doc("Returns latest price for every card in a set").
		Operation("getSetLatestPrices").
		Param(priceService.PathParameter("setName",
			"The name of a Magic: the Gathering set").DataType("string")).
		Param(priceService.QueryParameter("source",
			"The name of a valid price source").DataType("string")).
		Writes(priceDB.Prices{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "Latest prices for a set from DefaultPriceSource or a specific price source", nil))

	priceService.Route(priceService.
		GET("/Set/{setName}/EV").To(aService.getSetLatestBoxEV).
		// Docs
		Doc("Returns calculated expected value for a box of this set").
		Operation("getSetLatestBoxEV").
		Param(priceService.PathParameter("setName",
			"The name of a Magic: the Gathering set").DataType("string")).
		Param(priceService.QueryParameter("source",
			"The name of a valid price source").DataType("string")).
		Writes(EVResponse{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusInternalServerError, BadCalculation, nil).
		Returns(http.StatusOK, "Latest EV for a set from DefaultPriceSource or a specific price source", nil))


}


func (aService *PriceService) getSetLatestPrices(req *restful.Request,
	resp *restful.Response) {
	
	setName:= req.PathParameter("setName")
	if !sets[setName] {
		resp.WriteErrorString(http.StatusBadRequest, BadSet)
		return
	}

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	cardPrices, err:= priceDB.GetSetLatest(aService.pool,
		setName, sourceName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	setCacheHeader(resp)

	resp.WriteEntity(cardPrices)

}

func (aService *PriceService) getSetLatestBoxEV(req *restful.Request,
	resp *restful.Response) {
	
	// Administrative and price acquiring
	setName:= req.PathParameter("setName")
	if !sets[setName] {
		resp.WriteErrorString(http.StatusBadRequest, BadSet)
		return
	}

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	cardPrices, err:= priceDB.GetSetLatest(aService.pool,
		setName, sourceName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	ev, err:= calculateEV(setName, cardPrices)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			BadCalculation + err.Error())
		return
	}

	setCacheHeader(resp)

	resp.WriteEntity(ev)

}

// A response for expected value of a box is comprehensive and includes
// numbers acquired along with way. This provides deeper meaning to users.
type EVResponse struct{

	Name string

	EV int64
	MythicContributed int64
	RareContributed int64
	UncommonContributed int64
	CommonContributed int64

	IgnoredCommons, IgnoredUncommons, IgnoredRares, IgnoredMythics []string

}

// Calculates the EV of a box of this set.
//
// Takes an array of Points containing all cards in that set.
//
// EV is calculated as follows:
// for each Mythic
// 	MythicAdditionToEV = (MythicPrice * MythicsPerBox) / TotalMythicCount
// And repeat for rares.
// Uncommons and commons are much the same but we only consider those
// that are $1.5 or greater
func calculateEV(aSet string,
	cardPrices priceDB.Prices) (EVResponse, error) {
	
	ev:= EVResponse{
		Name:aSet,
		IgnoredCommons: make([]string, 0),
		IgnoredUncommons: make([]string, 0),
		IgnoredRares: make([]string, 0),
		IgnoredMythics: make([]string, 0),
	}

	// The value of the set
	valuation, err:= priceMapFromPrices(cardPrices)
	if err!=nil {
		return EVResponse{}, fmt.Errorf("Failed to acquire prices, ", err)
	}

	// Get the cards by rarity for the set
	mythics:= setsToCardsAndRarity.getCardsWithRarity(aSet, Mythic)
	rares:= setsToCardsAndRarity.getCardsWithRarity(aSet, Rare)
	uncommons:= setsToCardsAndRarity.getCardsWithRarity(aSet, Uncommon)
	commons:= setsToCardsAndRarity.getCardsWithRarity(aSet, Common)

	// Sum the values
	MythicContributed:= calculateEVForRarity(valuation,
		mythics, MythicsPerBox, MythicMinImpact, &ev.IgnoredMythics)
	RareContributed:= calculateEVForRarity(valuation,
		rares, RaresPerBox, RareMinImpact, &ev.IgnoredRares)
	UncommonContributed:= calculateEVForRarity(valuation,
		uncommons, UncommonsPerBox, UncommonMinImpact, &ev.IgnoredUncommons)
	CommonContributed:= calculateEVForRarity(valuation,
		commons, CommonsPerBox, CommonMinImpact, &ev.IgnoredCommons)

	sum:= MythicContributed + RareContributed +
	UncommonContributed + CommonContributed

	ev.EV = int64(sum)
	ev.MythicContributed = int64(MythicContributed)
	ev.RareContributed = int64(RareContributed)
	ev.UncommonContributed = int64(UncommonContributed)
	ev.CommonContributed = int64(CommonContributed)

	return ev, nil

}

// A small helper function to keep the calculation of rarities reasonable
//
// impactMinimum allows a selection of a minimum value to prevent bulk from
// having too much of an impact
func calculateEVForRarity(valuation map[string]int64, cards []string,
	RarityPerBox, impactMinimum float64,
	ignoredContainer *[]string) (contribution float64) {
	
	possibleOthers:= float64(len(cards))
	impactCoefficient:= RarityPerBox / possibleOthers

	var price int64
	var impact float64
	var ok bool
	for _, aCard:= range cards {
		price, ok = valuation[aCard]
		if !ok {
			*ignoredContainer = append(*ignoredContainer, aCard)
			continue
		}

		impact = float64(price) * impactCoefficient
		if impact > impactMinimum {
			contribution+= 	impact
		}
	}

	return

}

// Generates a map converting each card name, the name of each point,
// to its current value as provided.
//
// Duplicate points can override each other non-deterministically so don't
// expect coherence if that happens
func priceMapFromPrices(cardPrices priceDB.Prices) (map[string]int64,
	error) {
	
	nameToPrice:= make(map[string]int64)

	for _, p:= range cardPrices{
		nameToPrice[p.Name] = int64(p.Price)
	}

	return nameToPrice, nil

}