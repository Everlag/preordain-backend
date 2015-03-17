package ApiServices

import(

	"./../goPrices/UserStructs"
	"./../goPrices/UserLogger"

	"github.com/emicklei/go-restful"

	"net/http"
	"log"
)

const BadUserName string = "User lookup failed"
const BadSessionKey string = "Invalid Session Key"
const BadCredentials string = "Invalid Credentials"
const BadCaptcha string = "Invalid Re-Captcha"
const BadTradeContents string = "Invalid trade contents"

const SignupFailure string = "Failed to create user"
const BodyReadFailure string = "Failed to parse body parameter"

type UserService struct{

	manager *UserStructs.UserManager
	Service *restful.WebService
	logger *log.Logger
	actionLogger *UserLogger.Logger

}

// Returns a fresh UserService ready to be hooked up to restful
func NewUserService() *UserService {
	
	// Get necessary loggers
	userLogger:= GetLogger("userLogger.txt", "userLog")
	userActionLogger, err:= UserLogger.NewLogger()
	if err!=nil {
		userLogger.Fatalln("Failed to acquire userActionLogger, ", err)
	}

	// Get the metadata we need
	name, err:= getManagerName()
	if err!=nil {
		userLogger.Fatalln("Failed to acquire manager name, ", err)
	}

	// Acquire the manager
	aManager, err:= UserStructs.ReacquireManager(name)
	if err!=nil {
		userLogger.Fatalln("Failed to acquire user manager ", err)
	}

	aService:= UserService{
		manager: aManager,
		logger: userLogger,
		actionLogger: userActionLogger,
	}

	// Acquire and set up recaptcha
	err = setupRecaptcha()
	if err!=nil {
		userLogger.Fatalln("Failed to setup recaptcha, ", err)
	}

	// Finally, register the service
	err = aService.register()
	if err!=nil {
		userLogger.Fatalln("Failed to register UserService, ", err)
	}

	return &aService

}

func (aService *UserService) logAction(userName, action string,
	actionParameters map[string]string, bodyContents interface{}) {
	
	err:= aService.actionLogger.WriteAction(userName, action,
		actionParameters, bodyContents)
	if err!=nil {
		aService.logger.Println("Failed to log user action, err is ", err,
			", action is ",
			userName, action,
			actionParameters, bodyContents)
	}

}

func (aService *UserService) register() error {
	
	// Ensures we have a valid filter for card names/sets
	//
	// Other services may do this but better to take an extra .1s at
	// startup than to risk nuking every attempt at adding a trade.
	err:= populateCardMaps()
	if err!=nil {
		aService.logger.Fatalln("Failed to acquire ")
	}

	userService:= new(restful.WebService)
	userService.
		Path("/api/Users").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// Extremely gross code, which does documents itself
	// in an externally packaged pretty ui, follows.

	userService.Route(userService.
		POST("/{userName}").To(aService.createUser).
		// Docs
		Doc("Attempts to create a user").
		Operation("createUser").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Reads(NewUserData{}).
		Writes("string").
		Returns(http.StatusBadRequest, SignupFailure, nil).
		Returns(http.StatusBadRequest, BadCaptcha, nil).
		Returns(http.StatusOK, "A valid session code for the user", nil))

	userService.Route(userService.
		POST("/{userName}/Login").To(aService.loginUser).
		// Docs
		Doc("Attempts to login a user").
		Operation("loginUser").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Reads(PasswordBody{}).
		Returns(http.StatusBadRequest, SignupFailure, nil).
		Returns(http.StatusBadRequest, BadCaptcha, nil).
		Returns(http.StatusOK, "A valid session code for the user", nil))

	userService.Route(userService.
		GET("/{userName}/Collections/GetPublic").To(aService.getUserPublicCollections).
		// Docs
		Doc("Returns a list of collections designated public by that user").
		Operation("getUserPublicCollections").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Writes([]string{}).
		Returns(http.StatusBadRequest, BadUserName, nil).
		Returns(http.StatusOK, "Public collections for a specified user", nil))

	userService.Route(userService.
		POST("/{userName}/Collections/Get").To(aService.getUserCollections).
		// Docs
		Doc("Returns a list of collections an authenticated user").
		Operation("getUserCollections").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Reads(SessionKeyBody{}).
		Writes([]string{}).
		Returns(http.StatusBadRequest, BodyReadFailure, nil).
		Returns(http.StatusUnauthorized, BadCredentials, nil).
		Returns(http.StatusOK, "Collections for a specified user", nil))

	userService.Route(userService.
		POST("/{userName}/Collections/{collectionName}/Create").
		To(aService.newCollection).
		// Docs
		Doc("Adds a new collection with the given name to the user").
		Operation("getUserCollections").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Param(userService.PathParameter("collectionName",
			"The name of a collection for that user").DataType("string")).
		Reads(SessionKeyBody{}).
		Writes(true).
		Returns(http.StatusBadRequest, BodyReadFailure, nil).
		Returns(http.StatusUnauthorized, BadCredentials, nil).
		Returns(http.StatusOK, "Collection is added", nil))

	userService.Route(userService.
		POST("/{userName}/Collections/{collectionName}/Get").
		To(aService.getCollection).
		// Docs
		Doc("Attempts to retrieve a collection from an authenticated user").
		Operation("getUserCollections").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Param(userService.PathParameter("collectionName",
			"The name of a collection for that user").DataType("string")).
		Reads(SessionKeyBody{}).
		Writes(UserStructs.Collection{}).
		Returns(http.StatusBadRequest, BodyReadFailure, nil).
		Returns(http.StatusUnauthorized, BadCredentials, nil).
		Returns(http.StatusOK, "Collection is returned", nil))

	userService.Route(userService.
		GET("/{userName}/Collections/{collectionName}/GetPublic").
		To(aService.getCollectionPublic).
		// Docs
		Doc("Attempts to read a public collection for a user").
		Operation("getUserCollections").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Param(userService.PathParameter("collectionName",
			"The name of a collection for that user").DataType("string")).
		Writes(UserStructs.Collection{}).
		Returns(http.StatusBadRequest, BodyReadFailure, nil).
		Returns(http.StatusUnauthorized, BadCredentials, nil).
		Returns(http.StatusOK, "Collection is returned", nil))

	userService.Route(userService.
		PATCH("/{userName}/Collections/{collectionName}/Permissions").
		To(aService.setCollectionPermissions).
		// Docs
		Doc("Attempt to change public viewing permissions for a collection").
		Operation("getUserCollections").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Param(userService.PathParameter("collectionName",
			"The name of a collection for that user").DataType("string")).
		Reads(PermissionChangeBody{}).
		Returns(http.StatusBadRequest, BodyReadFailure, nil).
		Writes(true).
		Returns(http.StatusUnauthorized, BadCredentials, nil).
		Returns(http.StatusOK, "Permissions changed", nil))

	userService.Route(userService.
		POST("/{userName}/Collections/{collectionName}/Trades").
		To(aService.addTrade).
		// Docs
		Doc("Attempt to add a provided trade to a collection").
		Operation("addTrade").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Param(userService.PathParameter("collectionName",
			"The name of a collection for that user").DataType("string")).
		Reads(TradeAddBody{}).
		Returns(http.StatusBadRequest, BodyReadFailure, nil).
		Writes(true).
		Returns(http.StatusUnauthorized, BadCredentials, nil).
		Returns(http.StatusOK, "Trade Added", nil))

	userService.Route(userService.
		POST("/{userName}/PasswordResetRequest").
		To(aService.requestPasswordReset).
		// Docs
		Doc("Attempts to get a password reset email sent to the user's email").
		Operation("passwordResetRequest").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Reads(PasswordResetRequestBody{}).
		Returns(http.StatusBadRequest, BodyReadFailure, nil).
		Returns(http.StatusBadRequest, BadCaptcha, nil).
		Returns(http.StatusUnauthorized, BadCredentials, nil).
		Writes(true).
		Returns(http.StatusOK, "Reset Code Sent", nil))

	userService.Route(userService.
		POST("/{userName}/PasswordReset").
		To(aService.resetPassword).
		// Docs
		Doc("Attempts to change the user's password using a token").
		Operation("passwordReset").
		Param(userService.PathParameter("userName",
			"The name that identifies a user to our service").DataType("string")).
		Reads(PasswordResetBody{}).
		Returns(http.StatusBadRequest, BodyReadFailure, nil).
		Returns(http.StatusBadRequest, BadCaptcha, nil).
		Returns(http.StatusUnauthorized, BadCredentials, nil).
		Writes("A Valid Session Key").
		Returns(http.StatusOK, "Successfully reset", nil))


	aService.Service = userService

	return nil
}