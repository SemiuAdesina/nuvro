package googleweather

type value struct {
	Degrees float64 `json:"degrees"`
	Unit    string  `json:"unit"`
}

type speed struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type wind struct {
	Speed speed `json:"speed"`
}

type quantity struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type probability struct {
	Percent int    `json:"percent"`
	Type    string `json:"type"`
}

type precipitation struct {
	Probability probability `json:"probability"`
	QPF         quantity    `json:"qpf"`
}

type conditionType struct {
	Type        string `json:"type"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
}

type displayTime struct {
	Year    int `json:"year"`
	Month   int `json:"month"`
	Day     int `json:"day"`
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
}

type conditionsPayload struct {
	CurrentTime          string        `json:"currentTime"`
	WeatherCondition     conditionType `json:"weatherCondition"`
	Temperature          value         `json:"temperature"`
	FeelsLikeTemperature value         `json:"feelsLikeTemperature"`
	RelativeHumidity     int           `json:"relativeHumidity"`
	UVIndex              int           `json:"uvIndex"`
	Precipitation        precipitation `json:"precipitation"`
	Wind                 wind          `json:"wind"`
}

type hourPayload struct {
	Interval struct {
		StartTime string `json:"startTime"`
	} `json:"interval"`
	DisplayDateTime      displayTime   `json:"displayDateTime"`
	WeatherCondition     conditionType `json:"weatherCondition"`
	Temperature          value         `json:"temperature"`
	FeelsLikeTemperature value         `json:"feelsLikeTemperature"`
	RelativeHumidity     int           `json:"relativeHumidity"`
	UVIndex              int           `json:"uvIndex"`
	Precipitation        precipitation `json:"precipitation"`
	Wind                 wind          `json:"wind"`
}

type hourlyResponse struct {
	ForecastHours []hourPayload `json:"forecastHours"`
}

type dayPayload struct {
	DisplayDate struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	} `json:"displayDate"`
	DaytimeForecast struct {
		WeatherCondition conditionType `json:"weatherCondition"`
		Precipitation    precipitation `json:"precipitation"`
	} `json:"daytimeForecast"`
	MaxTemperature value `json:"maxTemperature"`
	MinTemperature value `json:"minTemperature"`
	SunEvents      struct {
		SunriseTime string `json:"sunriseTime"`
		SunsetTime  string `json:"sunsetTime"`
	} `json:"sunEvents"`
}

type dailyResponse struct {
	ForecastDays []dayPayload `json:"forecastDays"`
}
