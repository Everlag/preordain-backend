package main
/*
import(

	"./../influxdbHandler"

	"./priceSources"
	
)

// Where our influxdb client data is kept
const influxdbCredentials string = "influxdbCredentials.json"

func uploadSingleSourceResults(aPriceResult priceSources.PriceMap,
	aClient *influxdbHandler.Client) error {

	// Construct the points to send
	points:= make([]influxdbHandler.Point, 0)

	for aSetName, cardMap:= range aPriceResult.Prices{

		for aCardName, aPrice:= range cardMap{

			// Deal with the fact that some price sources may have multiple
			// currencies that were massaged into USD
			var aPoint influxdbHandler.Point
			if aPriceResult.HasEuro {
				// An original price in euros is recorded alongside the USD
				// conversion
				euroPrice:= aPriceResult.EURPrices[aSetName][aCardName]
				
				aPoint = influxdbHandler.BuildPointWithEuro(aCardName,
					aPriceResult.Time, aPrice, euroPrice,
					aSetName, aPriceResult.Source)
			
			}else{
			
				aPoint = influxdbHandler.BuildPoint(aCardName,
					aPriceResult.Time, aPrice, aSetName, aPriceResult.Source)
			
			}

			points = append(points, aPoint)

		}
	}

	// Send the points to the db
	err:= aClient.SendPoints(points)

	return err

}
*/