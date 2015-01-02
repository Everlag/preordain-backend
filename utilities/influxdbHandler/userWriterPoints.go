package influxdbHandler

// Builds a point for a user action
func (aClient *Client) SendUserActionPoint(userName, action string,
	actionParameters, bodyContents interface{}, time int64) error {
	
	data:= make([]PointData, UserColumnCount)
	data[0] = PointData(time)
	data[1] = PointData(action)
	data[2] = PointData(actionParameters)
	data[3] = PointData(bodyContents)

	WrappedData:= make([][]PointData, 1)
	WrappedData[0] = data

	aPoint:= Point{
		Name: userName,
		Columns: []string(UserColumns),
		Points: WrappedData,
	}

	return aClient.SendPoints(Points{aPoint})


}