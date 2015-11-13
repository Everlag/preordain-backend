package ApiServices

import(

	"net/http"
	"github.com/emicklei/go-restful"

	"time"
	"strconv"

	"./priceDBHandler.v2"
)

// Register price data returning singular points
func (aService *PriceService) registerClosest() {
	
	priceService:= aService.Service


	priceService.Route(priceService.
		GET("/Card/{cardName}/{setName}/{when}/Closest").
		To(aService.getCardClosestPoint).
		// Docs
		Doc("Price of a card with set closest to a given time").
		Operation("getCardLatestPoint").
		Param(priceService.PathParameter("cardName",
			"Name of a Magic: the Gathering card").DataType("string")).
		Param(priceService.PathParameter("setName",
			"Name of a Magic: the Gathering set").DataType("string")).
		Param(priceService.PathParameter("time",
			"Unix timestamp").DataType("int")).
		Param(priceService.QueryParameter("source",
			"Valid price source").DataType("string")).
		Writes(priceDB.Prices{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusBadRequest, BadTime, nil).
		Returns(http.StatusOK, "Latest price for a specific printing from DefaultPriceSource or a specific price source", nil))
}

func (aService *PriceService) getCardClosestPoint(req *restful.Request,
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

	// Fetch the timestamp as a string, convert that to an int64
	// and then to a priceDB.Timestamp
	timeString:= req.PathParameter("when")
	epoch, err:= strconv.ParseInt(timeString, 10, 64)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest,
			BadTime)
		return
	}
	t:= priceDB.Timestamp(time.Unix(epoch, 0))

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	cardPrice, err:= priceDB.GetCardClosest(aService.pool,
		cardName, setName, t, sourceName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(cardPrice)

}