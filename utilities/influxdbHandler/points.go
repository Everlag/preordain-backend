package influxdbHandler

import(

	"encoding/json"

)

type Points []Point

func (somePoints *Points) toJSON() ([]byte, error) {

	data, err:= json.Marshal(somePoints)
	if err!=nil {
		return nil, err
	}

	return data, nil

}

type Point struct{

	Name string `json:"name"`
	Columns []string `json:"columns"`
	Points [][]PointData `json:"points"`
}

// Returns the index of the column with the provided name. -1 is returned
// if we fail to find the desired column
func (aPoint *Point) GetColumnIndex(columnName string) int {
	
	for i, aColumn:= range aPoint.Columns{
		if aColumn == columnName {
			return i
		}
	}

	return -1

}

// We need to mix ints and strings so this gets us to where we need to be
type PointData interface{}

// Builds a standard point.
func BuildPoint(seriesName string, time int64, price int64,
	set, source string) Point {
	
	cleanedSetName:= NormalizeName(set)

	data:= make([]PointData, ColumnCount)
	data[0] = PointData(time)
	data[1] = PointData(price)
	data[2] = PointData(cleanedSetName)
	data[3] = PointData(source)

	wrappedData:= make([][]PointData, 1)
	wrappedData[0] = data

	aPoint:= Point{
		Name: seriesName,
		Columns: []string(Columns),
		Points: wrappedData,
	}

	return aPoint

}

// Builds a point suitable for use with prices that started as euros and
// were converted into usd.
func BuildPointWithEuro(seriesName string, time int64, price int64, euro int64,
	set, source string) Point {
	
	cleanedSetName:= NormalizeName(set)

	data:= make([]PointData, EuroColumnCount)
	data[0] = PointData(time)
	data[1] = PointData(price)
	data[2] = PointData(euro)
	data[3] = PointData(cleanedSetName)
	data[4] = PointData(source)

	wrappedData:= make([][]PointData, 1)
	wrappedData[0] = data

	aPoint:= Point{
		Name: seriesName,
		Columns: []string(EuroColumns),
		Points: wrappedData,
	}

	return aPoint

}

// Builds a point with multiple prices associated with multiple times.
//
// Typical use is for importing external price history.
func BuildPointMultiplePrices(seriesName string, times []int64, prices []int64,
	set, source string) Point {
	
	cleanedSetName:= NormalizeName(set)

	wrappedData:= make([][]PointData, len(times))

	var price int64
	for i, time:= range times{
		price = prices[i]

		data:= make([]PointData, ColumnCount)
		data[0] = PointData(time)
		data[1] = PointData(price)
		data[2] = PointData(cleanedSetName)
		data[3] = PointData(source)

		wrappedData[i] = data

	}

	aPoint:= Point{
		Name: seriesName,
		Columns: []string(Columns),
		Points: wrappedData,
	}

	return aPoint

}

func pointsFromBytes(pointBytes []byte) (Points, error) {
	
	var somePoints Points
	err:= json.Unmarshal(pointBytes, &somePoints)
	if err!=nil {
		return Points{}, err
	}

	return somePoints, nil

}