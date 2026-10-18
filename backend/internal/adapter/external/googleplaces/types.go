package googleplaces

type latLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type localizedText struct {
	Text string `json:"text"`
}

type authorAttribution struct {
	DisplayName string `json:"displayName"`
	URI         string `json:"uri"`
}

type photo struct {
	Name               string              `json:"name"`
	WidthPx            int                 `json:"widthPx"`
	HeightPx           int                 `json:"heightPx"`
	AuthorAttributions []authorAttribution `json:"authorAttributions"`
}

type period struct {
	Open  dayPoint `json:"open"`
	Close dayPoint `json:"close"`
}

type dayPoint struct {
	Day    int `json:"day"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

type openingHours struct {
	OpenNow *bool    `json:"openNow"`
	Periods []period `json:"periods"`
}

type placePayload struct {
	ID                  string        `json:"id"`
	DisplayName         localizedText `json:"displayName"`
	FormattedAddress    string        `json:"formattedAddress"`
	Location            latLng        `json:"location"`
	PrimaryType         string        `json:"primaryType"`
	Types               []string      `json:"types"`
	Rating              float64       `json:"rating"`
	UserRatingCount     int           `json:"userRatingCount"`
	PriceLevel          string        `json:"priceLevel"`
	RegularOpeningHours *openingHours `json:"regularOpeningHours"`
	CurrentOpeningHours *openingHours `json:"currentOpeningHours"`
	Photos              []photo       `json:"photos"`
	UTCOffsetMinutes    *int          `json:"utcOffsetMinutes"`
}

type searchResponse struct {
	Places []placePayload `json:"places"`
}
