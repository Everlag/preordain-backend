package ApiServices

import(

	"./userDBHandler"

	"github.com/emicklei/go-restful"

	"net/http"

	"fmt"

)

// Acquires the complete collection for a user
func (aService *UserService) getCollection(req *restful.Request,
	resp *restful.Response) {

	userName, sessionKey, err:= getUserNameAndSessionKey(req)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}
	collectionName:= req.PathParameter("collectionName")
	
	if sessionKey == nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	history, err:= userDB.GetCollectionHistory(aService.pool,
		sessionKey, userName, collectionName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	current, err:= userDB.GetCollectionContents(aService.pool,
		sessionKey, userName, collectionName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	aColl:= CollectionContents{
		Current: current,
		Historical: history,
	}

	resp.WriteEntity(aColl)

}

// Acquires a collection if and only if it is publicly available to view.
func (aService *UserService) getCollectionPublic(req *restful.Request,
	resp *restful.Response) {

	userName:= req.PathParameter("userName")
	collectionName:= req.PathParameter("collectionName")
	
	meta, err:= userDB.GetCollectionMeta(aService.pool,
		nil, userName, collectionName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	fmt.Println(meta)

	if meta.Privacy == "Private" {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return	
	}

	var history []userDB.Card
	if meta.Privacy == "History" {
		history, err = userDB.GetCollectionHistory(aService.pool,
		nil, userName, collectionName)
		if err!=nil {
			resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
			return
		}	
	}

	fmt.Println(history)

	current, err:= userDB.GetCollectionContents(aService.pool,
		nil, userName, collectionName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	aColl:= CollectionContents{
		Current: current,
		Historical: history,
	}

	resp.WriteEntity(aColl)

}

// Add a transaction to the user.
//
// Updates the historical use of a collection alongside its current
// contents.
func (aService *UserService) addTrade(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")
	collectionName:= req.PathParameter("collectionName")

	var tradeContainer TradeAddBody
	err:= req.ReadEntity(&tradeContainer)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	if tradeContainer.SessionKey == nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	// Ensure we have received a trade consisting of valid Magic cards
	// inside their specific sets
	for _, aCard:= range tradeContainer.Trade{
		validSets, validCard:= cardsToSets[aCard.Name]
		if !validCard {
			resp.WriteErrorString(http.StatusBadRequest, BadTradeContents)
			return
		}
		_, validSet:= validSets[aCard.Set]
		if !validSet {
			resp.WriteErrorString(http.StatusBadRequest, BadTradeContents)
			return
		}

	}

	err = userDB.AddCards(aService.pool,
		tradeContainer.SessionKey,
		userName, collectionName,
		tradeContainer.Trade)
	if err!=nil {
		aService.logger.Println(err)
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(true)

}