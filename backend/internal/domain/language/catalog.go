package language

var essentialPhrases = []Phrase{
	{Key: "greeting_hello", Category: General, SourceText: "Hello"},
	{Key: "greeting_thanks", Category: General, SourceText: "Thank you"},
	{Key: "general_please", Category: General, SourceText: "Please"},
	{Key: "general_do_you_speak_english", Category: General, SourceText: "Do you speak English?"},
	{Key: "general_i_do_not_understand", Category: General, SourceText: "I do not understand"},
	{Key: "hotel_check_in", Category: Hotel, SourceText: "I have a reservation"},
	{Key: "hotel_late_checkout", Category: Hotel, SourceText: "Can I check out later?"},
	{Key: "hotel_luggage", Category: Hotel, SourceText: "Can I leave my luggage here?"},
	{Key: "restaurant_table_for", Category: Restaurant, SourceText: "A table for two, please"},
	{Key: "restaurant_menu", Category: Restaurant, SourceText: "Could I see the menu?"},
	{Key: "restaurant_bill", Category: Restaurant, SourceText: "The bill, please"},
	{Key: "restaurant_allergy", Category: Restaurant, SourceText: "I have a food allergy"},
	{Key: "transport_ticket", Category: Transport, SourceText: "One ticket, please"},
	{Key: "transport_taxi", Category: Transport, SourceText: "Please take me to this address"},
	{Key: "transport_when_departs", Category: Transport, SourceText: "When does it leave?"},
	{Key: "directions_where_is", Category: Directions, SourceText: "Where is this place?"},
	{Key: "directions_how_far", Category: Directions, SourceText: "How far is it?"},
	{Key: "directions_lost", Category: Directions, SourceText: "I am lost"},
	{Key: "shopping_how_much", Category: Shopping, SourceText: "How much does this cost?"},
	{Key: "shopping_card", Category: Shopping, SourceText: "Can I pay by card?"},
	{Key: "emergency_help", Category: Emergency, SourceText: "I need help"},
	{Key: "emergency_call_services", Category: Emergency, SourceText: "Please call emergency services"},
	{Key: "emergency_police", Category: Emergency, SourceText: "Please call the police"},
	{Key: "medical_need_doctor", Category: Medical, SourceText: "I need a doctor"},
	{Key: "medical_pharmacy", Category: Medical, SourceText: "Where is the nearest pharmacy?"},
	{Key: "medical_allergic", Category: Medical, SourceText: "I am allergic to this"},
}

func EssentialPhrases(categories ...Category) []Phrase {
	if len(categories) == 0 {
		return append([]Phrase{}, essentialPhrases...)
	}
	wanted := make(map[Category]struct{}, len(categories))
	for _, category := range categories {
		wanted[category] = struct{}{}
	}
	selected := make([]Phrase, 0, len(essentialPhrases))
	for _, phrase := range essentialPhrases {
		if _, found := wanted[phrase.Category]; found {
			selected = append(selected, phrase)
		}
	}
	return selected
}
