package main

import(

	"net/http"
	"github.com/emicklei/go-restful"

	"./ApiServices"

	"fmt"
)

func main() {

	userService:= ApiServices.NewUserService()

	restful.Add(userService.Service)

	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"Access-Control-Allow-Origin"},
		AllowedHeaders: []string{"Content-Type"},
		CookiesAllowed: false,
		Container:      restful.DefaultContainer}
	restful.DefaultContainer.Filter(cors.Filter)

	// Ensure we aren't sending stack traces out in the event we panic.
	restful.DefaultContainer.RecoverHandler(ApiServices.RecoverHandler)

	fmt.Println("goPrices user server ready")

	http.ListenAndServe(":9032", nil)
	
}