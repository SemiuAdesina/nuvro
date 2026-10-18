import Foundation

public struct ReadinessItem: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let isReady: Bool

    public init(name: String, isReady: Bool) {
        self.name = name
        self.isReady = isReady
    }

    public var id: String { name }

    private enum CodingKeys: String, CodingKey {
        case name
        case isReady = "is_ready"
    }
}

public struct BudgetDetail: Sendable, Hashable, Codable, Identifiable {
    public let label: String
    public let amount: Money

    public init(label: String, amount: Money) {
        self.label = label
        self.amount = amount
    }

    public var id: String { label }
}

public struct BudgetLine: Sendable, Hashable, Codable, Identifiable {
    public let category: String
    public let amount: Money
    public let booked: Money?
    public let details: [BudgetDetail]

    public init(category: String, amount: Money, booked: Money? = nil, details: [BudgetDetail] = []) {
        self.category = category
        self.amount = amount
        self.booked = booked
        self.details = details
    }

    public var id: String { category }
    public var isBooked: Bool { booked != nil }
}

public struct TripPlan: Sendable, Hashable, Codable, Identifiable {
    public let id: String
    public let destination: DestinationMatch
    public let dateRangeLabel: String
    public let travelersLabel: String
    public let startDate: Date
    public let budgetCeiling: Money
    public let readiness: [ReadinessItem]
    public let budgetLines: [BudgetLine]
    public let days: [ItineraryDay]

    public init(id: String, destination: DestinationMatch, dateRangeLabel: String, travelersLabel: String, startDate: Date, budgetCeiling: Money, readiness: [ReadinessItem], budgetLines: [BudgetLine], days: [ItineraryDay]) {
        self.id = id
        self.destination = destination
        self.dateRangeLabel = dateRangeLabel
        self.travelersLabel = travelersLabel
        self.startDate = startDate
        self.budgetCeiling = budgetCeiling
        self.readiness = readiness
        self.budgetLines = budgetLines
        self.days = days
    }

    private enum CodingKeys: String, CodingKey {
        case id, destination, readiness, days
        case dateRangeLabel = "date_range_label"
        case travelersLabel = "travelers_label"
        case startDate = "start_date"
        case budgetCeiling = "budget_ceiling"
        case budgetLines = "budget_lines"
    }

    public func date(forDay dayNumber: Int, calendar: Calendar = .current) -> Date {
        calendar.date(byAdding: .day, value: dayNumber - 1, to: startDate) ?? startDate
    }

    public func day(_ dayNumber: Int) -> ItineraryDay? {
        days.first { $0.dayNumber == dayNumber }
    }

    public func replacing(day: ItineraryDay) -> TripPlan {
        TripPlan(
            id: id, destination: destination, dateRangeLabel: dateRangeLabel, travelersLabel: travelersLabel, startDate: startDate,
            budgetCeiling: budgetCeiling, readiness: readiness, budgetLines: budgetLines,
            days: days.map { $0.dayNumber == day.dayNumber ? day : $0 }
        )
    }

    public var budgetFraction: Double {
        guard let total = estimatedTotal, budgetCeiling.amountMinor > 0 else { return 0 }
        return Double(total.amountMinor) / Double(budgetCeiling.amountMinor)
    }

    public var readyCount: Int { readiness.filter(\.isReady).count }
    public var readinessFraction: Double { readiness.isEmpty ? 0 : Double(readyCount) / Double(readiness.count) }

    public var estimatedTotal: Money? {
        guard let currency = budgetLines.first?.currency else { return nil }
        let total = budgetLines.reduce(Int64(0)) { $0 + $1.amount.amountMinor }
        return try? Money(amountMinor: total, currency: currency)
    }

    public var remainingBudget: Money? {
        guard let total = estimatedTotal, total.currency == budgetCeiling.currency else { return nil }
        return try? Money(amountMinor: budgetCeiling.amountMinor - total.amountMinor, currency: total.currency)
    }
}

private extension BudgetLine {
    var currency: String { amount.currency }
}
