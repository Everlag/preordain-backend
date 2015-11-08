// Provides validation for recaptcha 2.0
package recaptcha

import(

	"net/url"
	"net/http"

	"io/ioutil"
	"encoding/json"

)

const verifyEndpoint string = "https://www.google.com/recaptcha/api/siteverify"

// A recaptcha 2.0 validator.
type Validator struct{
	priv string
}

// Builds a recaptcha validator using a provided private key
func GetValidator(priv string) *Validator {
	return &Validator{priv}
}

// Builds a recaptcha validator from a json file of structure recaptchaMeta
func GetValidatorFromFile(loc string) (*Validator, error) {
	metaRaw, err:= ioutil.ReadFile(loc)
	if err!=nil {
		return nil, err
	}

	var meta recaptchaMeta
	err = json.Unmarshal(metaRaw, &meta)
	if err!=nil {
		return nil, err	
	}

	return GetValidator(meta.Private), nil
}

type recaptchaMeta struct{
	Private string
}

// Returns whether or not a recaptcha response was valid.
//
// Defaults to the ip as localhost so we need not tie a domain name to this.
func (validator *Validator) Validate(response string) (bool, error) {
	resp, err:= http.PostForm(verifyEndpoint, url.Values{
		"secret": {validator.priv},
		"response": {response},
		})
	if err!=nil{
		return false, err
	}

	defer resp.Body.Close()

	respData, err:= ioutil.ReadAll(resp.Body)
	if err!=nil {
		return false, err
	}

	var result RecaptchaResponse
	err = json.Unmarshal(respData, &result)
	if err!=nil {
		return false, err
	}

	return result.Success, nil
}

type RecaptchaResponse struct{
	Success bool
}