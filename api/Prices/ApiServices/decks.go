package ApiServices

import(

	"net/http"
	"github.com/emicklei/go-restful"

	"./../../../common/priceDB"
	"./../../../common/deckDB/deckData"

	"fmt"

	"os"
	"io/ioutil"
	"encoding/json"

)

// Register price data for Decks api
func (aService *PriceService) registerDecks() {
	
	priceService:= aService.Service

	priceService.Route(priceService.
		GET("/Deck/{deckID}/Lowest").To(aService.getDeckPriceLowest).
		// Docs
		Doc("Latest and lowest prices for each card in a deck").
		Operation("getDeckPriceLowest").
		Param(priceService.PathParameter("deckID",
			"A valid deck identifer usable in the Decks api").DataType("string")).
		Param(priceService.QueryParameter("source",
			"Valid price source").DataType("string")).
		Writes(priceDB.Prices{}).
		Returns(http.StatusInternalServerError, PriceDBError, nil).
		Returns(http.StatusInternalServerError, RemoteAPIError, nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "Lowest card prices for a specific deck from DefaultPriceSource or specific price source", nil))

	priceService.Route(priceService.
		GET("/Deck/{deckID}/Highest").To(aService.getDeckPriceHighest).
		// Docs
		Doc("Latest and highest prices for each card in a deck").
		Operation("getDeckPriceHighest").
		Param(priceService.PathParameter("deckID",
			"A valid deck identifer usable in the Decks api").DataType("string")).
		Param(priceService.QueryParameter("source",
			"Valid price source").DataType("string")).
		Writes(priceDB.Prices{}).
		Returns(http.StatusInternalServerError, PriceDBError, nil).
		Returns(http.StatusInternalServerError, RemoteAPIError, nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "Highest card prices for a specific deck from DefaultPriceSource or specific price source", nil))

	priceService.Route(priceService.
		GET("/Deck/{deckID}/Weekly/Low").To(aService.getDeckWeeklyLowest).
		// Docs
		Doc("Lowest weekly prices for a deck").
		Operation("getDeckWeeklyLowest").
		Param(priceService.PathParameter("deckID",
			"A valid deck identifer usable in the Decks api").DataType("string")).
		Param(priceService.QueryParameter("source",
			"Valid price source").DataType("string")).
		Writes(priceDB.SummedWeek{}).
		Returns(http.StatusInternalServerError, PriceDBError, nil).
		Returns(http.StatusInternalServerError, RemoteAPIError, nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "Lowest weekly summed prices for a specific deck from DefaultPriceSource or specific price source", nil))

	priceService.Route(priceService.
		GET("/Deck/{deckID}/Weekly/High").To(aService.getDeckWeeklyHighest).
		// Docs
		Doc("Highest weekly prices for a deck").
		Operation("getDeckWeeklyHighest").
		Param(priceService.PathParameter("deckID",
			"A valid deck identifer usable in the Decks api").DataType("string")).
		Param(priceService.QueryParameter("source",
			"Valid price source").DataType("string")).
		Writes(priceDB.SummedWeek{}).
		Returns(http.StatusInternalServerError, PriceDBError, nil).
		Returns(http.StatusInternalServerError, RemoteAPIError, nil).
		Returns(http.StatusBadRequest, BadCardFilter, nil).
		Returns(http.StatusOK, "Highest weekly summed prices for a specific deck from DefaultPriceSource or specific price source", nil))
}


func (aService *PriceService) getDeckPriceLowest(req *restful.Request,
	resp *restful.Response) {
	
	deckid:= req.PathParameter("deckID")

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	// Grab the decklist from the remote
	d, err:= fetchRemoteDecklist(deckid)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, RemoteAPIError)
		return
	}

	cards, _:= deckToCardList(d)

	prices, err:= priceDB.GetBulkLatestLowest(aService.pool, cards, sourceName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, PriceDBError)
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(prices)
}

func (aService *PriceService) getDeckPriceHighest(req *restful.Request,
	resp *restful.Response) {
	
	deckid:= req.PathParameter("deckID")

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	// Grab the decklist from the remote
	d, err:= fetchRemoteDecklist(deckid)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, RemoteAPIError)
		return
	}

	cards, _:= deckToCardList(d)

	prices, err:= priceDB.GetBulkLatestHighest(aService.pool, cards, sourceName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, PriceDBError)
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(prices)
}

func (aService *PriceService) getDeckWeeklyLowest(req *restful.Request,
	resp *restful.Response) {
	
	deckid:= req.PathParameter("deckID")

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	// Grab the decklist from the remote
	d, err:= fetchRemoteDecklist(deckid)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, RemoteAPIError)
		return
	}

	cards, multipliers:= deckToCardList(d)

	weeks, err:= priceDB.GetBulkWeeklyLowest(aService.pool,
		cards, multipliers,
		sourceName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, PriceDBError)
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(weeks)
}

func (aService *PriceService) getDeckWeeklyHighest(req *restful.Request,
	resp *restful.Response) {
	
	deckid:= req.PathParameter("deckID")

	sourceName:= req.QueryParameter("source")
	if !validPriceSources[sourceName] {
		sourceName = DefaultPriceSource
	}

	// Grab the decklist from the remote
	d, err:= fetchRemoteDecklist(deckid)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, RemoteAPIError)
		return
	}

	cards, multipliers:= deckToCardList(d)

	weeks, err:= priceDB.GetBulkWeeklyHighest(aService.pool,
		cards, multipliers,
		sourceName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, PriceDBError)
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(weeks)
}


// Convert a deck to a list of cards and a corresponding
// list of integers for appearance numbers.
func deckToCardList(d *deckData.Deck) ([]string, []int32) {
	
	cards:= make([]string, 0)
	quantities:= make([]int32, 0)
	for _, c:= range d.Maindeck{
		cards = append(cards, c.Name)
		quantities = append(quantities, int32(c.Quantity))
	}
	for _, c:= range d.Sideboard{
		cards = append(cards, c.Name)
		quantities = append(quantities, int32(c.Quantity))
	}

	return cards, quantities
}

// Given a deckid, acquire the decklist from a remote database
func fetchRemoteDecklist(deckid string) (*deckData.Deck, error) {
	
	// Figure out which local port we can find
	// our Deck api on
	decksPort:= os.Getenv("DECK_API")	
	if len(decksPort) == 0 {
		return nil, fmt.Errorf("failed to get remote Deck api port")
	}

	// Basic sanity test of deckid
	if len(deckid) > 20 || len(deckid) == 0 {
		return nil, fmt.Errorf("bad deckid")
	}

	loc:= fmt.Sprintf("http://127.0.0.1:%s/api/Decks/Deck/%s",
		decksPort, deckid)

	resp, err:= http.Get(loc)
	if err!=nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err:= ioutil.ReadAll(resp.Body)
	if err!=nil {
		return nil, err
	}

	var d deckData.Deck
	err = json.Unmarshal(raw, &d)
	if err!=nil {
		return nil, err
	}

	return &d, nil

}