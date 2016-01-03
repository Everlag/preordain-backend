package ApiServices

import(

	"net/http"
	"github.com/emicklei/go-restful"

	"./../../../common/deckDB/nameNorm"

)

// Register deck data for entire archetypes
func (s *DeckService) registerMeta() {
	
	server:= s.Service


	server.Route(server.
		GET("/Archetypes").To(s.getArchetypeList).
		// Docs
		Doc("All available archetypes we support").
		Operation("getArchetypeList").
		Writes([]string{}).
		Returns(http.StatusOK, "All available archetypes", []string{}))
}

func (s *DeckService) getArchetypeList(req *restful.Request,
	resp *restful.Response) {

	resp.WriteEntity(nameNorm.Names())

}