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

	cards:= deckToCardList(d)

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

	cards:= deckToCardList(d)

	prices, err:= priceDB.GetBulkLatestHighest(aService.pool, cards, sourceName)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, PriceDBError)
		return
	}

	// Set cache header to reduce load.
	setCacheHeader(resp)

	resp.WriteEntity(prices)
}

// Convert a deck to a list of cards
func deckToCardList(d *deckData.Deck) []string {
	
	cards:= make([]string, 0)
	for _, c:= range d.Maindeck{
		cards = append(cards, c.Name)
	}
	for _, c:= range d.Sideboard{
		cards = append(cards, c.Name)
	}

	return cards
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