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
	envError:= godotenv.Load("prices.default.env")
	if envError!=nil {
		fmt.Println("failed to parse prices.default.env")
		os.Exit(1)
	}

	priceService:= ApiServices.NewPriceService()

	restful.Add(priceService.Service)

	// Expose docs json
	//
	// Developer note: Documentation endpoints are
	// called against the production servers. Modified
	// or added endpoints will have undefined behaviour.
	config := swagger.Config{
		WebServices:    restful.DefaultContainer.RegisteredWebServices(),
		WebServicesUrl: "https://preorda.in/backend",
		ApiPath:        "api/Prices/apidocs.json",
	}
	swagger.RegisterSwaggerService(config, restful.DefaultContainer)

	// Ensure we aren't sending stack traces out in the event we panic.
	restful.DefaultContainer.RecoverHandler(ApiServices.RecoverHandler)

	fmt.Println("goPrices price server ready")

	http.ListenAndServe(":9032", nil)
	
}