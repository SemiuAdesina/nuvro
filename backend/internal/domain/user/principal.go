package user

type Principal struct {
	UserID   string
	Subject  string
	Provider string
}

type Profile struct {
	UserID                   string
	PreferredLanguage        string
	Locale                   string
	HomeCountry              *string
	HomeTimezone             *string
	DepartureAirport         *string
	PreferredCurrency        string
	Units                    string
	HourCycle                string
	AccessibilityPreferences []string
}

type Preferences struct {
	TravelStyle           []string
	Weather               string
	Crowds                string
	MaxFlightMinutes      *int32
	Pace                  string
	Interests             []string
	TravelingWithChildren bool
	Scope                 string
	EntrySimplicity       string
}
