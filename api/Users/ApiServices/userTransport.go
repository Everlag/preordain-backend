package ApiServices

import(

	"./../../../utilities/userDBHandler"

)

// While gross, having a struct for each request lets me keep this as a json
// based api without too much pain in go-restful
type NewUserData struct{

	Email, Password string
	RecaptchaResponseField string

}

type PasswordBody struct{
	Password string
}

type SessionKeyBody struct{
	SessionKey []byte
}

type PermissionChangeBody struct{
	SessionKey []byte
	Privacy string
}

type TradeAddBody struct{

	Trade []userDB.Card
	SessionKey []byte

}

type PasswordResetRequestBody struct{

	RecaptchaResponseField string

}

type PasswordResetBody struct{

	Password string
	ResetRequestToken string

}

type CollectionContents struct{
	Current []userDB.Card
	Historical []userDB.Card
}