package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx"

	// Absurdly gross but this is very much a throwaway script
	"../." // priceDB
	"../../../api/Prices/priceReader"
	"../../influxdbHandler"
)

const setListLoc string = "setList.txt"

func main() {

	logger := GetLogger("priceLogger.txt", "migrate")

	migrate(logger)

	test(logger)
}

// Performs the migration split across 4 workers to bring the total
// work time from 8 hours to 2
func migrate(logger *log.Logger) {
	// Connect to the remote data source
	priceClient, err := priceReader.AcquireReader()
	if err != nil {
		logger.Fatalln("failed to acquire influxdb client, ", err)
	}

	// Connect to the remote data target
	pool, err := priceDB.Connect()
	if err != nil {
		logger.Fatalln("failed to acquire priceDB client", err)
	}

	series, err := priceClient.ListSeries()
	if err != nil {
		log.Fatalln(err)
	}

	// This is all returned as a single set of points
	validNames := make([]string, 0)
	points := series[0]
	nameIndex := points.GetColumnIndex("name")
	for _, p := range points.Points {
		// Find where the name is located
		name := (p[nameIndex]).(string)

		// Add all those that weren't mechanically generated for
		// continuous queries
		if !strings.Contains(name, "WeeksMedian") &&
			!strings.Contains(name, ".") {
			validNames = append(validNames, name)
		}

	}

	// Get translation between what sits on influxdb and what we deal with
	inverter := getSetConversions(priceClient, logger)

	// Create some communication routes
	cards := make(chan string, len(validNames))
	done := make(chan bool, len(validNames))

	// Launch the workers
	logger.Println("Launching workers")
	workerCount := 4
	for i := 0; i < workerCount; i++ {
		go migrateCards(cards, done, logger, priceClient, pool, inverter)
	}

	// Fill the channels
	logger.Println("Filling cards buffer")
	for _, card := range validNames {
		cards <- card
	}

	// Wait till all workers are done
	logger.Println("Waiting on migration")
	migrated := 0
	for migrated < len(validNames) {
		<-done
		migrated++
	}

	logger.Println("DONE")
}

// Reads card names off a channel and migrates them.
// Reports each success on the done channel.
func migrateCards(cards chan string, done chan bool, logger *log.Logger,
	priceClient *influxdbHandler.Client, pool *pgx.ConnPool,
	influxdbSetToReal map[string]string) {

	var err error
	var name string

	var totalTime time.Duration
	var delta time.Duration
	var successful int
	for {
		name = <-cards

		start := time.Now()
		err = migrateCard(name, influxdbSetToReal, priceClient, pool, logger)

		end := time.Now()
		if err != nil {
			logger.Println("ERROR failed to migrate", name, err)
		} else {
			delta = end.Sub(start)
			totalTime += delta
			successful++
			logger.Println("migrated", name, "took: ", delta,
				", avg time is, ", totalTime/time.Duration(successful))
		}

		done <- true
	}

}

// Migrate all of the cards various printings from the remote server
func migrateCard(name string, influxdbSetToReal map[string]string,
	priceClient *influxdbHandler.Client, pool *pgx.ConnPool,
	logger *log.Logger) error {

	series, err := priceClient.SelectEntireSeries(name)
	if err != nil {
		return err
	}

	points := series[0]
	prices := make(priceDB.Prices, 0)
	priceIndex := points.GetColumnIndex("price")
	euroIndex := points.GetColumnIndex("euro")
	sourceIndex := points.GetColumnIndex("source")
	setIndex := points.GetColumnIndex("set")
	timeIndex := points.GetColumnIndex("time")

	printed := false

	for _, p := range points.Points {

		if priceIndex < 0 {
			logger.Println(name, "has point with price index at ", priceIndex)
			continue
		}

		// We are fed floats where we expect ints.
		price := int32(p[priceIndex].(float64))

		var euro int32
		if euroIndex != -1 {
			euroFloating, ok := p[euroIndex].(float64)
			if ok {
				// Euro may or may not be present as a non-nil
				euro = int32(euroFloating)
			}
		}

		source := p[sourceIndex].(string)

		dirtySet := p[setIndex].(string)
		dirtySet = priceDB.NormalizeEMDash(dirtySet)
		set, ok := influxdbSetToReal[dirtySet]
		if !ok && !printed && strings.Contains(dirtySet, "Foil") {
			logger.Println("Invalid set name translation for", dirtySet)
			logger.Println([]byte(dirtySet))
			printed = true
		}

		// We get the timestamp as a float, we switch that
		// to an integer and then get a time as the result
		timestamp := int64(p[timeIndex].(float64))
		time := time.Unix(timestamp, 0)

		prices = append(prices, priceDB.Price{
			Name:   name,
			Set:    set,
			Time:   priceDB.Timestamp(time),
			Euro:   euro,
			Price:  price,
			Source: source,
		})

	}

	return priceDB.SendPrices(pool, prices)

}

// Performs some basic quality assurance of the sql we use
// to ensure everything is sqeaky clean and working
func test(logger *log.Logger) {

	// Connect to the remote data target
	pool, err := priceDB.Connect()
	if err != nil {
		logger.Fatalln("failed to acquire priceDB client", err)
	}

	card := "Thoughtseize"
	set := "Theros"
	var ds priceDB.Prices

	start:= time.Now()

	// Full set prices
	ds, err = priceDB.GetSetLatest(pool, "Theros", priceDB.Magiccardmarket)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}
	if len(ds) == 0 {
		logger.Fatalln("no data for", set)
	}
	ds, err = priceDB.GetSetLatest(pool, "Theros", priceDB.Mtgprice)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}
	if len(ds) == 0 {
		logger.Fatalln("no data for", set)
	}

	// Basic Latest
	_, err = priceDB.GetCardLatest(pool, card, set, priceDB.Mtgprice)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}
	_, err = priceDB.GetCardLatest(pool, card, set, priceDB.Magiccardmarket)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}

	// Latest Highest
	_, err = priceDB.GetCardLatestHighest(pool, card, priceDB.Mtgprice)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}
	_, err = priceDB.GetCardLatestHighest(pool, card, priceDB.Magiccardmarket)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}

	// Latest Lowest
	_, err = priceDB.GetCardLatestLowest(pool, card, priceDB.Mtgprice)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}
	_, err = priceDB.GetCardLatestLowest(pool, card, priceDB.Magiccardmarket)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}

	// Full card-set historical prices
	ds, err = priceDB.GetCardHistory(pool, card, "Theros",
		priceDB.Magiccardmarket)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}
	if len(ds) == 0 {
		logger.Fatalln("no data for", set)
	}
	ds, err = priceDB.GetCardHistory(pool, card, "Theros", priceDB.Mtgprice)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}
	if len(ds) == 0 {
		logger.Fatalln("no data for", set)
	}

	// Full card-set historical prices aggregated to weekly median values
	ds, err = priceDB.GetCardMedianHistory(pool, card, "Theros",
		priceDB.Magiccardmarket)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}
	if len(ds) == 0 {
		logger.Fatalln("no data for", set)
	}
	ds, err = priceDB.GetCardMedianHistory(pool, card, "Theros", priceDB.Mtgprice)
	if err != nil {
		logger.Println("failed")
		logger.Fatalln(err)
	}
	if len(ds) == 0 {
		logger.Fatalln("no data for", set)
	}

	end:= time.Now()
	logger.Println(end.Sub(start), "elapsed")
	logger.Println("Tests passed, please manually inspect")
}

// Most sets get transformed for easier handling within influxdb, we create
// a map that inverts that transformation.
func getSetConversions(priceClient *influxdbHandler.Client,
	logger *log.Logger) map[string]string {

	// Acquire the list of sets we deal with
	sets, err := getSetList()
	if err != nil {
		panic("failed to get setlist")
	}

	inverter := make(map[string]string)

	for _, set := range sets {
		dirty := priceClient.NormalizeName(set)
		// Remove some unicode that is causing headaches
		dirty = priceDB.NormalizeEMDash(dirty)
		set = priceDB.NormalizeEMDash(set)
		inverter[dirty] = set
	}

	return inverter
}

func getSetList() ([]string, error) {

	sets, err := ioutil.ReadFile(setListLoc)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(sets), "\n"), nil

}

func GetLogger(fName, name string) (aLogger *log.Logger) {
	file, err := os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Starting logger failed, cannot write to logger to say logger failed. Oh god.")
		fmt.Println(err)
		os.Exit(0)
	}

	multi := io.MultiWriter(file, os.Stdout)

	aLogger = log.New(multi, name, log.Lshortfile)

	return
}
