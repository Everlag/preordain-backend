package ApiServices

import(

	"io/ioutil"
	"strings"

	"io"
	"log"
	"os"
	"fmt"

	"github.com/dpapathanasiou/go-recaptcha"

	"github.com/emicklei/go-restful"

	"net/http"

)

const PanicRecoverMessage string = "Something really bad happened while completing your request :("

const setListLoc string = "setList.txt"
const managerNameLoc string = "managerMeta.txt"
const recaptchaKeyLoc string = "recaptchaPrivateKey.txt"

// A basic handler for recovery to ensure that we don't accidently start
// sending stack traces.
func RecoverHandler(issue interface{}, writer http.ResponseWriter) {
	
	writer.WriteHeader(http.StatusInternalServerError)
	writer.Write([]byte(PanicRecoverMessage))

}

func getUserNameAndSessionKey(req *restful.Request) (userName,
	sessionKey string, err error) {

	userName = req.PathParameter("userName")
	var sessionKeyContainer SessionKeyBody
	err = req.ReadEntity(&sessionKeyContainer)
	if err!=nil {
		return
	}

	sessionKey = sessionKeyContainer.SessionKey
	
	return

}

// Sets a cache header of 5 hours to a given request.
func setCacheHeader(resp *restful.Response) {
	resp.Header().Set("Cache-Control", "max-age=18000,s-maxage=18000")
}

func getSetList() ([]string, error) {

	sets, err:= ioutil.ReadFile(setListLoc)
	if err!=nil {
		return nil, err
	}

	return strings.Split(string(sets), "\n"), nil

}

func getManagerName() (string, error) {
	name, err:= ioutil.ReadFile(managerNameLoc)
	if err!=nil {
		return "", err
	}

	cleanedName:= strings.TrimSpace(string(name))

	return cleanedName, nil
}

func getRecaptchaKey() (string, error) {
	key, err:= ioutil.ReadFile(recaptchaKeyLoc)
	if err!=nil {
		return "", err
	}

	cleanedKey:= strings.TrimSpace(string(key))

	return cleanedKey, nil
}

func setupRecaptcha() error {
	key, err:= getRecaptchaKey()
	if err!=nil {
		return err
	}

	recaptcha.Init(key)

	return nil

}

func getIP(req *restful.Request) string {
	return strings.Split(req.Request.RemoteAddr, ":")[0]
}

//ensures that the captcha input is valid
func ValidateRecaptcha(req *restful.Request,
	challengeField, responseField string) bool {

	remoteIP:= getIP(req)

	return recaptcha.Confirm( remoteIP, challengeField, responseField )
}

func GetLogger(fName, name string) (aLogger *log.Logger) {
	file, err:= os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err!=nil {
		fmt.Println("Starting logger failed, cannot write to logger to say logger failed. Oh god.")
		fmt.Println(err)
		os.Exit(0)
	}

	multi:= io.MultiWriter(file, os.Stdout)

	aLogger = log.New(multi, name, log.Ldate|log.Ltime|log.Lshortfile)

	return
}