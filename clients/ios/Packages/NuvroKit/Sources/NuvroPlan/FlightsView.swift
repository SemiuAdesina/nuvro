import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation
import SwiftUI

@Observable
@MainActor
public final class FlightsModel {
    public private(set) var state: LoadState<[FlightOption]>
    public var selectedTier: FlightOption.Tier = .bestOverall
    public var handoff: FlightOption?
    public private(set) var freshness: Freshness
    public let copy: FlightsCopy
    public let locale: Locale
    private let source: (@Sendable () async throws -> [FlightOption])?

    public init(options: [FlightOption], freshness: Freshness = .fresh, source: (@Sendable () async throws -> [FlightOption])? = nil, language: Language = .english, locale: Locale = .current) {
        self.state = .loaded(options)
        self.freshness = freshness
        self.source = source
        self.copy = .resolve(language)
        self.locale = locale
    }

    public init(state: LoadState<[FlightOption]>, freshness: Freshness = .unavailable, source: (@Sendable () async throws -> [FlightOption])? = nil, language: Language = .english, locale: Locale = .current) {
        self.state = state
        self.freshness = freshness
        self.source = source
        self.copy = .resolve(language)
        self.locale = locale
    }

    public var options: [FlightOption] { state.value ?? [] }

    public func refresh() async {
        guard let source else { return }
        state = state.beginning()
        do {
            state = .loaded(try await source())
            freshness = .fresh
        } catch {
            state = state.failing(LoadFailure(error: error))
            if state.value != nil { freshness = .stale }
        }
    }

    public func markStale() {
        freshness = .stale
    }

    public var selected: FlightOption? { options.first { $0.tier == selectedTier } }
    public var others: [FlightOption] { options.filter { $0.tier != selectedTier } }

    public func label(for tier: FlightOption.Tier) -> String {
        switch tier {
        case .bestOverall: copy.bestOverall
        case .lowestPrice: copy.lowestPrice
        case .easiestJourney: copy.easiestJourney
        }
    }

    public func price(_ option: FlightOption) -> String { option.price.formattedWhole(locale: locale) }
}

public struct FlightsView: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable private var model: FlightsModel
    private let stateCopy = StateCopy.resolve(.english)

    @ViewBuilder
    private var flightStates: some View {
        switch model.state {
        case .idle, .loading:
            NuvroLoadingState(stateCopy.loading)
        case .failed(let failure, let cached) where cached == nil:
            NuvroErrorState(
                symbol: failure == .offline ? "wifi.slash" : "airplane.circle",
                title: failure == .offline ? stateCopy.offlineTitle : model.copy.providerUnavailable,
                message: failure == .offline ? stateCopy.offlineBody : stateCopy.unavailableBody,
                retry: (stateCopy.retry, { Task { await model.refresh() } })
            )
        case .loaded(let options) where options.isEmpty:
            NuvroEmptyState(symbol: "airplane", title: model.copy.emptyTitle, message: model.copy.emptyBody)
        default:
            HStack {
                FreshnessLabel(tone: freshnessTone, text: freshnessText)
                Spacer()
                if model.state.isLoading { ProgressView().controlSize(.small).tint(Palette.sky) }
            }
        }
    }

    private var freshnessTone: FreshnessLabel.Tone {
        switch model.freshness {
        case .fresh: .fresh
        case .cached: .cached
        case .stale: .stale
        case .unavailable: .unavailable
        }
    }

    private var freshnessText: String {
        switch model.freshness {
        case .fresh: "\(model.copy.pricesChecked) · \(stateCopy.updatedJustNow.lowercased())"
        case .cached: stateCopy.cached
        case .stale: model.copy.pricesStale
        case .unavailable: stateCopy.unavailableShort
        }
    }
    private let routeLabel: String
    private let travelers: String

    public init(model: FlightsModel, routeLabel: String, travelers: String) {
        self.model = model
        self.routeLabel = routeLabel
        self.travelers = travelers
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                flightStates
                Text(travelers).typeStyle(Typography.metaEmphasis).foregroundStyle(Palette.labelAlpha(0.7)).padding(.horizontal, metrics.points(12)).frame(height: metrics.points(32)).background(Palette.labelAlpha(0.06), in: Capsule())
                tierChips
                if let flight = model.selected { FlightHeroCard(flight: flight, model: model).id(flight.id).transition(.opacity) }
                VStack(alignment: .leading, spacing: metrics.points(10)) {
                    Text(model.copy.otherOptions).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                    ForEach(model.others) { flight in
                        Button { model.selectedTier = flight.tier } label: {
                            HStack(spacing: metrics.points(12)) {
                                PhotoStandIn(palette: flight.tier == .lowestPrice ? .gold : .sky, label: "", labelSize: 6).frame(width: metrics.points(30), height: metrics.points(30)).clipShape(RoundedRectangle(cornerRadius: metrics.points(9), style: .continuous))
                                VStack(alignment: .leading, spacing: metrics.points(2)) {
                                    Text(model.label(for: flight.tier).uppercased()).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 0.66)).foregroundStyle(flight.tier == .lowestPrice ? Palette.amber : Palette.sky)
                                    Text(flight.airline).typeStyle(TypeStyle(size: 14.5, weight: .semibold)).foregroundStyle(Palette.label)
                                    Text("\(flight.departs) → \(flight.arrives) · \(flight.duration) · \(flight.stopsLabel)").typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                                }
                                Spacer(minLength: 0)
                                Text(model.price(flight)).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.3)).foregroundStyle(Palette.label)
                            }
                            .padding(EdgeInsets(top: metrics.points(13), leading: metrics.points(14), bottom: metrics.points(13), trailing: metrics.points(14)))
                            .surfaceCard(radius: Radius.cardCompact)
                        }
                        .buttonStyle(PressScaleButtonStyle(scale: 0.99))
                    }
                    Text(model.copy.disclaimer).typeStyle(TypeStyle(size: 11, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.45)).multilineTextAlignment(.center).frame(maxWidth: .infinity)
                }
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
            .animation(Motion.fade(0.3), value: model.selectedTier)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .refreshable { await model.refresh() }
        .screenTitle(model.copy.title, subtitle: routeLabel)
    }

    private var tierChips: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: metrics.points(8)) {
                ForEach(FlightOption.Tier.allCases, id: \.self) { tier in
                    let selected = model.selectedTier == tier
                    Button { model.selectedTier = tier } label: {
                        Text(model.label(for: tier)).typeStyle(Typography.control).foregroundStyle(selected ? Palette.background : Palette.label)
                            .padding(.horizontal, metrics.points(14)).frame(height: metrics.points(36))
                            .background(selected ? Palette.label : Palette.surface, in: Capsule())
                            .overlay(Capsule().strokeBorder(selected ? Palette.label : Palette.labelAlpha(0.1), lineWidth: 0.5))
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.96))
                    .accessibilityAddTraits(selected ? .isSelected : [])
                }
            }
            .padding(.horizontal, metrics.points(20))
        }
        .scrollClipDisabled()
        .padding(.horizontal, -metrics.points(20))
    }
}

struct FlightHeroCard: View {
    @Environment(\.designMetrics) private var metrics
    let flight: FlightOption
    let model: FlightsModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            HStack {
                Text(model.label(for: flight.tier).uppercased()).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.1)).foregroundStyle(Palette.matchTeal)
                    .padding(.horizontal, metrics.points(9)).padding(.vertical, metrics.points(4)).background(Palette.aqua.opacity(0.16), in: Capsule())
                Spacer()
                Text(flight.checkedLabel).typeStyle(TypeStyle(size: 11.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.5))
            }
            HStack(spacing: metrics.points(11)) {
                PhotoStandIn(palette: .sky, label: "", labelSize: 6).frame(width: metrics.points(34), height: metrics.points(34)).clipShape(RoundedRectangle(cornerRadius: metrics.points(10), style: .continuous))
                Text(flight.airline).typeStyle(TypeStyle(size: 16, weight: .semibold)).foregroundStyle(Palette.label)
            }
            HStack(spacing: metrics.points(12)) {
                VStack(alignment: .leading, spacing: 1) {
                    Text(flight.departs).typeStyle(TypeStyle(size: 21, weight: .bold, tracking: -0.4)).foregroundStyle(Palette.label)
                    Text("JFK").typeStyle(TypeStyle(size: 12, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.6))
                }
                VStack(spacing: metrics.points(3)) {
                    Text(flight.duration).typeStyle(TypeStyle(size: 11.5, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.6))
                    HStack(spacing: metrics.points(4)) {
                        Circle().fill(Palette.sky).frame(width: 6, height: 6)
                        Rectangle().fill(Palette.sky.opacity(0.3)).frame(height: 1.5)
                        if flight.stopCount > 0 {
                            Circle().fill(Palette.gold).frame(width: 5, height: 5)
                            Rectangle().fill(Palette.sky.opacity(0.3)).frame(height: 1.5)
                        }
                        Circle().fill(Palette.sky).frame(width: 6, height: 6)
                    }
                    Text(flight.stopsLabel).typeStyle(TypeStyle(size: 11.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.55))
                }
                .frame(maxWidth: .infinity)
                VStack(alignment: .trailing, spacing: 1) {
                    Text(flight.arrives).typeStyle(TypeStyle(size: 21, weight: .bold, tracking: -0.4)).foregroundStyle(Palette.label)
                    Text("FNC").typeStyle(TypeStyle(size: 12, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.6))
                }
            }
            Divider().overlay(Palette.labelAlpha(0.08))
            HStack(alignment: .bottom, spacing: metrics.points(12)) {
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    Text(flight.reason).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.62))
                    Text(flight.baggage).typeStyle(TypeStyle(size: 11.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.5))
                }
                Spacer()
                Text(model.price(flight)).typeStyle(TypeStyle(size: 24, weight: .bold, tracking: -0.5)).foregroundStyle(Palette.label)
            }
            CallToActionButton(model.copy.viewFlight) { model.handoff = flight }
        }
        .padding(metrics.points(16))
        .surfaceCard(radius: 22, elevated: true)
    }
}
