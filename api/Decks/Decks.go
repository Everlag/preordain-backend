package main

import(

	"net/http"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"

	"./ApiServices"

	"fmt"

	"github.com/joho/godotenv"
	"os"
)

func main() {

	// Populate config locations not explicitly set
	envError:= godotenv.Load("decks.default.env")
	if envError!=nil {
		fmt.Println("failed to parse decks.default.env")
		os.Exit(1)
	}

	deckService:= ApiServices.NewPriceService()

	restful.Add(deckService.Service)

	// Expose docs json
	//
	// Developer note: Documentation endpoints are
	// called against the production servers. Modified
	// or added endpoints will have undefined behaviour.
	config := swagger.Config{
		WebServices:    restful.DefaultContainer.RegisteredWebServices(),
		WebServicesUrl: "https://preorda.in/backend",
		ApiPath:        "api/Decks/apidocs.json",
	}
	swagger.RegisterSwaggerService(config, restful.DefaultContainer)

	// Ensure we aren't sending stack traces out in the event we panic.
	restful.DefaultContainer.RecoverHandler(ApiServices.RecoverHandler)

	fmt.Println("goPrices deck server ready")

	http.ListenAndServe(":9037", nil)
	
}