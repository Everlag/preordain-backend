package main

import(

	"net/http"
	"encoding/json"

	"./../influxdbHandler"

	"io/ioutil"

	"fmt"
	"os"

)


const sourceBaseUrl string = "http://localhost:7801"

const setListPath string = "/SetList"
const priceListPartial string = "/Prices/"

// Where we acquired these prices from
const priceSource string = "mtgprice"

// Imports from the old format of json flat files to influxdb.
//
// Due to the change of paradigmn, set orientation to card orientation,
// this imports every set then uploads in batches of sets at a time
//
// Writes the setList.txt file to disk which contains every set we deal with
func Import(aClient *influxdbHandler.Client) {

	var err error
	
	// Acquire the points from the remote source
	importedPoints, setList, err := importAllSets()
	if err!=nil {
		fmt.Println("Failed to import points, ", err)
		os.Exit(1)
	}

	// Write to disk the setList
	setListText:= ""
	for _, aSet:= range setList{
		setListText+= aSet + "\n"
	}
	
	ioutil.WriteFile("setList.txt", []byte(setListText), 0666)

	
	// Upload the points to influxdb
	for _, somePoints:= range importedPoints{

		err = aClient.SendPoints(somePoints)
		if err!=nil {
			fmt.Println("Failed to send these points, continuing: ", somePoints)
		}

	}
	
}

func importAllSets() ([]influxdbHandler.Points, []string, error) {
	
	// First, acquire the set list from the remote source
	resp, err:= http.Get(sourceBaseUrl + setListPath)
	if err!=nil {
		return nil, nil, fmt.Errorf("Failed to get setList, ", err)
	}
	defer resp.Body.Close()

	var setList []string
	err = json.NewDecoder(resp.Body).Decode(&setList)
	if err!=nil {
		return nil, nil, fmt.Errorf("Failed to decode set list")
	}


	// then populate the data one set at a time.
	points:= make([]influxdbHandler.Points, 0)

	for _, aSet:= range setList{
		if aSet == "" {
			continue
		}
		aPoint, err:= importSet(aSet)
		if err!=nil {
			fmt.Println("Failed to import ", aSet, ", no data has been written to influxdb")
			os.Exit(1)
		}
		points = append(points, aPoint)
	}

	return points, setList, nil

}

type PriceList struct{
	Cards []CardPrices
	UpdateTimes []int64
	LatestUpdate int64
}

type CardPrices struct{
	Name string
	Prices []int
}

func importSet(setName string) (influxdbHandler.Points, error) {
	
	// Grab the set with all data
	path:= sourceBaseUrl + priceListPartial + setName + "/-1"

	resp, err:= http.Get(path)
	if err!=nil {
		return influxdbHandler.Points{}, fmt.Errorf("Failed to get set, ", err)
	}
	defer resp.Body.Close()

	var setPrices PriceList
	err = json.NewDecoder(resp.Body).Decode(&setPrices)
	if err!=nil {
		return influxdbHandler.Points{}, err
	}

	points:= make(influxdbHandler.Points, len(setPrices.Cards))

	for cardIndex, aCard := range setPrices.Cards{

		prices:= make([]int64, len(aCard.Prices))
		for i, _:= range aCard.Prices{
			prices[i] = int64(aCard.Prices[i])
		}
		times:= make([]int64, len(prices))

		// we iterate through at the temporally closest prices
		// to get the correct time for free. starting at the farthest
		// time could cause problems with cards added after the
		// creation of the set
		for i := (len(prices) - 1); i >= 0; i-- {
			
			// index into time is given by the length of the available
			// times moved
			// (the length of prices for this card - the index into those prices)
			// back
			timeIndex:= (len(setPrices.UpdateTimes) - 1) - (len(prices) - 1 - i)

			if timeIndex <= -1{
				//fmt.Println(timeIndex, setName, aCard.Name)
				continue		
			}

			time:= setPrices.UpdateTimes[timeIndex]

			times[i] = time

		}

		aPoint:= influxdbHandler.BuildPointMultiplePrices(aCard.Name,
			times, prices,
			setName, priceSource)

		points[cardIndex] = aPoint

	}

	return points, nil

}