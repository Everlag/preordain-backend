package ApiServices

import(

	"./../../../utilities/userDBHandler"

	"./../../../utilities/mailer"

	"github.com/emicklei/go-restful"

	"net/http"

)

type subEmailContents struct{
	Name, Plan string
}

// Changes the current status of the user's subscription to another
// plan
//
// By default all users are on the passive free plan.
func (aService *UserService) modSubUser(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")
	var subContainer SubBody
	err:= req.ReadEntity(&subContainer)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	// Check to ensure we actually got a valid session
	if subContainer.SessionKey == nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	// Grab the customer's identification.
	sub, err:= userDB.GetSub(aService.pool, userName, subContainer.SessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, DBfailure)
		return
	}
	// Make sure they've signed up before
	if sub.CustomerID == userDB.DefaultID ||
	sub.SubID == userDB.DefaultID {
		resp.WriteErrorString(http.StatusBadRequest, BadPlanChoice)
		return
	}

	// Make sure we aren't double charging them.
	validChoice, err:= userDB.DifferentPlan(aService.pool,
		userName, subContainer.Plan)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, DBfailure)
		return
	}
	if !validChoice {
		resp.WriteErrorString(http.StatusBadRequest, BadPlanChoice)
		return
	}

	// Change the customer's payment method to the one they just provided
	err = aService.merch.UpdateCustomer(sub.CustomerID,
		subContainer.PaymentMethod)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, StripeCustFailure)
		return
	}

	// Change the user's subscription status
	err = aService.merch.UpdateSubCustomer(sub.CustomerID, sub.SubID,
		subContainer.Plan)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, StripeSubFailure)
		return
	}

	// Retain everything regarding their subscription apart from the plan
	// and its various effects.
	//
	// Notice that we IGNORE the error as we want to send the user
	// the following email to ensure they can contact us if we fail
	// to change the DB but already have stripe charging them.
	_ = userDB.ModSub(aService.pool,
		userName, subContainer.Plan,
		sub.CustomerID, sub.SubID, nil)

	// Grab their email so we can let them know
	u, err:= userDB.GetUser(aService.pool, userName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, DBfailure)
		return
	}

	// Email them that we were successful!
	contents:= subEmailContents{
		Name: userName,
		Plan: subContainer.Plan,
	}
	targetAddress:= mailer.FormatAddress(userName, u.Email)
	err = aService.mailer.SendPrepared("subSuccess", contents,
		targetAddress, "Subscribed! - Preorda.in")
	if err!=nil {
		aService.logger.Println("failed to send email", err)
	}


	resp.WriteEntity(true)

}

// Adds a user to a paid subscription plan
//
// Yes, this makes WAY too many round trips to the database...
func (aService *UserService) addSubUser(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")
	var subContainer SubBody
	err:= req.ReadEntity(&subContainer)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	// Check to ensure we actually got a valid session
	if subContainer.SessionKey == nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	// Grab their email
	u, err:= userDB.GetUser(aService.pool, userName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, DBfailure)
		return
	}

	// Make sure we aren't double charging them.
	validChoice, err:= userDB.DifferentPlan(aService.pool,
		userName, subContainer.Plan)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, DBfailure)
		return
	}
	if !validChoice {
		resp.WriteErrorString(http.StatusBadRequest, BadPlanChoice)
		return
	}

	// Grab the customer's identification.
	sub, err:= userDB.GetSub(aService.pool, userName, subContainer.SessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, DBfailure)
		return
	}

	// Check if we need to add them as a customer
	custID:= sub.CustomerID
	if custID == userDB.DefaultID {
		// Add them as a customer as needed
		custID, err = aService.merch.AddCustomer(subContainer.PaymentMethod,
			u.Email, subContainer.Coupon)
		if err!=nil {
			resp.WriteErrorString(http.StatusBadRequest, StripeCustFailure)
			return
		}	
	}

	// Add them as a subscriber.
	subID, err:= aService.merch.SubCustomer(custID, subContainer.Plan)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, StripeSubFailure)
		return
	}

	// Now update their entries in the database.
	//
	// Notice that we IGNORE the error as we want to send the user
	// the following email to ensure they can contact us if we fail
	// to change the DB but already have stripe charging them.
	userDB.ModSub(aService.pool, userName, subContainer.Plan,
		custID, subID, subContainer.SessionKey)


	// Email them that we were successful!
	contents:= subEmailContents{
		Name: userName,
		Plan: subContainer.Plan,
	}
	targetAddress:= mailer.FormatAddress(userName, u.Email)
	err = aService.mailer.SendPrepared("subSuccess", contents,
		targetAddress, "Subscribed! - Preorda.in")
	if err!=nil {
		aService.logger.Println("failed to send email", err)
	}


	resp.WriteEntity(true)

}

// Sets a user's subscription to the 'free' plan.
//
// Free is defined as userDB.DefaultSubLevel
func (aService *UserService) unSubUser(req *restful.Request,
	resp *restful.Response) {
	
	userName:= req.PathParameter("userName")
	var subContainer SubBody
	err:= req.ReadEntity(&subContainer)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	// Check to ensure we actually got a valid session
	if subContainer.SessionKey == nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	// Grab the customer's identification.
	sub, err:= userDB.GetSub(aService.pool, userName, subContainer.SessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, DBfailure)
		return
	}

	// Remove their subscription but retain their customerID
	err = aService.merch.UnSubCustomer(sub.SubID, sub.CustomerID)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, StripeSubFailure)
		return
	}

	// Mod the sub to be the default free version.
	//
	// Set dummy sub ID but hold onto that customerID
	err = userDB.ModSub(aService.pool, userName, userDB.DefaultSubLevel,
		sub.CustomerID, userDB.DefaultID, subContainer.SessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, DBWriteFailure)
		return
	}

	// Grab their email so we can let them know
	u, err:= userDB.GetUser(aService.pool, userName)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, DBfailure)
		return
	}

	// Email them that we were successful!
	contents:= subEmailContents{
		Name: userName,
		Plan: subContainer.Plan,
	}
	targetAddress:= mailer.FormatAddress(userName, u.Email)
	err = aService.mailer.SendPrepared("unSubSuccess", contents,
		targetAddress, "unSubscribed! - Preorda.in")
	if err!=nil {
		aService.logger.Println("failed to send email", err)
	}

	resp.WriteEntity(true)

}

// Returns the user's subscription status as a string
func (aService *UserService) getSubUser(req *restful.Request,
	resp *restful.Response) {
	
	userName, sessionKey, err:= getUserNameAndSessionKey(req)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BodyReadFailure)
		return
	}

	if sessionKey == nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	s, err:= userDB.GetSub(aService.pool, userName, sessionKey)
	if err!=nil {
		resp.WriteErrorString(http.StatusBadRequest, BadCredentials)
		return
	}

	resp.WriteEntity(s.Plan)

}