package ApiServices

import(

	"./../../../utilities/userDBHandler"

	"github.com/emicklei/go-restful"

	"net/http"

)

// Creates a new collection for the named user
func (aService *UserService) newCollection(req *restful.Request,
	resp *restful.Response)  {

	userName, sessionKey, err:= getUserNameAndSessionKey(req)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}
	collectionName:= req.PathParameter("collectionName")

	if sessionKey == nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	err = userDB.AddCollection(aService.pool,
		sessionKey,
		userName, collectionName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(true)

}

// Set the viewing levels for a collection under a user
func (aService *UserService) setCollectionPermissions(req *restful.Request,
	resp *restful.Response) {

	userName:= req.PathParameter("userName")
	collectionName:= req.PathParameter("collectionName")

	var permissionsContainer PermissionChangeBody
	err:= req.ReadEntity(&permissionsContainer)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	if permissionsContainer.SessionKey == nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}
	
	err = userDB.SetCollectionPrivacy(aService.pool,
		permissionsContainer.SessionKey,
		userName, collectionName,
		permissionsContainer.Privacy)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(true)

}

// Get the viewing levels for a collection under a user
func (aService *UserService) getCollectionPermissions(req *restful.Request,
	resp *restful.Response) {

	var userName string
	collectionName:= req.PathParameter("collectionName")

	userName, sessionKey, err:= getUserNameAndSessionKey(req)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	if sessionKey == nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}
	
	meta, err:= userDB.GetCollectionMeta(aService.pool,
		sessionKey,
		userName, collectionName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(meta.Privacy)

}

// Acquire all non-private collections for a named user.
func (aService *UserService) getUserPublicCollections(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")

	collections, err:= userDB.GetCollectionList(aService.pool, userName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	public:= make([]string, 0)
	for _, c:= range collections{
		if c.Privacy != "Private" {
			public = append(public, c.Name)
		}
	}

	resp.WriteEntity(public)

}

// Acquire all collections for a named and authenticated user
func (aService *UserService) getUserCollections(req *restful.Request,
	resp *restful.Response) {

	userName, sessionKey, err:= getUserNameAndSessionKey(req)
	if err!=nil {
		aService.logger.Println(err)
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	if sessionKey == nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	err = userDB.SessionAuth(aService.pool, userName, sessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusUnauthorized, BadCredentials)
		return
	}

	collections, err:= userDB.GetCollectionList(aService.pool, userName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	all:= make([]string, 0)
	for _, c:= range collections{
		all = append(all, c.Name)
	}

	resp.WriteEntity(all)

}