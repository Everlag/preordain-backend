package ApiServices

import(

	"net/http"
	"github.com/emicklei/go-restful"
	"./../../utilities/influxdbHandler"

	"./../goPrices/priceReader"

	"log"

)

const DefaultPriceSource string = "mtgprice"
const PriceDBError string = "Price DB lookup failed"
const BadCard string = "Illegal Card Name"
const BadSet string = "Illegal Set Name"
const BadCardFilter string = BadCard + " || " + BadSet

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

	err = aService.register()
	if err!=nil {
		priceLogger.Fatalln("Failed to register PriceService, ", err)
	}

	return &aService

}

// Returns a restful.WebService exposing a priceReader based api
//
// Ugly but self documenting to the point of producing its own documentation
// as a json file on demand in a fancy ui.
func (aService *PriceService) register() error {
	
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
		GET("/Card/{cardName}").To(aService.getCard).
		// Docs
		Doc("Returns all prices for all printings of a card since the start of time").
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
		Writes(influxdbHandler.Points{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "All prices for a specific printing from DefaultPriceSource", nil))

	priceService.Route(priceService.
		GET("/Card/{cardName}/{setName}/WeeksMedian").To(aService.getCardWeeksMedian).
		// Docs
		Doc("Returns all prices for a printing of a card since the start of time; Price Data is granular down to a week").
		Operation("getCardWeeksMedian").
		Param(priceService.PathParameter("cardName",
			"The name of a Magic: the Gathering card").DataType("string")).
		Param(priceService.PathParameter("setName",
			"The name of a Magic: the Gathering set").DataType("string")).
		Writes(influxdbHandler.Points{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "All prices for a specific printing from DefaultPriceSource at week granularity", nil))

	priceService.Route(priceService.
		GET("/Card/{cardName}/{setName}/Latest").To(aService.getCardLatestPoint).
		// Docs
		Doc("Returns latest price for a printing of a card").
		Operation("getCardLatestPoint").
		Param(priceService.PathParameter("cardName",
			"The name of a Magic: the Gathering card").DataType("string")).
		Param(priceService.PathParameter("setName",
			"The name of a Magic: the Gathering set").DataType("string")).
		Writes(influxdbHandler.Points{}).
		Returns(http.StatusInternalServerError, "Price DB lookup failed", nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "Latest price for a specific printing from DefaultPriceSource", nil))

	aService.Service = priceService

	return nil

}

func (aService *PriceService) getCard(req *restful.Request, resp *restful.Response) {
	
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

	cardPrices, err:= aService.client.SelectFilteredSeries(cardName,
		setName, DefaultPriceSource, 0)
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

	cardPrices, err:= aService.client.SelectWeeksMedian(cardName,
		setName, DefaultPriceSource, 0)
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

	cardPrices, err:= aService.client.SelectFilteredSeriesLatestPoint(cardName,
		setName, DefaultPriceSource, 0)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError,
			"Price DB lookup failed, ")
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(cardPrices)

}
