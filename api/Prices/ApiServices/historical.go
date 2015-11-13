package ApiServices

import(

	"net/http"
	"github.com/emicklei/go-restful"

	"./priceDBHandler.v2"

)

// Register price data returning multiple points
func (aService *PriceService) registerHistorical() {
	
	priceService:= aService.Service

	priceService.Route(priceService.
		GET("/Card/{cardName}/{setName}").To(aService.getCardSpecificSet).
		// Docs
		Doc("All prices for a printing of a card since the start of time").
		Operation("getCardSpecificSet").
		Param(priceService.PathParameter("cardName",
			"Name of a Magic: the Gathering card").DataType("string")).
		Param(priceService.PathParameter("setName",
			"Name of a Magic: the Gathering set").DataType("string")).
		Param(priceService.QueryParameter("source",
			"Valid price source").DataType("string")).
		Writes(priceDB.Prices{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "All prices for a specific printing from DefaultPriceSource or specific price source", nil))

	priceService.Route(priceService.
		GET("/Card/{cardName}/{setName}/WeeksMedian").
		To(aService.getCardWeeksMedian).
		// Docs
		Doc("All prices for a printing of a card since the start of time with week granularity").
		Operation("getCardWeeksMedian").
		Param(priceService.PathParameter("cardName",
			"Name of a Magic: the Gathering card").DataType("string")).
		Param(priceService.PathParameter("setName",
			"Name of a Magic: the Gathering set").DataType("string")).
		Param(priceService.QueryParameter("source",
			"Valid price source").DataType("string")).
		Writes(priceDB.Prices{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "All prices for a specific printing from DefaultPriceSource or specific price source at week granularity", nil))

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

	cardPrices, err:= priceDB.GetCardHistory(aService.pool,
		cardName, setName, sourceName)
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

	cardPrices, err:= priceDB.GetCardMedianHistory(aService.pool,
		cardName, setName, sourceName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(cardPrices)
}