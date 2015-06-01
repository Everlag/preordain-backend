package ApiServices

import(

	"./../../../utilities/userDBHandler"
	"./../../../utilities/mailer"

	"github.com/emicklei/go-restful"

	"net/http"

)

// The contents of a reset email formatted to match the template.
type resetEmailContents struct{
	Name, ResetCode string
}

// Creates a user after validating the password. The remote database
// should prevent duplicates
func (aService *UserService) createUser(req *restful.Request,
	resp *restful.Response) {

	userName:= req.PathParameter("userName")
	var someUserData NewUserData
	err:= req.ReadEntity(&someUserData)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	if (someUserData.Nonce == 0){
		resp.WriteErrorString(http.StatusBadRequest, BadCaptcha)
		return
	}

	if !passwordMeetsRequirements(someUserData.Password) {
		resp.WriteErrorString(http.StatusBadRequest, BadPassword)
		return
	}

	sessionKey, err:= userDB.AddUser(aService.pool,
		userName, someUserData.Email,
		someUserData.Password)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, SignupFailure)
		return
	}

	resp.WriteEntity(sessionKey)

}

// Attempts to log the user in. Returns a valid session key
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

	sessionKey, err:= userDB.Login(aService.pool, userName, password)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(sessionKey)

}

// Requests that a valid reset token be created, recorded, and sent to the user's email.
//
// Sends mail to the user via the service embedded mailer
func (aService *UserService) requestPasswordReset(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")

	var resetRequestContainer PasswordResetRequestBody
	err:= req.ReadEntity(&resetRequestContainer)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	if (resetRequestContainer.Nonce == 0){
		resp.WriteErrorString(http.StatusBadRequest, BadCaptcha)
		return
	}

	code, err:= userDB.RequestReset(aService.pool, userName) 
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	// Fetch the user so we know their email
	u, err:= userDB.GetUser(aService.pool, userName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	contents:= resetEmailContents{
		Name: userName,
		ResetCode: code,
	}
	targetAddress:= mailer.FormatAddress(userName, u.Email)
	aService.mailer.SendPrepared("reset", contents,
		targetAddress, "Password Reset - Preorda.in")


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

	err = userDB.ChangePassword(aService.pool,
		userName, resetContainer.Password,
		resetContainer.ResetRequestToken)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(true)

}