package main

import(

	"net/http"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"

	"./ApiServices"

	"fmt"
)

func setupSwagger() {
	
	swaggerConfig:= swagger.Config{
		WebServices:    restful.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: "http://localhost:9032",
		ApiPath:        "/api/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/api/apidocs/",
		SwaggerFilePath: "swaggerUI",
	}
	swagger.InstallSwaggerService(swaggerConfig)

}

func setupCors() {
	cors:= restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"X-My-Header"},
		AllowedHeaders: []string{"Content-Type"},
		CookiesAllowed: false,
		Container:       restful.DefaultContainer}
	restful.DefaultContainer.Filter(cors.Filter)
}

func main() {

	priceService:= ApiServices.NewPriceService()
	userService:= ApiServices.NewUserService()
	typeAheadService:= ApiServices.NewTypeAheadService()

	restful.Add(priceService.Service)
	restful.Add(userService.Service)
	restful.Add(typeAheadService.Service)
	// BUG - user service breaks when compression is enabled.
	//restful.DefaultContainer.EnableContentEncoding(true)

	setupSwagger()
	setupCors()

	fmt.Println("ready")

	http.ListenAndServe(":9032", nil)
	
}