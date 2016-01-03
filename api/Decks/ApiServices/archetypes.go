package ApiServices

import(

	"net/http"
	"github.com/emicklei/go-restful"

	"./../../../common/deckDB"
	"./../../../common/deckDB/deckData"
	"./../../../common/deckDB/nameNorm"

)


// Register deck data for entire archetypes
func (s *DeckService) registerArchetype() {
	
	server:= s.Service


	server.Route(server.
		GET("/Archetype/{archetypeName}/Contents").To(s.getArchetypeContents).
		// Docs
		Doc("Cards which have appeared in an archetype").
		Operation("getArchetypeContents").
		Param(server.PathParameter("archetypeName",
			"Name of a Modern archetype we support").DataType("string")).
		Writes([]*deckData.Card{}).
		Returns(http.StatusInternalServerError, deckDBError, nil).
		Returns(http.StatusBadRequest, BadArchetype, nil).
		Returns(http.StatusOK, "Cards and number of appearances", nil))

	server.Route(server.
		GET("/Archetype/{archetypeName}/Latest").To(s.getArchetypeLatest).
		// Docs
		Doc("Latest deck to appear in a archetype").
		Operation("getArchetypeLatest").
		Param(server.PathParameter("archetypeName",
			"Name of a Modern archetype we support").DataType("string")).
		Writes(deckData.TaggedDeck{}).
		Returns(http.StatusInternalServerError, deckDBError, nil).
		Returns(http.StatusBadRequest, BadArchetype, nil).
		Returns(http.StatusOK, "Tagged decklist with metadata", deckData.TaggedDeck{}))
}

func (s *DeckService) getArchetypeContents(req *restful.Request,
	resp *restful.Response) {

	// Fetch and ensure this is a valid archetype
	a:= req.PathParameter("archetypeName")
	if !nameNorm.Valid(a) {
		resp.WriteErrorString(http.StatusBadRequest, BadArchetype)
		return
	}

	cards, err:= deckDB.GetArchetypeContents(s.pool, a)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, deckDBError)
		return
	}

	resp.WriteEntity(cards)

}

func (s *DeckService) getArchetypeLatest(req *restful.Request,
	resp *restful.Response) {

	a:= req.PathParameter("archetypeName")
	if !nameNorm.Valid(a) {
		resp.WriteErrorString(http.StatusBadRequest, BadArchetype)
		return
	}

	
	cards, err:= deckDB.GetArchetypeLatest(s.pool, a)
	if err!=nil {
		resp.WriteErrorString(http.StatusInternalServerError, deckDBError)
		return
	}

	resp.WriteEntity(cards)

}