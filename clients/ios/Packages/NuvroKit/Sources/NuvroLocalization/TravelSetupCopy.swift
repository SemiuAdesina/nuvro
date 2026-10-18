public struct TravelSetupCopy: Equatable, Sendable {
    public let progressOf: String
    public let back: String
    public let skip: String
    public let continueLabel: String
    public let done: String
    public let change: String
    public let changeAnytime: String
    public let welcomeTitle: String
    public let welcomeBody: String
    public let residenceQuestion: String
    public let residenceBody: String
    public let residencePrivacy: String
    public let chooseCountry: String
    public let searchCountries: String
    public let passportQuestion: String
    public let passportBody: String
    public let passportPrivacy: String
    public let addPassport: String
    public let addAnotherPassport: String
    public let declinePassport: String
    public let passportDeclinedNote: String
    public let permitQuestion: String
    public let permitBody: String
    public let permitYes: String
    public let permitNo: String
    public let permitSkip: String
    public let addPermit: String
    public let airportQuestion: String
    public let airportBody: String
    public let chooseAirport: String
    public let searchAirports: String
    public let currencyQuestion: String
    public let currencyBody: String
    public let currencyPrivacy: String
    public let suggestedForYou: String
    public let chooseCurrency: String
    public let searchCurrencies: String
    public let paperworkQuestion: String
    public let paperworkNote: String
    public let interestsQuestion: String
    public let paceQuestion: String
    public let weatherQuestion: String
    public let crowdsQuestion: String
    public let flightsQuestion: String
    public let staysQuestion: String
    public let groupQuestion: String
    public let accessibilityQuestion: String
    public let accessibilityBody: String
    public let reviewTitle: String
    public let reviewBody: String
    public let finish: String
    public let noMatches: String
    public let noMatchesBody: String

    public static let english = TravelSetupCopy(
        progressOf: "of", back: "Back", skip: "Skip for now", continueLabel: "Continue", done: "Done", change: "Change", changeAnytime: "You can change these anytime.",
        welcomeTitle: "Let’s make NUVRO yours.", welcomeBody: "A few travel details help us recommend places that make sense for you.",
        residenceQuestion: "Where do you live?", residenceBody: "This helps with departure, currency, and local travel context.", residencePrivacy: "Used for travel context and nearby departures.",
        chooseCountry: "Choose a country", searchCountries: "Search countries",
        passportQuestion: "Which passport do you travel with?", passportBody: "We use this to understand entry requirements for international trips.",
        passportPrivacy: "We only need the country, not your passport number.", addPassport: "Add passport", addAnotherPassport: "Add another passport",
        declinePassport: "I don’t want to add this now", passportDeclinedNote: "Entry guidance stays general until you add a passport country.",
        permitQuestion: "Do you have another residence permit or long-term travel document?", permitBody: "This can make some destinations easier to enter.",
        permitYes: "Yes", permitNo: "No", permitSkip: "Skip for now", addPermit: "Add a country or region",
        airportQuestion: "Where do you usually fly from?", airportBody: "We’ll use this when estimating flights and realistic destinations.", chooseAirport: "Choose an airport", searchAirports: "City, airport or code",
        currencyQuestion: "What currency do you plan in?", currencyBody: "NUVRO will show budgets and trip estimates in this currency.", currencyPrivacy: "Used to show budgets in the currency you prefer.",
        suggestedForYou: "Suggested for you", chooseCurrency: "Choose a currency", searchCurrencies: "Search currencies",
        paperworkQuestion: "How much travel paperwork are you okay with?", paperworkNote: "A preference only. NUVRO always checks real entry requirements before you book.",
        interestsQuestion: "What makes a great trip for you?", paceQuestion: "How do you like your days?", weatherQuestion: "What weather do you enjoy?",
        crowdsQuestion: "How do you feel about busy places?", flightsQuestion: "What makes a flight work for you?", staysQuestion: "Where do you like to stay?",
        groupQuestion: "Who do you usually travel with?", accessibilityQuestion: "Anything we should consider for accessibility?", accessibilityBody: "Optional. It shapes what NUVRO suggests, nothing else.",
        reviewTitle: "You’re ready.", reviewBody: "Here’s what NUVRO will use. Tap any line to change it.", finish: "Show me the world",
        noMatches: "No matches", noMatchesBody: "Try a different spelling."
    )

    public static func resolve(_ language: Language) -> TravelSetupCopy { english }
}

public struct TravelOptionCopy: Equatable, Sendable {
    public let keepSimple: String
    public let keepSimpleDetail: String
    public let evisaFine: String
    public let evisaFineDetail: String
    public let anyDestination: String
    public let anyDestinationDetail: String
    public let relaxed: String
    public let relaxedDetail: String
    public let balanced: String
    public let balancedDetail: String
    public let full: String
    public let fullDetail: String
    public let warm: String
    public let mild: String
    public let cool: String
    public let cold: String
    public let noPreference: String
    public let preferQuiet: String
    public let crowdsBalanced: String
    public let crowdsFine: String
    public let fewerStops: String
    public let avoidOvernight: String
    public let shortest: String
    public let lowestPrice: String
    public let quietNeighbourhoods: String
    public let central: String
    public let nearNightlife: String
    public let nearBeach: String
    public let hotel: String
    public let apartment: String
    public let resort: String
    public let budgetFriendly: String
    public let premium: String
    public let solo: String
    public let partner: String
    public let friends: String
    public let family: String
    public let itChanges: String
    public let noRequirements: String
    public let mobility: String
    public let wheelchair: String
    public let reducedWalking: String
    public let vision: String
    public let hearing: String
    public let other: String
    public let livesIn: String
    public let travelsWith: String
    public let permits: String
    public let leavesFrom: String
    public let currency: String
    public let entryPreference: String
    public let travelStyle: String
    public let pace: String
    public let weather: String
    public let crowds: String
    public let flights: String
    public let stays: String
    public let group: String
    public let accessibility: String
    public let notAdded: String
    public let passport: String
    public let passports: String

    public static let english = TravelOptionCopy(
        keepSimple: "Keep it simple", keepSimpleDetail: "Prioritize destinations with easier entry.", evisaFine: "eVisa is fine", evisaFineDetail: "I’m okay applying online before I travel.",
        anyDestination: "Any destination", anyDestinationDetail: "Show me great options even if more preparation is required.",
        relaxed: "Relaxed", relaxedDetail: "A few things, plenty of breathing room.", balanced: "Balanced", balancedDetail: "A good mix of activities and downtime.", full: "Full", fullDetail: "Make the most of every day.",
        warm: "Warm", mild: "Mild", cool: "Cool", cold: "Cold / Snow", noPreference: "No preference",
        preferQuiet: "Prefer quiet", crowdsBalanced: "Balanced", crowdsFine: "Crowds are fine",
        fewerStops: "Prefer fewer stops", avoidOvernight: "Avoid overnight flights", shortest: "Shortest journey", lowestPrice: "Lowest price",
        quietNeighbourhoods: "Quiet neighbourhoods", central: "Central location", nearNightlife: "Near nightlife", nearBeach: "Near beach", hotel: "Hotel", apartment: "Apartment", resort: "Resort", budgetFriendly: "Budget-friendly", premium: "Premium",
        solo: "Solo", partner: "Partner", friends: "Friends", family: "Family", itChanges: "It changes",
        noRequirements: "No requirements", mobility: "Mobility", wheelchair: "Wheelchair access", reducedWalking: "Reduced walking", vision: "Vision", hearing: "Hearing", other: "Other",
        livesIn: "Lives in", travelsWith: "Travels with", permits: "Also holds", leavesFrom: "Usually leaves from", currency: "Currency", entryPreference: "Entry preference", travelStyle: "Travel style",
        pace: "Pace", weather: "Weather", crowds: "Crowds", flights: "Flights", stays: "Stays", group: "Travels", accessibility: "Accessibility", notAdded: "Not added yet", passport: "passport", passports: "passports"
    )

    public static func resolve(_ language: Language) -> TravelOptionCopy { english }
}
