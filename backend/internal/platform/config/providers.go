package config

import "time"

type Providers struct {
	GoogleMapsAPIKey  string
	PlacesBaseURL     string
	RoutesBaseURL     string
	WeatherBaseURL    string
	RequestTimeout    time.Duration
	ReferenceCurrency string
}

func (p Providers) UsesGoogle() bool { return p.GoogleMapsAPIKey != "" }

func (r *reader) providers() Providers {
	return Providers{
		GoogleMapsAPIKey:  r.text("GOOGLE_MAPS_API_KEY", ""),
		PlacesBaseURL:     r.text("NUVRO_PLACES_BASE_URL", "https://places.googleapis.com"),
		RoutesBaseURL:     r.text("NUVRO_ROUTES_BASE_URL", "https://routes.googleapis.com"),
		WeatherBaseURL:    r.text("NUVRO_WEATHER_BASE_URL", "https://weather.googleapis.com"),
		RequestTimeout:    r.duration("NUVRO_PROVIDER_TIMEOUT", 5*time.Second),
		ReferenceCurrency: r.text("NUVRO_REFERENCE_CURRENCY", "USD"),
	}
}

func (c Config) providerProblems() []string {
	var problems []string
	if c.Providers.RequestTimeout <= 0 || c.Providers.RequestTimeout > 30*time.Second {
		problems = append(problems, "NUVRO_PROVIDER_TIMEOUT must be between 1s and 30s")
	}
	if len(c.Providers.ReferenceCurrency) != 3 {
		problems = append(problems, "NUVRO_REFERENCE_CURRENCY must be an ISO 4217 code")
	}
	for name, value := range map[string]string{
		"NUVRO_PLACES_BASE_URL":  c.Providers.PlacesBaseURL,
		"NUVRO_ROUTES_BASE_URL":  c.Providers.RoutesBaseURL,
		"NUVRO_WEATHER_BASE_URL": c.Providers.WeatherBaseURL,
	} {
		if value == "" {
			problems = append(problems, name+" must not be empty")
		}
	}
	return problems
}
