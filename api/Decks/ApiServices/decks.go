package ApiServices

import(

	"net/http"
	"github.com/emicklei/go-restful"

	"./../../../common/deckDB"
	"./../../../common/deckDB/deckData"

	"fmt"

)


// Register deck data for entire archetypes
func (s *DeckService) registerDeck() {
	
	server:= s.Service


	server.Route(server.
		GET("/Deck/{deckID}").To(s.getDeck).
		// Docs
		Doc("Full decklist for a provided deckid").
		Operation("getDeck").
		Param(server.PathParameter("deckID",
			"A valid deck identifer").DataType("string")).
		Writes(deckData.Deck{}).
		Returns(http.StatusInternalServerError, deckDBError, nil).
		Returns(http.StatusBadRequest, BadArchetype, nil).
		Returns(http.StatusOK, "Deck object", deckData.Deck{}))
}

func (s *DeckService) getDeck(req *restful.Request,
	resp *restful.Response) {

	// Fetch and ensure this is a valid archetype
	deckid:= req.PathParameter("deckID")

	d, err:= deckDB.GetDeck(s.pool, deckid)
	if err!=nil {
		fmt.Println(err)
		resp.WriteErrorString(http.StatusInternalServerError, deckDBError)
		return
	}

	resp.WriteEntity(d)

}