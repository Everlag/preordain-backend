package ApiServices

import(

	"github.com/emicklei/go-restful"
	"net/http"

	"strings"
	"sort"

)

const BadKey string = "No suggestions available"

// How many characters into a word we work to
const TypeAheadDepth int = 20

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
	// Populate the cardsmap with unsorted arrays indexing
	//into TypeAheadDepth characters
	addMap(cardsMap, cards)
	addMap(cardsMap, sets)

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
		Path("/api/TypeAhead").
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

// Acquire suggestions for a given string
func (aService *TypeAheadService) getSuggestions(req *restful.Request,
	resp *restful.Response) {
	
	key:= strings.ToLower(req.PathParameter("key"))

	suggestions, ok:= aService.cards[key]
	if !ok{
		// Return an empty list of suggestions rather than 404ing.
		suggestions = []string{}
	}

	setCacheHeader(resp)
	resp.WriteEntity(suggestions)

}

func addMap(targetMap map[string][]string, names map[string]bool) {
	var key string
	for aName:= range names{

		aName = strings.Replace(aName, "Æ", "AE", -1)
		aLowerName:= strings.ToLower(aName)

		// Develop subarrays for each depth of key
		for keyIndexEnd := 1; keyIndexEnd < (TypeAheadDepth + 1); keyIndexEnd++ {
			
			if keyIndexEnd > len(aName) {
				break
			}
			key = aLowerName[0:keyIndexEnd]

			_, ok:= targetMap[key]
			if !ok {
				targetMap[key] = make([]string, 0)
			}

			targetMap[key] = append(targetMap[key], aName)

		}

	}
}