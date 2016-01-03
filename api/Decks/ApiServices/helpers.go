package ApiServices

import(

	"io"
	"log"
	"os"
	"fmt"

	"net/http"

)

const PanicRecoverMessage string = "Something really bad happened while completing your request :("

// A basic handler for recovery to ensure that we don't accidently start
// sending stack traces.
func RecoverHandler(issue interface{}, writer http.ResponseWriter) {
	
	writer.WriteHeader(http.StatusInternalServerError)
	writer.Write([]byte(PanicRecoverMessage))

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