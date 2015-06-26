package priceDB

import (
	"fmt"

	"github.com/jackc/pgx"

	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"crypto/tls"
)

const magiccardmarket string = "mkm"
const mtgprice string = "mtgprice"

// A list of all statements we support, these are prepared on a per
// connection basis.
const mtgpriceInsert string = "addPriceMTGprice"
const mkmpriceInsert string = "addPriceMKMprice"

const mtgpriceHistory string = "historicalPriceMTGprice"
const mkmpriceHistory string = "historicalPriceMKMprice"

const mtgPriceLatest string = "mtgPriceLatest"
const mkmPriceLatest string = "mkmPriceLastest"

const mtgpriceMedian string = "medianMtgprice"
const mkmMedian string = "medianMKM"

const mtgPriceLatestLowest string = "mtgPriceLatestLowest"
const mkmPriceLatestLowest string = "mkmPriceLatestLowest"
const mtgPriceLatestHighest string = "mtgPriceLatestHighest"
const mkmPriceLatestHighest string = "mkmPriceLatestHighest"

const mtgPriceSetLatest string = "mtgPriceSetLatest"
const mkmPriceSetLatest string = "mkmPriceSetLatest"

var statements = []string{
	mtgpriceInsert, mkmpriceInsert,
	mtgpriceHistory, mkmpriceHistory,
	mtgPriceLatest, mkmPriceLatest,
	mtgpriceMedian, mkmMedian,
	mtgPriceLatestLowest, mkmPriceLatestLowest,
	mtgPriceLatestHighest, mkmPriceLatestHighest,
	mtgPriceSetLatest, mkmPriceSetLatest,
}

const statementLoc string = "sql"
const statementExtension string = ".sql"

const configLoc string = "postgres.config.json"
const certLoc string = "certs/server.crt"

// Exposed and available
const Magiccardmarket string = magiccardmarket
const Mtgprice string = mtgprice
var Sources []string = []string{Mtgprice, Magiccardmarket}

func fetchRawStatement(name string) (string, error) {

	loc := filepath.Join(statementLoc, name)
	loc = loc + statementExtension

	result, err := ioutil.ReadFile(loc)
	if err != nil {
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
	for _, statementName := range statements {

		// Each connection loads the statement off of disk. Not optimal
		text, err = fetchRawStatement(statementName)
		if err != nil {
			return err
		}

		_, err = conn.Prepare(statementName, text)
		if err != nil {
			err = fmt.Errorf("Failed to prepare statement ", statementName, err)
			return err
		}

	}

	return

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
	connPoolConfig, err := readConfig(configLoc)
	if err != nil {
		return nil, err
	}

	pool, err := pgx.NewConnPool(*connPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool,", err)
	}

	return pool, nil

}

type Config struct {
	Host, User, Password, Database string
}

func readConfig(loc string) (*pgx.ConnPoolConfig, error) {
	raw, err := ioutil.ReadFile(loc)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire config", err)
	}

	var c Config
	err = json.Unmarshal(raw, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config", err)
	}

	// Acquire our trust chain so we can connect
	trustRoot, err := grabCert(certLoc)
	if err != nil {
		return nil, err
	}

	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     c.Host,
			User:     c.User,
			Password: c.Password,
			Database: c.Database,
			TLSConfig: &tls.Config{
				RootCAs:            trustRoot,
				InsecureSkipVerify: true,
			},
		},
		MaxConnections: 50,
		AfterConnect:   afterConnect,
	}

	return &connPoolConfig, nil

}
