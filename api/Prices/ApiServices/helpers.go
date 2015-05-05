package ApiServices

import(

	"io/ioutil"
	"strings"

	"io"
	"log"
	"os"
	"fmt"

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

// Sets a cache header of 5 hours to a given request.
func setCacheHeader(resp *restful.Response) {
	resp.Header().Set("Cache-Control", "max-age=18000,s-maxage=18000")
}

// Sets a cache header of 5 hours to a given request.
func setPrivateHeader(resp *restful.Response) {
	resp.Header().Set("Cache-Control", "private")
}

func getSetList() ([]string, error) {

	sets, err:= ioutil.ReadFile(setListLoc)
	if err!=nil {
		return nil, err
	}

	return strings.Split(string(sets), "\n"), nil

}

func getIP(req *restful.Request) string {
	return strings.Split(req.Request.RemoteAddr, ":")[0]
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