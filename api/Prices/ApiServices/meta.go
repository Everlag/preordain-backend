package ApiServices

import(

	"net/http"
	"github.com/emicklei/go-restful"

)

// Register non-price metadata endpoints
func (aService *PriceService) registerMeta() {
	
	priceService:= aService.Service

	priceService.Route(priceService.
		GET("/SetList").To(aService.getSetList).
		// Docs
		Doc("All available sets").
		Operation("getSetList").
		Writes([]string{}).
		Returns(http.StatusOK, "All available sets", nil))

	priceService.Route(priceService.
		GET("/SetToShort").To(aService.getSetToShortMap).
		// Docs
		Doc("Map from set name to short code").
		Operation("getSetToShortMap").
		Writes(map[string]string{}).
		Returns(http.StatusOK, "Set names to short codes", nil))

	priceService.Route(priceService.
		GET("/SourceList").To(aService.getPriceSourcesList).
		// Docs
		Doc("All available price sources").
		Operation("getPriceSourcesList").
		Writes([]string{}).
		Returns(http.StatusOK, "All available sources", nil))

}

func (aService *PriceService) getSetList(req *restful.Request,
	resp *restful.Response) {
	
	setList:= make([]string, 0)
	for aSet, _:= range sets{
		if aSet != "" {
			setList = append(setList, aSet)	
		}
	}

	setCacheHeader(resp)

	resp.WriteEntity(setList)

}

func (aService *PriceService) getSetToShortMap(req *restful.Request,
	resp *restful.Response) {

	setCacheHeader(resp)

	resp.WriteEntity(setsToShort)

}

func (aService *PriceService) getPriceSourcesList(req *restful.Request,
	resp *restful.Response) {
	
	//
	sourcesList:= make([]string, 0)
	for aSource, _:= range validPriceSources{
		if aSource != "" {
			sourcesList = append(sourcesList, aSource)	
		}
	}

	setCacheHeader(resp)

	resp.WriteEntity(sourcesList)

}