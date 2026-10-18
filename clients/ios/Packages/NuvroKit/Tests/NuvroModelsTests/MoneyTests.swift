import Foundation
import Testing

@testable import NuvroModels

@Suite("Money")
struct MoneyTests {
    private let decoder = JSONDecoder()
    private let encoder = JSONEncoder()

    @Test("decodes the contract payload without losing precision")
    func decodesContractPayload() throws {
        let payload = Data(#"{"amount_minor":12550,"currency":"USD"}"#.utf8)

        let money = try decoder.decode(Money.self, from: payload)

        #expect(money.amountMinor == 12550)
        #expect(money.currency == "USD")
    }

    @Test("rejects a malformed currency code during decoding")
    func rejectsMalformedCurrency() {
        for raw in ["usd", "US", "USDD", "US1"] {
            let payload = Data(#"{"amount_minor":1,"currency":"\#(raw)"}"#.utf8)
            #expect(throws: DecodingError.self) {
                _ = try decoder.decode(Money.self, from: payload)
            }
        }
    }

    @Test("rejects a malformed currency code during construction")
    func rejectsMalformedCurrencyOnInit() {
        #expect(throws: Money.DecodingFailure.malformedCurrency("eur")) {
            _ = try Money(amountMinor: 1, currency: "eur")
        }
    }

    @Test("round trips through the contract field names")
    func roundTrips() throws {
        let original = try Money(amountMinor: -4999, currency: "NGN")

        let encoded = try encoder.encode(original)
        let rendered = String(decoding: encoded, as: UTF8.self)

        #expect(rendered.contains("amount_minor"))
        #expect(try decoder.decode(Money.self, from: encoded) == original)
    }

    @Test("formats two decimal currencies for the traveller locale")
    func formatsTwoDecimalCurrency() throws {
        let money = try Money(amountMinor: 12550, currency: "USD")

        let formatted = money.formatted(locale: Locale(identifier: "en_US"))

        #expect(formatted.contains("125"))
        #expect(formatted.contains("50"))
    }

    @Test("formats zero decimal currencies without a fractional part")
    func formatsZeroDecimalCurrency() throws {
        let money = try Money(amountMinor: 12550, currency: "JPY")

        let formatted = money.formatted(locale: Locale(identifier: "en_US"))

        #expect(formatted.contains("12,550"))
    }

    @Test("large amounts survive the Int64 contract range")
    func handlesLargeAmounts() throws {
        let payload = Data(#"{"amount_minor":9223372036854775807,"currency":"JPY"}"#.utf8)

        let money = try decoder.decode(Money.self, from: payload)

        #expect(money.amountMinor == Int64.max)
    }
}
