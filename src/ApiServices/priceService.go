package ApiServices

import(

	"net/http"
	"github.com/emicklei/go-restful"
	"./../../utilities/influxdbHandler"

	"./../goPrices/priceReader"

	"log"

	"fmt"

)

// Defaults
const DefaultPriceSource string = "mtgprice"

// Responses
const PriceDBError string = "Price DB lookup failed"
const BadCard string = "Illegal Card Name"
const BadSet string = "Illegal Set Name"
const BadCardFilter string = BadCard + " || " + BadSet
const BadCalculation string = "Failed Calculation"

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

// Which sources we currently support specific queries for
var validPriceSources = map[string]bool{"mtgprice":true,
	"magiccardmarket":true}

type PriceService struct{
	client *influxdbHandler.Client
	Service *restful.WebService
	logger *log.Logger
}

// Returns a fresh PriceService ready to be hooked up to restful
func NewPriceService() *PriceService {

	priceLogger:= GetLogger("priceLogger.txt", "priceLog")

	priceClient, err:= priceReader.AcquireReader()
	if err!=nil {
		priceLogger.Fatalln("Failed to acquire influxdb client, ", err)
	}
	
	aService:= PriceService{
		client: priceClient,
		logger: priceLogger,
	}

	err = aService.register(priceClient)
	if err!=nil {
		priceLogger.Fatalln("Failed to register PriceService, ", err)
	}

	return &aService

}

// Returns a restful.WebService exposing a priceReader based api
//
// Ugly but self documenting to the point of producing its own documentation
// as a json file on demand in a fancy ui.
func (aService *PriceService) register(aClient *influxdbHandler.Client) error {
	
	// Ensures we have a valid filter for card names/sets
	err:= populateCardMaps()
	if err!=nil {
		return err
	}

	priceService:= new(restful.WebService)
	priceService.
		Path("/api/Prices").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		ApiVersion("0.1")

	priceService.Route(priceService.
		GET("/SetList").To(aService.getSetList).
		// Docs
		Doc("Returns all available sets").
		Operation("getSetList").
		Writes([]string{}).
		Returns(http.StatusOK, "All available sets", nil))

	priceService.Route(priceService.
		GET("/SourceList").To(aService.getPriceSourcesList).
		// Docs
		Doc("Returns all available price sources").
		Operation("getPriceSourcesList").
		Writes([]string{}).
		Returns(http.StatusOK, "All available sources", nil))

	priceService.Route(priceService.
		GET("/Card/{cardName}").To(aService.getCard).
		// Docs
		Doc("Returns all prices for all printings of a card since the start of recording for all price sources").
		Operation("getCard").
		Param(priceService.PathParameter("cardName",
			"The name of a Magic: the Gathering card card").DataType("string")).
		Writes(influxdbHandler.Points{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCard, nil).
		Returns(http.StatusOK, "All prices for all printings from all sources", nil))

	priceService.Route(priceService.
		GET("/Card/{cardName}/{setName}").To(aService.getCardSpecificSet).
		// Docs
		Doc("Returns all prices for a printing of a card since the start of time").
		Operation("getCardSpecificSet").
		Param(priceService.PathParameter("cardName",
			"The name of a Magic: the Gathering card").DataType("string")).
		Param(priceService.PathParameter("setName",
			"The name of a Magic: the Gathering set").DataType("string")).
		Param(priceService.QueryParameter("source",
			"The name of a valid price source").DataType("string")).
		Writes(influxdbHandler.Points{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "All prices for a specific printing from DefaultPriceSource or specific price source", nil))

	priceService.Route(priceService.
		GET("/Card/{cardName}/{setName}/WeeksMedian").
		To(aService.getCardWeeksMedian).
		// Docs
		Doc("Returns all prices for a printing of a card since the start of time; Price Data is granular down to a week. Only for mtgprice").
		Operation("getCardWeeksMedian").
		Param(priceService.PathParameter("cardName",
			"The name of a Magic: the Gathering card").DataType("string")).
		Param(priceService.PathParameter("setName",
			"The name of a Magic: the Gathering set").DataType("string")).
		Param(priceService.QueryParameter("source",
			"The name of a valid price source").DataType("string")).
		Writes(influxdbHandler.Points{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "All prices for a specific printing from DefaultPriceSource or specific price source at week granularity", nil))

	priceService.Route(priceService.
		GET("/Card/{cardName}/{setName}/Latest").To(aService.getCardLatestPoint).
		// Docs
		Doc("Returns latest price for a printing of a card").
		Operation("getCardLatestPoint").
		Param(priceService.PathParameter("cardName",
			"The name of a Magic: the Gathering card").DataType("string")).
		Param(priceService.PathParameter("setName",
			"The name of a Magic: the Gathering set").DataType("string")).
		Param(priceService.QueryParameter("source",
			"The name of a valid price source").DataType("string")).
		Writes(influxdbHandler.Points{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "Latest price for a specific printing from DefaultPriceSource or a specific price source", nil))

	priceService.Route(priceService.
		GET("/Card/{cardName}/LatestLowest").To(aService.getCardLatestLowestPoint).
		// Docs
		Doc("Returns latest price for a printing of a card").
		Operation("getCardLatestLowestPoint").
		Param(priceService.PathParameter("cardName",
			"The name of a Magic: the Gathering card").DataType("string")).
		Param(priceService.QueryParameter("source",
			"The name of a valid price source").DataType("string")).
		Writes(ExtremaResponse{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "Lowest price across every printing from DefaultPriceSource or a specific price source", nil))

	priceService.Route(priceService.
		GET("/Card/{cardName}/LatestHighest").To(aService.getCardLatestHighestPoint).
		// Docs
		Doc("Returns latest price for a printing of a card").
		Operation("getCardLatestHighestPoint").
		Param(priceService.PathParameter("cardName",
			"The name of a Magic: the Gathering card").DataType("string")).
		Param(priceService.QueryParameter("source",
			"The name of a valid price source").DataType("string")).
		Writes(ExtremaResponse{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "Highest price across every printing from DefaultPriceSource or a specific price source", nil))

	priceService.Route(priceService.
		GET("/Set/{setName}/Latest").To(aService.getSetLatestPrices).
		// Docs
		Doc("Returns latest price for every card in a set").
		Operation("getSetLatestPrices").
		Param(priceService.PathParameter("setName",
			"The name of a Magic: the Gathering set").DataType("string")).
		Param(priceService.QueryParameter("source",
			"The name of a valid price source").DataType("string")).
		Writes(influxdbHandler.Points{}).
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



	aService.Service = priceService

	return nil

}

func (aService *PriceService) getSetList(req *restful.Request,
	resp *restful.Response) {
	
	setList:= make([]string, 0)
	for aSet, _:= range sets{
		if aSet != "" {
			setList = append(setList, aSet)	
		}
	}

	setCacheHeader(resp)

	resp.WriteEntity(setList)

}

func (aService *PriceService) getPriceSourcesList(req *restful.Request,
	resp *restful.Response) {
	
	//
	sourcesList:= make([]string, 0)
	for aSource, _:= range validPriceSources{
		if aSource != "" {
			sourcesList = append(sourcesList, aSource)	
		}
	}

	setCacheHeader(resp)

	resp.WriteEntity(sourcesList)

}

func (aService *PriceService) getCard(req *restful.Request,
	resp *restful.Response) {
	
	cardName:= req.PathParameter("cardName")

	// Ensure the card exists and that we want to find it
	if !cards[cardName] {
		resp.WriteErrorString(http.StatusBadRequest, BadCard)
		return
	}

	cardPrices, err:= aService.client.SelectEntireSeries(cardName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed")
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(cardPrices)

}

func (aService *PriceService) getCardSpecificSet(req *restful.Request,
	resp *restful.Response) {
	
	cardName:= req.PathParameter("cardName")
	setName:= req.PathParameter("setName")
	if !cards[cardName] {
		resp.WriteErrorString(http.StatusBadRequest, BadCard)
		return
	}
	if !sets[setName] {
		resp.WriteErrorString(http.StatusBadRequest, BadSet)
		return
	}

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	cardPrices, err:= aService.client.SelectFilteredSeries(cardName,
		setName, sourceName, 0)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(cardPrices)
}

func (aService *PriceService) getCardWeeksMedian(req *restful.Request,
	resp *restful.Response) {
	
	cardName:= req.PathParameter("cardName")
	setName:= req.PathParameter("setName")
	if !cards[cardName] {
		resp.WriteErrorString(http.StatusBadRequest, BadCard)
		return
	}
	if !sets[setName] {
		resp.WriteErrorString(http.StatusBadRequest, BadSet)
		return
	}

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	cardPrices, err:= aService.client.SelectWeeksMedian(cardName,
		setName, sourceName, 0)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(cardPrices)
}

func (aService *PriceService) getCardLatestPoint(req *restful.Request,
	resp *restful.Response) {
	
	cardName:= req.PathParameter("cardName")
	setName:= req.PathParameter("setName")
	if !cards[cardName] {
		resp.WriteErrorString(http.StatusBadRequest, BadCard)
		return
	}
	if !sets[setName] {
		resp.WriteErrorString(http.StatusBadRequest, BadSet)
		return
	}

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	cardPrices, err:= aService.client.SelectFilteredSeriesLatestPoint(cardName,
		setName, sourceName, 0)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(cardPrices)

}


// A response for a lowest latest price is minimalistic and aimed at providing
// exactly what is required
type ExtremaResponse struct{

	Name string

	Set string

	Price int64
}

const invalidPrice int64 = -1

func (aService *PriceService) getCardLatestLowestPoint(req *restful.Request,
	resp *restful.Response) {
	
	cardName:= req.PathParameter("cardName")
	if !cards[cardName] {
		resp.WriteErrorString(http.StatusBadRequest, BadCard)
		return
	}

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	prices, priceToSet, err:= aService.fetchAllPrices(cardName, sourceName)
	if err!=nil {
		// If we completely failed to grab a single price, then we can error out
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	lowestPrice:= findLowestPrice(prices)

	lowest:= ExtremaResponse{
		Name: cardName,
		Set: priceToSet[lowestPrice],
		Price: lowestPrice,
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(lowest)

}

func (aService *PriceService) getCardLatestHighestPoint(req *restful.Request,
	resp *restful.Response) {
	
	cardName:= req.PathParameter("cardName")
	if !cards[cardName] {
		resp.WriteErrorString(http.StatusBadRequest, BadCard)
		return
	}

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	prices, priceToSet, err:= aService.fetchAllPrices(cardName, sourceName)
	if err!=nil {
		// If we completely failed to grab a single price, then we can error out
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	highestPrice:= findHighestPrice(prices)

	highest:= ExtremaResponse{
		Name: cardName,
		Set: priceToSet[highestPrice],
		Price: highestPrice,
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(highest)

}

// Helper that grabs all possible prices for a card and dumps them into
// an array and map
func (aService *PriceService) fetchAllPrices(cardName,
	sourceName string) ([]int64, map[int64]string, error) {
	// Grab the prices across all sets
	priceToSet:= make(map[int64]string)
	prices:= make([]int64, len(cardsToSets[cardName]))

	i:= 0
	for setName, _:= range cardsToSets[cardName] {

		// Grind out the price for the card
		cardPrice, err:= aService.client.SelectFilteredSeriesLatestPoint(cardName,
		setName, sourceName, 0)
		if err!=nil {
			// Set the price to be unusable
			prices[i] = invalidPrice
			continue
		}
		parsedPriceMap, err := priceMapFromPoints(cardPrice)
		if err!=nil {
			// Set the price to be unusable
			prices[i] = invalidPrice
			continue
		}
		parsedPrice, ok:= parsedPriceMap[cardName]
		if !ok {
			// Set the price to be unusable
			prices[i] = invalidPrice
			continue
		}

		prices[i] = parsedPrice
		priceToSet[parsedPrice] = setName
		i++
	}

	// In the case to grab any prices, the map will be empty
	if len(priceToSet) == 0 {
		return nil, nil, fmt.Errorf("Failed to acquire any prices")
	}

	return prices, priceToSet, nil
}

// Returns the lowest price among an array of prices
func findLowestPrice(prices []int64) int64 {
	var lowest int64 = 1e10

	for _, aPrice:= range prices{
		if aPrice < lowest && aPrice!=invalidPrice {
			lowest = aPrice
		}
	}

	return lowest
}

// Returns the lowest price among an array of prices
func findHighestPrice(prices []int64) int64 {
	var highest int64 = -1

	for _, aPrice:= range prices{
		if aPrice > highest && aPrice!=invalidPrice {
			highest = aPrice
		}
	}

	return highest
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

	cardsToAcquire:= setsToCardsAndRarity.getCardName(setName)

	cardPrices, err:= aService.client.SelectSetsLatest(cardsToAcquire,
		setName, sourceName, 0)
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

	cardsToAcquire:= setsToCardsAndRarity.getCardName(setName)

	cardPrices, err:= aService.client.SelectSetsLatest(cardsToAcquire,
		setName, sourceName, 0)
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
	cardPrices influxdbHandler.Points) (EVResponse, error) {
	
	ev:= EVResponse{
		Name:aSet,
		IgnoredCommons: make([]string, 0),
		IgnoredUncommons: make([]string, 0),
		IgnoredRares: make([]string, 0),
		IgnoredMythics: make([]string, 0),
	}

	// The value of the set
	valuation, err:= priceMapFromPoints(cardPrices)
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
func priceMapFromPoints(cardPrices influxdbHandler.Points) (map[string]int64,
	error) {
	
	nameToPrice:= make(map[string]int64)

	// Turn the array of points into a map from card prices
	var priceIndex int
	for _, aPoint:= range cardPrices{
		priceIndex = aPoint.GetColumnIndex("price")
		if priceIndex < 0 || len(aPoint.Points) == 0 {
			return nameToPrice,
			fmt.Errorf("Failed to find price for ", aPoint)
		}

		price, ok:= aPoint.Points[0][priceIndex].(float64)
		if !ok {
			return nameToPrice,
			fmt.Errorf("Failed to assert price for ", aPoint)	
		}

		nameToPrice[aPoint.Name] = int64(price)
	}

	return nameToPrice, nil

}