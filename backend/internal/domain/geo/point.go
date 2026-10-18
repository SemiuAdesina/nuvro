package geo

import (
	"errors"
	"math"
)

var ErrInvalidPoint = errors.New("geo: invalid coordinate")

const earthRadiusMeters = 6371000

type Point struct {
	Latitude  float64
	Longitude float64
}

func (p Point) Validate() error {
	if math.IsNaN(p.Latitude) || math.IsNaN(p.Longitude) {
		return ErrInvalidPoint
	}
	if p.Latitude < -90 || p.Latitude > 90 || p.Longitude < -180 || p.Longitude > 180 {
		return ErrInvalidPoint
	}
	return nil
}

func (p Point) IsZero() bool { return p.Latitude == 0 && p.Longitude == 0 }

func (p Point) DistanceMeters(other Point) float64 {
	first := p.Latitude * math.Pi / 180
	second := other.Latitude * math.Pi / 180
	deltaLatitude := (other.Latitude - p.Latitude) * math.Pi / 180
	deltaLongitude := (other.Longitude - p.Longitude) * math.Pi / 180
	haversine := math.Sin(deltaLatitude/2)*math.Sin(deltaLatitude/2) +
		math.Cos(first)*math.Cos(second)*math.Sin(deltaLongitude/2)*math.Sin(deltaLongitude/2)
	return 2 * earthRadiusMeters * math.Atan2(math.Sqrt(haversine), math.Sqrt(1-haversine))
}
