package ApiServices

import(

	"github.com/emicklei/go-restful"

	"./../../../common/deckDB"
	"github.com/jackc/pgx"

	"log"
)

// Unpleasant Responses
const deckDBError string = "Deck DB lookup failed"
const BadArchetype string = "Unknown Archetype"

type DeckService struct{
	pool *pgx.ConnPool
	Service *restful.WebService
	logger *log.Logger
}

// Returns a fresh DeckService ready to be hooked up to restful
func NewPriceService() *DeckService {

	deckLogger:= GetLogger("deckLogger.txt", "deckLog")

	// Connect to the remote priceDB
	pool, err:= deckDB.Connect()
	if err!=nil {
		deckLogger.Fatalln("failed to acquire deckDB client", err)
	}
	
	s:= DeckService{
		pool: pool,
		logger: deckLogger,
	}

	// Register everything
	err = s.register()
	if err!=nil {
		deckLogger.Fatalln("failed to register DeckService, ", err)
	}

	return &s

}

func (s *DeckService) register() error {
	
	server:= new(restful.WebService)
	server.
		Path("/api/Decks").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		ApiVersion("0.1")

	s.Service = server

	s.registerMeta()
	s.registerDeck()
	s.registerArchetype()

	return nil
}