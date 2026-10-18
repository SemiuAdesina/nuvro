package googleroutes

type computedRoute struct {
	Duration       string `json:"duration"`
	DistanceMeters int    `json:"distanceMeters"`
}

type computeResponse struct {
	Routes []computedRoute `json:"routes"`
}

type matrixElement struct {
	OriginIndex      int    `json:"originIndex"`
	DestinationIndex int    `json:"destinationIndex"`
	Duration         string `json:"duration"`
	DistanceMeters   int    `json:"distanceMeters"`
	Condition        string `json:"condition"`
}

func (m matrixElement) usable(points int) bool {
	if m.OriginIndex < 0 || m.OriginIndex >= points || m.DestinationIndex < 0 || m.DestinationIndex >= points {
		return false
	}
	if m.OriginIndex == m.DestinationIndex {
		return false
	}
	return m.Condition == "ROUTE_EXISTS"
}
