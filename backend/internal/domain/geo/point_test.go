package geo_test

import (
	"math"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
)

func TestCoordinateValidation(t *testing.T) {
	if err := (geo.Point{Latitude: 32.6478, Longitude: -16.9048}).Validate(); err != nil {
		t.Fatal(err)
	}
	for name, point := range map[string]geo.Point{
		"north":        {Latitude: 91, Longitude: 0},
		"south":        {Latitude: -91, Longitude: 0},
		"east":         {Latitude: 0, Longitude: 181},
		"west":         {Latitude: 0, Longitude: -181},
		"not a number": {Latitude: math.NaN(), Longitude: 0},
	} {
		if point.Validate() == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestDistanceIsSymmetricAndSane(t *testing.T) {
	funchal := geo.Point{Latitude: 32.6478, Longitude: -16.9048}
	monte := geo.Point{Latitude: 32.6690, Longitude: -16.9000}
	forward := funchal.DistanceMeters(monte)
	backward := monte.DistanceMeters(funchal)
	if math.Abs(forward-backward) > 0.001 {
		t.Fatalf("distance is not symmetric: %v vs %v", forward, backward)
	}
	if forward < 2000 || forward > 3000 {
		t.Fatalf("distance = %v metres, which is not plausible", forward)
	}
	if funchal.DistanceMeters(funchal) > 0.001 {
		t.Fatal("a point is not zero distance from itself")
	}
}

func FuzzPointValidation(f *testing.F) {
	f.Add(32.6478, -16.9048)
	f.Add(0.0, 0.0)
	f.Add(91.0, 181.0)
	f.Fuzz(func(t *testing.T, latitude, longitude float64) {
		point := geo.Point{Latitude: latitude, Longitude: longitude}
		if point.Validate() != nil {
			return
		}
		if latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
			t.Fatalf("accepted an impossible coordinate: %v %v", latitude, longitude)
		}
		distance := point.DistanceMeters(geo.Point{})
		if math.IsNaN(distance) || distance < 0 || distance > 20100000 {
			t.Fatalf("distance %v is not on this planet", distance)
		}
	})
}
