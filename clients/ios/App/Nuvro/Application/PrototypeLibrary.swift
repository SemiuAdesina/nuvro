import NuvroDesignSystem
import NuvroModels
import NuvroProfile

enum PrototypeLibrary {
    private static func media(_ placeholder: String, _ alt: String) -> DestinationMedia {
        DestinationMedia(placeholder: placeholder, altText: alt, attribution: "Photo selected by NUVRO · provider attribution")
    }

    static let entries = [
        TripLibraryEntry(id: "madeira", destinationName: "Madeira, Portugal", dateLabel: "Oct 12–19", phase: .active, progress: 0.43, media: media("madeira", "Madeira coastline"), statusLabel: "Day 3 of 7", weatherLabel: "23° light clouds", nextLabel: "Funchal Cable Car · 10:00"),
        TripLibraryEntry(id: "barcelona", destinationName: "Barcelona", dateLabel: "Nov 4–10 · 47 days away", phase: .upcoming, progress: 0.75, readinessLabel: "6 of 8", media: media("twilight", "Barcelona rooftops")),
        TripLibraryEntry(id: "japan", destinationName: "Japan", dateLabel: "March · flexible dates", phase: .planning, media: media("twilight", "Kyoto temple in spring")),
        TripLibraryEntry(id: "past-madeira", destinationName: "Madeira", dateLabel: "Oct 12–19", phase: .past, media: media("madeira", "Madeira")),
        TripLibraryEntry(id: "paris", destinationName: "Paris", dateLabel: "May 4–9", phase: .past, media: media("twilight", "Paris")),
        TripLibraryEntry(id: "amalfi", destinationName: "Amalfi", dateLabel: "Jul 2–8", phase: .past, media: media("curacao", "Amalfi coast")),
        TripLibraryEntry(id: "reykjavik", destinationName: "Reykjavík", dateLabel: "Feb 9–13", phase: .past, media: media("sky", "Reykjavík harbour"))
    ]

    static let savedDestinations = [
        SavedDestination(id: "kyoto", name: "Kyoto, Japan", country: "Japan", matchPercent: 91, reason: "Saved from your spring search", note: "Trip context from March · prices not current", media: media("twilight", "Kyoto")),
        SavedDestination(id: "madeira", name: "Madeira, Portugal", country: "Portugal", matchPercent: 94, reason: "Your best match in October", note: "Trip planned · Oct 12–19", media: media("madeira", "Madeira")),
        SavedDestination(id: "lisbon", name: "Lisbon, Portugal", country: "Portugal", matchPercent: 91, reason: "Saved while comparing city trips", note: "Trip context from October", media: media("gold", "Lisbon")),
        SavedDestination(id: "azores", name: "Azores, Portugal", country: "Portugal", matchPercent: 88, reason: "Suggested after your Madeira trip", note: "No dates yet", media: media("guanacaste", "Azores"))
    ]

    static let savedPlaces = [
        SavedPlace(name: "Mercado dos Lavradores", subtitle: "Funchal · saved during Madeira trip", media: media("gold", "Market")),
        SavedPlace(name: "Barreirinha Bar Café", subtitle: "Funchal · saved during Madeira trip", media: media("curacao", "Café")),
        SavedPlace(name: "Time Out Market", subtitle: "Lisbon · saved while comparing", media: media("madeira", "Market hall")),
        SavedPlace(name: "Fushimi Inari", subtitle: "Kyoto · saved from spring search", media: media("twilight", "Torii gates"))
    ]

    static let comparison = DestinationComparison(
        subtitle: "Madeira and Lisbon, on what you asked for",
        columns: [
            ComparisonColumn(name: "Madeira", matchPercent: 94, gist: "Quieter, greener, built around the coastline", media: media("madeira", "Madeira")),
            ComparisonColumn(name: "Lisbon", matchPercent: 91, gist: "Denser, more to eat, easier to reach", media: media("gold", "Lisbon"))
        ],
        metrics: [
            ComparisonMetric(key: "Estimated trip cost", left: "$2,780", right: "$2,410", winner: .right), ComparisonMetric(key: "Events this week", left: "", right: "", winner: .none),
            ComparisonMetric(key: "Weather in Oct", left: "Warm · 23°", right: "Warm · 24°", winner: .none),
            ComparisonMetric(key: "Flight time", left: "9h · 1 stop", right: "7h 20m · direct", winner: .right),
            ComparisonMetric(key: "Entry simplicity", left: "Same as Lisbon", right: "Same as Madeira", winner: .none),
            ComparisonMetric(key: "Crowds", left: "Quiet", right: "Busy in centre", winner: .left),
            ComparisonMetric(key: "Food", left: "Strong local", right: "Exceptional", winner: .right),
            ComparisonMetric(key: "Beach", left: "Coastal pools", right: "Day trip needed", winner: .left),
            ComparisonMetric(key: "Nature & walking", left: "Levadas, cliffs", right: "City parks", winner: .left),
            ComparisonMetric(key: "Nightlife", left: "Low key", right: "Wide choice", winner: .right)
        ],
        differences: [
            ComparisonDifference(name: "Madeira", text: "Better for relaxation, walking and scenery."),
            ComparisonDifference(name: "Lisbon", text: "Better for food, city activities and a shorter flight.")
        ]
    )

    static let flights = [
        FlightOption(tier: .bestOverall, airline: "TAP Air Portugal", departs: "22:10", arrives: "13:25", duration: "9h 15m", stopsLabel: "1 stop · LIS", stopCount: 1, price: PrototypeContent.usd(742), reason: "Best balance of price and travel time.", baggage: "1 checked bag included", checkedLabel: "Checked 6 min ago"),
        FlightOption(tier: .lowestPrice, airline: "Iberia + Binter", departs: "18:40", arrives: "16:05", duration: "15h 25m", stopsLabel: "2 stops · MAD, LPA", stopCount: 2, price: PrototypeContent.usd(611), reason: "Cheapest option, but a long overnight layover.", baggage: "Cabin bag only", checkedLabel: "Checked 6 min ago"),
        FlightOption(tier: .easiestJourney, airline: "TAP Air Portugal", departs: "20:55", arrives: "11:40", duration: "8h 45m", stopsLabel: "1 stop · LIS · 1h 10m", stopCount: 1, price: PrototypeContent.usd(806), reason: "Shortest total time and a same-terminal connection.", baggage: "1 checked bag included", checkedLabel: "Checked 6 min ago")
    ]

    static let heroExperience = HeroExperience(name: "West Madeira Island Tour", subtitle: "Full day · Câmara de Lobos, Cabo Girão, Porto Moniz", durationLabel: "~8h", fromPriceLabel: "$68 pp", rating: 4.8, fitTitle: "Fits best on Thursday", fitBody: "Good weather that day and no conflicting reservations. Your Friday is already a full outdoor day.", media: media("twilight", "West Madeira coastline"))

    static let experiences = [
        Experience(name: "Levada das 25 Fontes", meta: "Half day · from $42 pp", fit: "Best on Friday morning", media: media("guanacaste", "Levada walk")),
        Experience(name: "Funchal food walk", meta: "3h · from $58 pp", fit: "Pairs with your Day 2 market stop", media: media("gold", "Food walk")),
        Experience(name: "Dolphin watching", meta: "2h 30m · from $46 pp", fit: "Calm seas forecast midweek", media: media("curacao", "Dolphins"))
    ]

    static let transport = TransportGuide(
        recommendationTitle: "Rental car for Days 3–5",
        recommendationBody: "Your itinerary includes several places outside Funchal where bus service is limited. Walking covers the rest of the week.",
        arrivals: [
            ArrivalOption(name: "Airport transfer", meta: "Pre-booked · 25 min · door to door", costLabel: "$38", tag: .recommended, placeholder: "curacao"),
            ArrivalOption(name: "Rental car", meta: "25 min · needed Days 3–5 anyway", costLabel: "$34/day", tag: .bestForPlan, placeholder: "sky"),
            ArrivalOption(name: "Taxi", meta: "25 min · metered, no booking", costLabel: "$45–55", placeholder: "gold"),
            ArrivalOption(name: "Public bus", meta: "55 min · line 113, every 30 min", costLabel: "$5", placeholder: "twilight")
        ],
        transferLabel: "25 min · 22 km",
        airport: GeoCoordinate(latitude: 32.6942, longitude: -16.7745),
        hotel: GeoCoordinate(latitude: 32.6497, longitude: -16.9086),
        local: [
            LocalTransport(name: "Walking", note: "Funchal centre and Old Town are compact and walkable.", placeholder: "madeira"),
            LocalTransport(name: "Bus (Horários do Funchal)", note: "Good along the south coast, thin inland and after 20:00.", placeholder: "twilight"),
            LocalTransport(name: "Taxi", note: "Easy in Funchal, harder to find in villages — book a return.", placeholder: "gold"),
            LocalTransport(name: "Rental car", note: "The only practical way to reach Porto Moniz and the levadas.", placeholder: "sky")
        ]
    )

    static let checklist = [
        ChecklistGroup(name: "Documents", items: [ChecklistItem(name: "Passport", note: "Valid through Apr 2029", isDone: true), ChecklistItem(name: "Flight confirmation", isDone: true), ChecklistItem(name: "Hotel confirmation", isDone: true), ChecklistItem(name: "Travel insurance", note: "Optional for EU travel")]),
        ChecklistGroup(name: "Pack", items: [ChecklistItem(name: "Light rain layer", note: "Suggested for your dates"), ChecklistItem(name: "Walking shoes", note: "Four outdoor days", isDone: true), ChecklistItem(name: "Swimwear", isDone: true), ChecklistItem(name: "Power adapter", note: "Type C / F", isDone: true)]),
        ChecklistGroup(name: "Before you go", items: [ChecklistItem(name: "Download offline trip", isDone: true), ChecklistItem(name: "Set up eSIM"), ChecklistItem(name: "Confirm airport transfer", note: "Oct 12 · 17:05", isDone: true)])
    ]

    static let checklistInsight = "Bring a light rain layer. Showers are common in Madeira during your dates and four of your days include outdoor plans."

    static let members = [
        GroupMember(name: "You", subtitle: "alex@example.com", role: .owner, initial: "A", placeholder: "twilight"),
        GroupMember(name: "Maya", subtitle: "Invited Sep 2", role: .editor, initial: "M", placeholder: "curacao"),
        GroupMember(name: "Chris", subtitle: "Invited Sep 4", role: .viewer, initial: "C", placeholder: "gold")
    ]

    static let decisions = [
        GroupDecision(question: "Where should we stay?", meta: "3 options · 3 of 3 voted", options: [DecisionOption(name: "Quinta do Monte", votes: 2, hasMyVote: true), DecisionOption(name: "Hotel do Carmo", votes: 1, hasMyVote: false), DecisionOption(name: "Estalagem da Ponta", votes: 0, hasMyVote: false)]),
        GroupDecision(question: "Which tour on Thursday?", meta: "2 options · 3 of 3 voted", options: [DecisionOption(name: "West Island Tour", votes: 2, hasMyVote: false), DecisionOption(name: "Food Walking Tour", votes: 2, hasMyVote: true)])
    ]

    static let changes = [
        GroupChange(text: "Maya moved dinner to 8 PM", when: "2h ago", tone: .aqua),
        GroupChange(text: "You saved Quinta do Monte", when: "Yesterday", tone: .sky),
        GroupChange(text: "Chris joined the trip", when: "Sep 4", tone: .neutral)
    ]

    static let entryRows = [
        EntryRequirementRow(key: "Passport", value: "Must be valid for your full stay. Yours expires Apr 2029.", state: .met),
        EntryRequirementRow(key: "Visa", value: "Depends on your passport. US passports: no visa for short stays.", state: .info),
        EntryRequirementRow(key: "Return ticket", value: "Officers may ask to see onward travel. Yours is booked.", state: .met)
    ]

    static let infoSections = [
        InfoSection(name: "Money", gist: "Euro · cards accepted almost everywhere", rows: [InfoRow(key: "Currency", value: "Euro (€)"), InfoRow(key: "Cards", value: "Widely accepted"), InfoRow(key: "Cash", value: "Useful in small villages"), InfoRow(key: "ATMs", value: "Common in Funchal"), InfoRow(key: "Tipping", value: "Round up; 5–10% for good service")], source: "Source: local banking data · checked Sep 1", placeholder: "gold"),
        InfoSection(name: "Connectivity", gist: "eSIM available · EU roaming applies", rows: [InfoRow(key: "eSIM", value: "Supported, ~$18 for 7 days"), InfoRow(key: "Roaming", value: "EU rules apply to EU plans"), InfoRow(key: "Wi-Fi", value: "Free in most cafés and hotels")], source: "Source: carrier data · checked Sep 1", placeholder: "sky"),
        InfoSection(name: "Local basics", gist: "Portuguese · Type C/F · WET", rows: [InfoRow(key: "Language", value: "Portuguese; English widely spoken"), InfoRow(key: "Plug", value: "Type C / F"), InfoRow(key: "Voltage", value: "230 V, 50 Hz"), InfoRow(key: "Time zone", value: "WET (UTC+0), same as Lisbon")], source: "Source: reference data", placeholder: "curacao"),
        InfoSection(name: "Safety", gist: "Low risk · normal precautions", rows: [InfoRow(key: "General", value: "Low crime; usual care with valuables"), InfoRow(key: "Hiking", value: "Levadas can be slippery — check trail status"), InfoRow(key: "Sea", value: "Swim at marked areas; Atlantic swell is strong")], source: "Source: official advisories · checked Sep 1", placeholder: "guanacaste"),
        InfoSection(name: "Customs", gist: "Greet, queue, don’t rush meals", rows: [InfoRow(key: "Greeting", value: "Handshake; two cheek kisses among friends"), InfoRow(key: "Dining", value: "Meals are unhurried; ask for the bill"), InfoRow(key: "Dress", value: "Smart casual in restaurants")], source: "Source: editorial", placeholder: "madeira"),
        InfoSection(name: "Health", gist: "EHIC/GHIC accepted · pharmacies plentiful", rows: [InfoRow(key: "Pharmacies", value: "Farmácias, green cross sign"), InfoRow(key: "Tap water", value: "Safe to drink"), InfoRow(key: "Insurance", value: "Recommended for outdoor activities")], source: "Source: official health guidance · checked Sep 1", placeholder: "twilight")
    ]

    static let emergency = [EmergencyContact(key: "All services", value: "112"), EmergencyContact(key: "Medical", value: "112"), EmergencyContact(key: "Non-urgent police", value: "291 208 400")]

    static let profile = TravelerProfile(
        name: "Alex Rivera", initial: "A", homeLabel: "New York · JFK · 4 trips",
        styleChips: [StyleChip(name: "Beach", tone: .aqua), StyleChip(name: "Food", tone: .gold), StyleChip(name: "Culture", tone: .twilight), StyleChip(name: "Relaxed pace", tone: .sky), StyleChip(name: "Mild to warm", tone: .gold)],
        preferences: [
            PreferenceRow(name: "Interests", value: "Beach, Food, Culture", placeholder: "curacao"), PreferenceRow(name: "Travel pace", value: "Relaxed", placeholder: "madeira"),
            PreferenceRow(name: "Weather", value: "Mild to warm", placeholder: "gold"), PreferenceRow(name: "Crowds", value: "Avoid very busy places", placeholder: "twilight"),
            PreferenceRow(name: "Flight preferences", value: "Fewer stops, no red-eyes", placeholder: "sky"), PreferenceRow(name: "Stay preferences", value: "Quiet neighbourhoods, 4★", placeholder: "twilight"),
            PreferenceRow(name: "Budget style", value: "Balanced", placeholder: "gold"), PreferenceRow(name: "Travel group", value: "2 adults, no children", placeholder: "curacao"),
            PreferenceRow(name: "Accessibility", value: "No requirements set", placeholder: "guanacaste")
        ],
        learned: [LearnedInsight(text: "You often save coastal destinations", source: "From 6 saves"), LearnedInsight(text: "You usually prefer quieter neighbourhoods", source: "From 3 trips"), LearnedInsight(text: "You rarely book anything before 9 AM", source: "From 2 trips")],
        usesTripHistory: true
    )

    static let region = RegionSettings(appLanguage: "English", spokenLanguages: "English, French", region: "United States", regionNote: "Used for pricing and availability, not nationality", currency: "USD — US Dollar", homeAirport: "New York · JFK")

    static let settingsGroups = [
        SettingsGroup(name: "Account", rows: [SettingsRow(name: "Account", value: "alex@example.com"), SettingsRow(name: "Connected sign-in", value: "Apple"), SettingsRow(name: "Devices", value: "2 devices")]),
        SettingsGroup(name: "Privacy", rows: [SettingsRow(name: "Location", value: "While using"), SettingsRow(name: "Personalization", value: "On"), SettingsRow(name: "Trip history", value: "4 trips"), SettingsRow(name: "NUVRO data controls", value: "Manage"), SettingsRow(name: "Download my data"), SettingsRow(name: "Delete account", isDestructive: true)]),
        SettingsGroup(name: "Security", rows: [SettingsRow(name: "Sign-in & security"), SettingsRow(name: "Require Face ID for documents", value: "Off")])
    ]

    static let notifications = [
        NotificationGroup(name: "Trip alerts", settings: ["Flight changes", "Weather", "Reservations", "Leave reminders"].map { NotificationSetting(name: $0, isOn: true) }),
        NotificationGroup(name: "Planning", settings: ["Price changes", "Upcoming tasks"].map { NotificationSetting(name: $0, isOn: true) }),
        NotificationGroup(name: "NUVRO", settings: [NotificationSetting(name: "Product updates", isOn: false)])
    ]

    static let permissions = [
        PermissionStatus(name: "Location", value: "While Using", tone: .granted), PermissionStatus(name: "Notifications", value: "On", tone: .granted),
        PermissionStatus(name: "Camera", value: "Ask when needed", tone: .ask), PermissionStatus(name: "Photos", value: "Not requested", tone: .none)
    ]
}

extension PrototypeLibrary {
    static let regionChoices = RegionChoices(
        regions: ["United States", "United Kingdom", "Canada", "Nigeria", "Portugal", "Spain", "France", "Germany", "United Arab Emirates", "Japan", "Australia"].map { PickerOption($0) },
        currencies: [("USD", "US Dollar"), ("EUR", "Euro"), ("GBP", "British Pound"), ("CAD", "Canadian Dollar"), ("NGN", "Nigerian Naira"), ("AED", "UAE Dirham"), ("JPY", "Japanese Yen"), ("AUD", "Australian Dollar")].map { PickerOption(id: $0.0, label: $0.0, detail: $0.1) },
        airports: PrototypeContent.airports
    )

    static let preferenceChoices: [String: PreferenceChoices] = [
        "Interests": PreferenceChoices(options: ["Beach", "Food", "Culture", "Nature", "Nightlife", "Adventure", "Relaxation", "Family"].map { PickerOption($0) }, allowsMultiple: true),
        "Travel pace": PreferenceChoices(options: [PickerOption("Relaxed", detail: "Slow mornings, fewer stops"), PickerOption("Balanced"), PickerOption("Packed", detail: "See as much as possible")]),
        "Weather": PreferenceChoices(options: ["Warm", "Mild to warm", "Mild", "Cool"].map { PickerOption($0) }),
        "Crowds": PreferenceChoices(options: ["Avoid very busy places", "Don’t mind crowds", "Prefer lively places"].map { PickerOption($0) }),
        "Flight preferences": PreferenceChoices(options: ["Fewer stops, no red-eyes", "Cheapest, any stops", "Direct only"].map { PickerOption($0) }),
        "Stay preferences": PreferenceChoices(options: ["Quiet neighbourhoods, 4★", "Central, any rating", "Premium only"].map { PickerOption($0) }),
        "Budget style": PreferenceChoices(options: ["Budget", "Balanced", "Premium"].map { PickerOption($0) }),
        "Accessibility": PreferenceChoices(options: ["Step-free access", "Fewer walking distances", "Hearing support", "Visual support"].map { PickerOption($0) }, allowsMultiple: true)
    ]
}
