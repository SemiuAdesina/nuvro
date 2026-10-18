import Foundation

public struct Money: Sendable, Hashable, Codable {
    public enum DecodingFailure: Error, Equatable, Sendable {
        case malformedCurrency(String)
    }

    public let amountMinor: Int64
    public let currency: String

    private enum CodingKeys: String, CodingKey {
        case amountMinor = "amount_minor"
        case currency
    }

    public init(amountMinor: Int64, currency: String) throws {
        guard Self.isISO4217Alphabetic(currency) else {
            throw DecodingFailure.malformedCurrency(currency)
        }
        self.amountMinor = amountMinor
        self.currency = currency
    }

    public init(from decoder: any Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        let currency = try container.decode(String.self, forKey: .currency)
        guard Self.isISO4217Alphabetic(currency) else {
            throw DecodingError.dataCorruptedError(
                forKey: .currency,
                in: container,
                debugDescription: "Currency must be three uppercase letters."
            )
        }
        self.amountMinor = try container.decode(Int64.self, forKey: .amountMinor)
        self.currency = currency
    }

    public func formatted(locale: Locale) -> String {
        let style = FloatingPointFormatStyle<Double>.Currency(code: currency, locale: locale)
        return majorUnits.formatted(style)
    }

    public func formattedWhole(locale: Locale) -> String {
        let style = FloatingPointFormatStyle<Double>.Currency(code: currency, locale: locale).precision(.fractionLength(0))
        return majorUnits.formatted(style)
    }

    public var majorUnits: Double {
        Double(amountMinor) / Double(Self.minorUnitDivisor(for: currency))
    }

    private static func minorUnitDivisor(for currency: String) -> Int {
        zeroDecimalCurrencies.contains(currency) ? 1 : 100
    }

    private static let zeroDecimalCurrencies: Set<String> = [
        "BIF", "CLP", "DJF", "GNF", "ISK", "JPY", "KMF", "KRW",
        "PYG", "RWF", "UGX", "UYI", "VND", "VUV", "XAF", "XOF", "XPF"
    ]

    private static func isISO4217Alphabetic(_ candidate: String) -> Bool {
        candidate.count == 3 && candidate.allSatisfy { $0.isASCII && $0.isUppercase && $0.isLetter }
    }
}
