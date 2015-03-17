package ApiServices

import(

	"./../goPrices/UserStructs"

	"github.com/emicklei/go-restful"

	"net/http"

)

// While gross, having a struct for each request lets me keep this as a json
// based api without too much pain in go-restful
type NewUserData struct{

	Email, Password string
	RecaptchaChallengeField, RecaptchaResponseField string

}

type PasswordBody struct{
	Password string
}

type SessionKeyBody struct{
	SessionKey string
}

type PermissionChangeBody struct{
	SessionKey string
	Viewing, History, Comments bool
}

type TradeAddBody struct{

	Trade UserStructs.Trade
	SessionKey string

}

type PasswordResetRequestBody struct{

	RecaptchaChallengeField, RecaptchaResponseField string

}

type PasswordResetBody struct{

	Password string
	ResetRequestToken string

}

func (aService *UserService) createUser(req *restful.Request,
	resp *restful.Response) {

	userName:= req.PathParameter("userName")
	var someUserData NewUserData
	err:= req.ReadEntity(&someUserData)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	/* COMMENTED OUT FOR TESTING
	validCaptcha:= ValidateRecaptcha(req, someUserData.RecaptchaChallengeField,
		someUserData.RecaptchaResponseField)
	if !validCaptcha {
		resp.WriteErrorString(http.StatusBadRequest, BadCaptcha)
		return
	}
	*/

	sessionKey, err:= aService.manager.AddUser(userName,
		someUserData.Email,
		someUserData.Password,
		UserStructs.StandardCollectionCount)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, SignupFailure)
		return
	}
	
	// Log the successful user action to the remote db after removing their password
	// from the log
	someUserData.Password = ""
	aService.logAction(userName, req.SelectedRoutePath(),
		req.PathParameters(), someUserData)

	resp.WriteEntity(sessionKey)

}

func (aService *UserService) loginUser(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")
	var passwordContainer PasswordBody
	err:= req.ReadEntity(&passwordContainer)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	password:= passwordContainer.Password

	sessionKey, err:= aService.manager.GetNewSession(userName, password)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(sessionKey)

}

func (aService *UserService) newCollection(req *restful.Request,
	resp *restful.Response)  {

	userName, sessionKey, err:= getUserNameAndSessionKey(req)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}
	collectionName:= req.PathParameter("collectionName")

	err = aService.manager.NewCollection(userName, collectionName, sessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	aService.logAction(userName, req.SelectedRoutePath(),
		req.PathParameters(), nil)

	resp.WriteEntity(true)

}

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
	
	err = aService.manager.SetPermissions(userName, collectionName,
		permissionsContainer.SessionKey,
		permissionsContainer.Viewing,
		permissionsContainer.History,
		permissionsContainer.Comments)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	aService.logAction(userName, req.SelectedRoutePath(),
		req.PathParameters(), permissionsContainer)

	resp.WriteEntity(true)

}

func (aService *UserService) addTrade(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")
	collectionName:= req.PathParameter("collectionName")

	var tradeContainer TradeAddBody
	err:= req.ReadEntity(&tradeContainer)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	// Ensure we have received a trade consisting of valid Magic cards
	// inside their specific sets
	for _, aCard:= range tradeContainer.Trade.Transaction{
		validSets, validCard:= cardsToSets[aCard.Name]
		if !validCard {
			resp.WriteErrorString(http.StatusBadRequest, BadTradeContents)
			return
		}
		_, validSet:= validSets[aCard.Set]
		if !validSet {
			resp.WriteErrorString(http.StatusBadRequest, BadTradeContents)
			return
		}


	}

	err = aService.manager.AddTrade(userName, collectionName,
		tradeContainer.SessionKey,
		tradeContainer.Trade)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	aService.logAction(userName, req.SelectedRoutePath(),
		req.PathParameters(), tradeContainer)

	resp.WriteEntity(true)

}

func (aService *UserService) requestPasswordReset(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")

	var resetRequestContainer PasswordResetRequestBody
	err:= req.ReadEntity(&resetRequestContainer)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	/*
	validCaptcha:= ValidateRecaptcha(req, resetRequestContainer.RecaptchaChallengeField,
		resetRequestContainer.RecaptchaResponseField)
	if !validCaptcha {
		resp.WriteErrorString(http.StatusBadRequest, BadCaptcha)
		return
	}
	*/

	err = aService.manager.GetPasswordResetToken(userName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	aService.logAction(userName, req.SelectedRoutePath(),
		req.PathParameters(), resetRequestContainer)

	resp.WriteEntity(true)

}

func (aService *UserService) resetPassword(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")

	var resetContainer PasswordResetBody
	err:= req.ReadEntity(&resetContainer)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}


	newSessionKey, err:= aService.manager.ChangePassword(userName,
		resetContainer.ResetRequestToken,
		resetContainer.Password)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resetContainer.Password = ""
	aService.logAction(userName, req.SelectedRoutePath(),
		req.PathParameters(), resetContainer)

	resp.WriteEntity(newSessionKey)

}