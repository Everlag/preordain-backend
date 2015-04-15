package userDB

import(

	"fmt"

	"github.com/jackc/pgx"

	"io/ioutil"
	"path/filepath"

	"time"

)

// A list of all statements we support, these are prepared on a per
// connection basis.
var statements = []string{"addCard", "addCardHistorical" , "getCard",
						"addCollection", "getCollectionMeta", "getCollectionList",
						"getCollectionContents", "getCollectionHistory",
						"getSessions", "addSession", "removeSession",
						"getReset", "addReset",
						"addUser", "getUser", "setPassword",
						"setMaxCollections", "setCollectionPermissions"}
const statementLoc string = "sql"
const statementExtension string = ".sql"

// Each session can be valid for up to a month and
// each reset request valid up to one day
const hoursPerDay int = 24
const hoursPerMonth int = 30 * hoursPerDay
const sessionValidTime = time.Duration(hoursPerMonth) * time.Hour
const resetValidTime = time.Duration(hoursPerDay) * time.Hour

var ScanError string = "failed to scan row"

func fetchRawStatement(name string) (string, error) {
	
	loc:= filepath.Join(statementLoc, name)
	loc = loc + statementExtension

	result, err:= ioutil.ReadFile(loc)
	if err!= nil{
		return "", fmt.Errorf("failed to acquire statement, ", err)
	}

	return string(result), nil

}

// Sets everything a connection could need up.
//
// Lets us refer to our stored sql very easily.
func afterConnect(conn *pgx.Conn) (err error) {

	// Prepare all the predefined statements
	var text string
	for _, statementName:= range statements{

		// Each connection loads the statement off of disk. Not optimal
		text, err = fetchRawStatement(statementName)
		if err!=nil {
			return err
		}

		_, err = conn.Prepare(statementName, text)
		if err!=nil {
			err = fmt.Errorf("Failed to prepare statement ", statementName, err)
			return err
		}

	}

	return

}

// Handles errors with special cases. Prompt is the error that
// prompted an error to be generated and message is the general
// message to display in this context without special case.
func errorHandle(prompt error, message string) error {
	// Special cases
	if prompt == pgx.ErrNoRows {
		return prompt
	}else if prompt == pgx.ErrDeadConn {
		return prompt
	}
	return fmt.Errorf(message, prompt)
}

func Connect() (*pgx.ConnPool, error) {
	
	// TODO: add tls support!
	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     "127.0.0.1",
			User:     "usermanager",
			Password: "$insertPasswordHere",
			Database: "userdata",
			TLSConfig: nil,
		},
		MaxConnections: 50,
		AfterConnect:   afterConnect,
	}

	pool, err:= pgx.NewConnPool(connPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool,", err)
	}

	return pool, nil

}

/*
func main() {

	cards, err:= getCollectionContents(pool, nil, "everlag", "specs")
	if err!=nil {
		fmt.Println("failed to get card list, ", err)
		os.Exit(1)
	}
	fmt.Println(cards)
}
*/