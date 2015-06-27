package ApiServices


import(

	"github.com/emicklei/go-restful"

	"./../../../utilities/priceDBHandler.v2"
	"github.com/jackc/pgx"

	"log"

)
// Defaults
const DefaultPriceSource string = "mtgprice"

// Responses
const PriceDBError string = "Price DB lookup failed"
const BadCard string = "Illegal Card Name"
const BadSet string = "Illegal Set Name"
const BadTime string = "Illegible time"
const BadCardFilter string = BadCard + " || " + BadSet
const BadCalculation string = "Failed Calculation"

// Which sources we currently support specific queries for
var validPriceSources = make(map[string]bool)

type PriceService struct{
	pool *pgx.ConnPool
	Service *restful.WebService
	logger *log.Logger
}

// Returns a fresh PriceService ready to be hooked up to restful
func NewPriceService() *PriceService {

	priceLogger:= GetLogger("priceLogger.txt", "priceLog")

	// Connect to the remote priceDB
	pool, err:= priceDB.Connect()
	if err!=nil {
		priceLogger.Fatalln("Failed to acquire priceDb client", err)
	}

	// Initialize our queryable sources
	for _, source:= range priceDB.Sources{
		validPriceSources[source] = true
	}
	
	aService:= PriceService{
		pool: pool,
		logger: priceLogger,
	}

	// Register everything
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
		ApiVersion("0.2")

	aService.Service = priceService

	// Register all of our necessary services
	aService.registerMeta()
	aService.registerHistorical()
	aService.registerLatest()
	aService.registerClosest()
	aService.registerSets()

	return nil

}