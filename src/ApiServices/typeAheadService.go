package ApiServices

import(

	"github.com/emicklei/go-restful"
	"net/http"

	"strings"
	"sort"

)

const BadKey string = "No suggestions available"

// How many characters into a word we work to
const TypeAheadDepth int = 10

type TypeAheadService struct{

	// Mapping from starting character to an alphabetically sorted
	// array of strings containing that character 
	cards map[string][]string
	Service *restful.WebService
}

func NewTypeAheadService() (*TypeAheadService) {
	

	typeAheadLogger:= GetLogger("typeAheadLogger.txt", "typeAheadLog")

	// Ensures we have a valid filter for card names/sets
	// to work off of.
	err:= populateCardMaps()
	if err!=nil {
		typeAheadLogger.Fatalln("Failed to acquire ")
	}

	cardsMap:= make(map[string][]string)
	// Populate the cardsmap with unsorted arrays indexing into two characters
	var key string
	for aCardName:= range cards{

		aCardName = strings.Replace(aCardName, "Æ", "AE", -1)
		aLowerName:= strings.ToLower(aCardName)

		// Develop subarrays for each depth of key
		for keyIndexEnd := 1; keyIndexEnd < (TypeAheadDepth + 1); keyIndexEnd++ {
			
			if keyIndexEnd > len(aCardName) {
				break
			}
			key = aLowerName[0:keyIndexEnd]

			_, ok:= cardsMap[key]
			if !ok {
				cardsMap[key] = make([]string, 0)
			}

			cardsMap[key] = append(cardsMap[key], aCardName)

		}

	}

	for aKey, _:= range cardsMap{
		sort.Strings(cardsMap[aKey])
	}

	aService:= TypeAheadService{
		cards: cardsMap,
	}

	aService.register()

	return &aService

}

func (aService *TypeAheadService) register() {

	typeAheadService:= new(restful.WebService)
	typeAheadService.
		Path("/TypeAhead").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	typeAheadService.Route(typeAheadService.
		GET("/{key}").
		To(aService.getSuggestions).
		// Docs
		Doc("Attempts to match the provided key to a preset group of suggestions.\nCase Insensitive").
		Operation("getSuggestions").
		Param(typeAheadService.PathParameter("key",
			"The first characters of matching strings we return").DataType("string")).
		Returns(http.StatusBadRequest, BadKey, nil).
		Writes([]string{"aCardName", "anotherCardName"}).
		Returns(http.StatusOK, "Suggestions as the payload", nil))

	aService.Service = typeAheadService
}

func (aService *TypeAheadService) getSuggestions(req *restful.Request,
	resp *restful.Response) {
	
	key:= strings.ToLower(req.PathParameter("key"))

	suggestions, ok:= aService.cards[key]
	if !ok{
		resp.WriteErrorString(http.StatusBadRequest, BadKey)
		return
	}

	resp.WriteEntity(suggestions)

}