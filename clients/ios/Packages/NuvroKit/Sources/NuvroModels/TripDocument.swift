import Foundation

public struct TripDocument: Sendable, Hashable, Codable, Identifiable {
    public enum Category: String, Sendable, Codable, CaseIterable {
        case flight, stay, activity, transfer, other
    }

    public let id: String
    public let category: Category
    public let title: String
    public let subtitle: String
    public let whenLabel: String
    public let travelers: Int
    public let isAvailableOffline: Bool
    public let reference: String
    public let canAddToWallet: Bool
    public let isNeededNext: Bool
    public let isExpired: Bool
    public let hasBarcode: Bool

    public init(id: String, category: Category, title: String, subtitle: String, whenLabel: String, travelers: Int, isAvailableOffline: Bool, reference: String, canAddToWallet: Bool = false, isNeededNext: Bool = false, isExpired: Bool = false, hasBarcode: Bool = true) {
        self.id = id
        self.category = category
        self.title = title
        self.subtitle = subtitle
        self.whenLabel = whenLabel
        self.travelers = travelers
        self.isAvailableOffline = isAvailableOffline
        self.reference = reference
        self.canAddToWallet = canAddToWallet
        self.isNeededNext = isNeededNext
        self.isExpired = isExpired
        self.hasBarcode = hasBarcode
    }

    public var maskedReference: String {
        String(reference.map { $0.isLetter || $0.isNumber ? "•" : $0 })
    }

    private enum CodingKeys: String, CodingKey {
        case id, category, title, subtitle, travelers, reference
        case whenLabel = "when_label"
        case isAvailableOffline = "is_available_offline"
        case canAddToWallet = "can_add_to_wallet"
        case isNeededNext = "is_needed_next"
        case isExpired = "is_expired"
        case hasBarcode = "has_barcode"
    }
}
