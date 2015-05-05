package main

import(

	"net/http"
	"github.com/emicklei/go-restful"

	"./ApiServices"

	"fmt"
)

func main() {

	priceService:= ApiServices.NewPriceService()

	restful.Add(priceService.Service)

	// Ensure we aren't sending stack traces out in the event we panic.
	restful.DefaultContainer.RecoverHandler(ApiServices.RecoverHandler)

	fmt.Println("goPrices price server ready")

	http.ListenAndServe(":9032", nil)
	
}