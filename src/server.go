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
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "swaggerUI",
	}
	swagger.InstallSwaggerService(swaggerConfig)

}

func main() {

	priceService:= ApiServices.NewPriceService()
	userService:= ApiServices.NewUserService()

	restful.Add(priceService.Service)
	restful.Add(userService.Service)
	// BUG - user service breaks when compression is enabled.
	//restful.DefaultContainer.EnableContentEncoding(true)

	setupSwagger()

	fmt.Println("ready")

	http.ListenAndServe(":9032", nil)
}