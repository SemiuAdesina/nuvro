import NuvroModels

enum PrototypeItinerary {
    static let days: [ItineraryDay] = [
        ItineraryDay(dayNumber: 1, title: "Day 1 — Arrival", summary: "Arrive · Funchal · Easy evening", activities: [
            ItineraryActivity(time: "16:40", title: "Land at Funchal (FNC)", detail: "Flight TP1691 · terminal exit"),
            ItineraryActivity(time: "17:30", title: "Check in at Quinta do Monte", detail: "Funchal · 900 m from Old Town", photoReference: "gold", status: .booked, travelToNext: "25 min by car"),
            ItineraryActivity(time: "19:30", title: "Sunset walk on the promenade", detail: "Avenida do Mar · 40 min", travelToNext: "12 min walk", isOutdoor: true)
        ]),
        ItineraryDay(dayNumber: 2, title: "Day 2 — Funchal", summary: "Old Town · Food market · Cable car", departureTravel: "6 min walk", activities: [
            ItineraryActivity(time: "08:30", title: "Breakfast near hotel", detail: "Café do Museu · 40 min", travelToNext: "6 min walk"),
            ItineraryActivity(time: "10:00", title: "Funchal Old Town", detail: "Rua de Santa Maria · street art", photoReference: "madeira", travelToNext: "4 min walk", isOutdoor: true),
            ItineraryActivity(time: "12:30", title: "Mercado dos Lavradores", detail: "Fruit, fish and flowers", photoReference: "gold", travelToNext: "18 min by car"),
            ItineraryActivity(time: "14:30", title: "Funchal Cable Car", detail: "To Monte · book ahead", photoReference: "sky", status: .ticket, travelToNext: "12 min walk"),
            ItineraryActivity(time: "17:00", title: "Hotel / free time", detail: "Rest before dinner", travelToNext: "9 min walk"),
            ItineraryActivity(time: "19:30", title: "Dinner at Armazém do Sal", detail: "Reservation for 2", status: .reserved)
        ]),
        ItineraryDay(dayNumber: 3, title: "Day 3 — Monte & gardens", summary: "West Island experience", departureTravel: "12 min walk", activities: [
            ItineraryActivity(time: "10:00", title: "Funchal Cable Car", detail: "To Monte · 15 min ride", photoReference: "sky", status: .ticket, travelToNext: "12 min walk"),
            ItineraryActivity(time: "12:30", title: "Lunch at Alto Monte", detail: "Reservation for 2", status: .reserved, travelToNext: "10 min walk"),
            ItineraryActivity(time: "15:00", title: "Botanical Garden", detail: "Outdoor · rain expected", photoReference: "curacao", status: .weather, travelToNext: "25 min by car", isOutdoor: true),
            ItineraryActivity(time: "19:30", title: "Dinner in Old Town", detail: "Table for 2 at 19:30", status: .reserved)
        ]),
        ItineraryDay(dayNumber: 4, title: "Day 4 — West island", summary: "Full-day drive", activities: [
            ItineraryActivity(time: "09:00", title: "Pick up rental car", detail: "Funchal city office", status: .booked, travelToNext: "5 min walk"),
            ItineraryActivity(time: "10:30", title: "Cabo Girão skywalk", detail: "Europe’s highest sea cliff", photoReference: "twilight", travelToNext: "35 min by car", isOutdoor: true),
            ItineraryActivity(time: "13:00", title: "Lunch in Câmara de Lobos", detail: "Fishing village", travelToNext: "20 min by car"),
            ItineraryActivity(time: "16:00", title: "Calheta beach", detail: "Golden sand · swim", photoReference: "curacao", isOutdoor: true)
        ]),
        ItineraryDay(dayNumber: 5, title: "Day 5 — Levada walk", summary: "Outdoor day", activities: [
            ItineraryActivity(time: "08:00", title: "Transfer to Rabaçal", detail: "Mountain road", travelToNext: "55 min by car"),
            ItineraryActivity(time: "09:30", title: "Levada das 25 Fontes", detail: "Half-day walk · waterfalls", photoReference: "guanacaste", isOutdoor: true),
            ItineraryActivity(time: "15:00", title: "Back to Funchal", detail: "Rest and swim")
        ]),
        ItineraryDay(dayNumber: 6, title: "Day 6 — East coast", summary: "Coast and cliffs", activities: [
            ItineraryActivity(time: "09:30", title: "Ponta de São Lourenço", detail: "Coastal trail · 2h", photoReference: "twilight", travelToNext: "45 min by car", isOutdoor: true),
            ItineraryActivity(time: "13:30", title: "Lunch in Caniçal", detail: "Seafood", travelToNext: "15 min by car"),
            ItineraryActivity(time: "16:30", title: "Machico beach", detail: "Swim and coffee", photoReference: "curacao", isOutdoor: true)
        ]),
        ItineraryDay(dayNumber: 7, title: "Day 7 — Departure", summary: "Slow morning", activities: [
            ItineraryActivity(time: "09:00", title: "Breakfast and packing", detail: "Hotel"),
            ItineraryActivity(time: "11:00", title: "Last walk in Old Town", detail: "Souvenirs", photoReference: "madeira", travelToNext: "20 min by car", isOutdoor: true),
            ItineraryActivity(time: "14:15", title: "Flight home", detail: "TP1692 · check in 2h before", status: .booked)
        ])
    ]

    static let rainAdjustment = PlanAdjustment(
        trigger: "It’s raining this afternoon.",
        dayNumber: 3,
        before: ItineraryActivity(time: "15:00", title: "Botanical Garden", detail: "Outdoor · 25 min by car", photoReference: "curacao", status: .weather, travelToNext: "25 min by car", isOutdoor: true),
        after: ItineraryActivity(time: "15:00", title: "Madeira Story Centre", detail: "Indoor · 8 min walk", photoReference: "twilight", travelToNext: "8 min walk"),
        relocationNote: "Botanical Garden moved to tomorrow at 10:30",
        reasoning: "Rain is expected from 3–6 PM. The indoor museum is nearby, fits your interests, and keeps your dinner reservation unchanged.",
        preserved: ["Dinner reservation unchanged", "Budget unchanged", "Travel time still works", "Outdoor activity moved, not removed"]
    )
}
