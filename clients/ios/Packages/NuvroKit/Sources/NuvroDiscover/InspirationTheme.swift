import Foundation
import NuvroModels

public enum InspirationTheme: String, CaseIterable, Sendable, Identifiable {
    case beach = "Beach"
    case city = "City"
    case nature = "Nature"
    case food = "Food"
    case adventure = "Adventure"
    case relax = "Relax"
    case family = "Family"
    case hiddenGems = "Hidden gems"

    public var id: String { rawValue }
    public var title: String { rawValue }
}
