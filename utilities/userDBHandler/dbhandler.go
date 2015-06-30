package userDB

import(

	"fmt"

	"github.com/jackc/pgx"

	"io/ioutil"
	"encoding/json"
	"path/filepath"

	"time"

	"crypto/tls"

)

// The basic free tier of subs.
const DefaultSubLevel = "Peek"

// Which collections we allow
var SubTiers = []string{
	DefaultSubLevel,
	"Preordain",
	"Sensei's Top",
}

// How many collections each tier is allowed to own
var SubTiersToCollections = map[string]int{
	DefaultSubLevel: 1,
	"Preordain": 4,
	"Sensei's Top": 30,
}

// 290 years from now there should be no back records.
const noTimeLimit = time.Duration(31560000000000000 * 270)
// A full year!
const defaultTimeLimit = time.Duration(31560000000000000)


const DefaultID string = "invalidByDesign"

const configLog string = "postgres.config.json"

const certLoc string = "certs/server.crt"

// A list of all statements we support, these are prepared on a per
// connection basis.
var statements = []string{"addCard", "addCardHistorical" , "getCard",
						"addCollection", "getCollectionMeta", "getCollectionList",
						"getCollectionContents", "getCollectionHistory",
						"getSessions", "addSession", "removeSession",
						"getReset", "getAllResets", "addReset",
						"addUser", "getUser", "setPassword",
						"setMaxCollections", "setCollectionPermissions",
						"getSub", "modSub", "setSubEffects"}
const statementLoc string = "sql"
const statementExtension string = ".sql"

// Each session can be valid for up to a month and
// each reset request valid up to one day
//
// The total time a reset is valid is also the time between resets
// being able to be sent to that user
const hoursPerDay int = 24
const hoursPerMonth int = 30 * hoursPerDay
const sessionValidTime = time.Duration(hoursPerMonth) * time.Hour
const resetValidTime = time.Duration(hoursPerDay) * time.Hour

var ScanError string = "failed to scan row"

func fetchRawStatement(name string) (string, error) {
	
	loc:= filepath.Join(statementLoc, name)
	loc = loc + statementExtension

	result, err:= Asset(loc)
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

// Connects to the remote postgres server defined in postgres.config.json
//
// Uses the certificate found in certs/server.crt to establish trust
func Connect() (*pgx.ConnPool, error) {

	// Create our pool.
	//
	// In most cases InsecureSkipVerify would be very poor but since
	// we are handling our cert chain ourselves with self signed certs
	// this is not an issue.
	connPoolConfig, err:= readConfig(configLog) 
	if err!=nil {
		return nil, err
	}

	pool, err:= pgx.NewConnPool(*connPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool,", err)
	}

	return pool, nil

}

type Config struct{
	Host, User, Password, Database string
}

func readConfig(loc string) (*pgx.ConnPoolConfig, error) {
	raw, err:= ioutil.ReadFile(loc)
	if err!= nil{
		return nil, fmt.Errorf("failed to acquire config", err)
	}

	var c Config
	err = json.Unmarshal(raw, &c)
	if err!=nil {
		return nil, fmt.Errorf("failed to unmarshal config", err)
	}
	
	// Acquire our trust chain so we can connect
	trustRoot, err:= grabCert(certLoc)
	if err!=nil {
		return nil, err
	}

	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     c.Host,
			User:     c.User,
			Password: c.Password,
			Database: c.Database,
			TLSConfig: &tls.Config{
				RootCAs: trustRoot,
				InsecureSkipVerify: true,
			},
		},
		MaxConnections: 50,
		AfterConnect:   afterConnect,
	}

	return &connPoolConfig, nil

}