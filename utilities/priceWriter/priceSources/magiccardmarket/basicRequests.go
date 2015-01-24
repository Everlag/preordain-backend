package magiccardmarket

import(

	"fmt"

	"time"

	"net/url"
	"strings"
	"strconv"

	"crypto/hmac"
	"crypto/sha1"

	"encoding/base64"

	"net/http"

	"io/ioutil"

)

func getResource(target, consumerKey, consumerSecret string,
	aClient *http.Client) ([]byte, error) {

	req, err:= buildRequest("GET", target, consumerKey, consumerSecret)
	if err!=nil {
		return nil, fmt.Errorf("Failed to build request, ", err)
	}

	// We perform some wizardry with the url to ensure that we have colons
	// url-encoded. Basically, we set the non-scheme section of the url to be
	// properly url encoded. 
	req.URL.Opaque = strings.Replace(target, "https:", "", -1)	

	resp, err:= aClient.Do(req)
	if err!=nil {
		return nil, fmt.Errorf("Request failed, ", err)
	}
	if resp.StatusCode != 200 {
		return nil,
		fmt.Errorf("Request failed, status code of ",resp.StatusCode)	
	}

	defer resp.Body.Close()
	body, err:= ioutil.ReadAll(resp.Body)
	if err!=nil {
		return nil, fmt.Errorf("Failed to read request body, ", err)
	}

	return body, nil

}

func buildRequest(method, realm, consumerKey,
	appSecret string) (*http.Request, error) {

	signature, time, nonce:= getSignature(method, realm, consumerKey, appSecret)

	authContents:= buildAuthContents(method, realm, consumerKey, signature,
		time, nonce)


	req, err:= http.NewRequest("GET", realm, nil)
	if err!=nil {
		return nil, err
	}
	req.Header.Add("Authorization", authContents)

	return req, nil

}

func buildAuthContents(method, realm, consumerKey, signature,
		time, nonce string) string {
	
	realmReady:= "realm=\"" + realm + "\""
	consumerKeyReady:= "oauth_consumer_key=\"" + consumerKey + "\""
	nonceReady:= "oauth_nonce=\"" + nonce + "\""
	signatureMethodReady:= "oauth_signature_method=\"" + oauthSignatureMethod + "\""
	timeStampReady:= "oauth_timestamp=\"" + time + "\""
	tokenReady:= "oauth_token=\"\""
	versionReady:= "oauth_version=\"" + oauthVersion + "\""
	signatureReady:= "oauth_signature=\"" + signature + "\""

	authContents:= "OAuth" + " " +
		realmReady + "," +
		consumerKeyReady + "," +
		tokenReady + "," +
		nonceReady + "," +
		timeStampReady + "," +
		signatureMethodReady + "," +
		versionReady + "," +
		signatureReady

	return authContents

}

// Absolutely gross manual OAuth sig building
func getSignature(method, realm, consumerKey,
	appSecret string) (signature, timeStampString, nonce string) {
	
	// We build the oauth signature we require
	baseRequest:= method + "&" + realm + "&"

	timeStamp:= time.Now().UTC().Unix()
	timeStampString = strconv.FormatInt(timeStamp, 10)
	nonce = randString(13)

	// Get these ready to be stuck on the url.
	//
	// These must be added in THIS order.
	consumerKeyReady:= "oauth_consumer_key=" + consumerKey
	nonceReady:= "oauth_nonce=" + nonce
	signatureMethodReady:= "oauth_signature_method=" + oauthSignatureMethod
	timeStampReady:= "oauth_timestamp=" + timeStampString
	tokenReady:= "oauth_token="
	versionReady:= "oauth_version=" + oauthVersion

	parameters:= consumerKeyReady + "&" +
					nonceReady + "&" +
					signatureMethodReady + "&" +
					timeStampReady + "&" +
					tokenReady + "&" +
					versionReady

	// Percent encode the request
	parameters = url.QueryEscape(parameters)
	// Clean up the incompatibility introduced by go's use of +
	// for spaces
	parameters = strings.Replace(parameters, "+", "%20", -1)

	// Deal with some odd issues in the baseRequest
	baseRequest = url.QueryEscape(baseRequest)
	baseRequest = strings.Replace(baseRequest, "+", "%20", -1)
	baseRequest = strings.Replace(baseRequest, "%26", "&", -1)

	// Bring request together with parameters
	fullRequest:= baseRequest + parameters

	// Now actual create the signature
	signingKey:= url.QueryEscape(appSecret) + "&"

	signatureComputer:= hmac.New(sha1.New, []byte(signingKey))
	signatureComputer.Write([]byte(fullRequest))
	signatureBytes:= signatureComputer.Sum(nil)

	signature = base64.StdEncoding.EncodeToString(signatureBytes)

	return 

}