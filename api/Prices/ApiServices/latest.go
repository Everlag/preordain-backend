package ApiServices

import(

	"net/http"
	"github.com/emicklei/go-restful"

	"./../../../utilities/priceDBHandler.v2"
)

// Register price data returning singular points
func (aService *PriceService) registerLatest() {
	
	priceService:= aService.Service


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
		Writes(priceDB.Prices{}).
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
		Writes(priceDB.Price{}).
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
		Writes(priceDB.Price{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "Highest price across every printing from DefaultPriceSource or a specific price source", nil))

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

	cardPrice, err:= priceDB.GetCardLatest(aService.pool,
		cardName, setName, sourceName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(cardPrice)

}

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

	lowest, err:= priceDB.GetCardLatestLowest(aService.pool,
		cardName, sourceName)
	if err!=nil {
		// If we completely failed to grab a single price, then we can error out
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
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

	highest, err:= priceDB.GetCardLatestHighest(aService.pool,
		cardName, sourceName)
	if err!=nil {
		// If we completely failed to grab a single price, then we can error out
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(highest)

}