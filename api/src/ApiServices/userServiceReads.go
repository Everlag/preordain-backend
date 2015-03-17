package ApiServices

import(

	"./../goPrices/UserStructs"

	"github.com/emicklei/go-restful"

	"net/http"

)


func (aService *UserService) getUserPublicCollections(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")

	collectionList, err:= aService.manager.GetCollectionList(userName,
		UserStructs.PublicSessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(collectionList)

}

func (aService *UserService) getUserCollections(req *restful.Request,
	resp *restful.Response) {

	userName, sessionKey, err:= getUserNameAndSessionKey(req)
	if err!=nil {
		aService.logger.Println(err)
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	collectionList, err:= aService.manager.GetCollectionList(userName,
		sessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusUnauthorized, BadCredentials)
		return
	}

	resp.WriteEntity(collectionList)

}

func (aService *UserService) getCollection(req *restful.Request,
	resp *restful.Response) {

	userName, sessionKey, err:= getUserNameAndSessionKey(req)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}
	collectionName:= req.PathParameter("collectionName")
	
	aColl, err:= aService.manager.GetCollection(userName, collectionName, sessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(*aColl)

}

func (aService *UserService) getCollectionPublic(req *restful.Request,
	resp *restful.Response) {

	userName:= req.PathParameter("userName")
	collectionName:= req.PathParameter("collectionName")
	
	aColl, err:= aService.manager.GetCollection(userName,
		collectionName, UserStructs.PublicSessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(*aColl)

}