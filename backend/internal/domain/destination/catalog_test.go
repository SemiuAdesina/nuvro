package destination_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
)

func valid() destination.Destination {
	return destination.Destination{
		ID: "zanzibar-tz", Name: "Zanzibar", CountryCode: "TZ", Timezone: "Africa/Dar_es_Salaam",
		Latitude: -6.1659, Longitude: 39.2026, PrimaryAirport: "ZNZ",
	}
}

func TestCatalogRecordValidation(t *testing.T) {
	if err := valid().Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*destination.Destination){
		"no id":         func(d *destination.Destination) { d.ID = "" },
		"no name":       func(d *destination.Destination) { d.Name = "" },
		"bad country":   func(d *destination.Destination) { d.CountryCode = "TZA" },
		"bad airport":   func(d *destination.Destination) { d.PrimaryAirport = "Zanzibar" },
		"no timezone":   func(d *destination.Destination) { d.Timezone = "" },
		"bad latitude":  func(d *destination.Destination) { d.Latitude = 91 },
		"bad longitude": func(d *destination.Destination) { d.Longitude = -181 },
	} {
		record := valid()
		mutate(&record)
		if err := record.Validate(); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}
