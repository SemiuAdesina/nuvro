package destination

import (
	"errors"
	"regexp"
)

var ErrInvalidDestination = errors.New("destination: invalid catalog record")

var (
	countryPattern = regexp.MustCompile(`^[A-Z]{2}$`)
	airportPattern = regexp.MustCompile(`^[A-Z]{3}$`)
)

type Destination struct {
	ID             string
	Name           string
	Region         string
	CountryCode    string
	Timezone       string
	Latitude       float64
	Longitude      float64
	PrimaryAirport string
	ClimateTags    []string
	InterestTags   []string
	StyleTags      []string
	CrowdProfile   string
	PaceProfile    string
}

func (d Destination) Validate() error {
	if d.ID == "" || d.Name == "" || !countryPattern.MatchString(d.CountryCode) {
		return ErrInvalidDestination
	}
	if !airportPattern.MatchString(d.PrimaryAirport) || d.Timezone == "" {
		return ErrInvalidDestination
	}
	if d.Latitude < -90 || d.Latitude > 90 || d.Longitude < -180 || d.Longitude > 180 {
		return ErrInvalidDestination
	}
	return nil
}
