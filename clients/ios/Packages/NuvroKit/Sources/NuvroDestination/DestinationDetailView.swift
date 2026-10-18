import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

public struct DestinationDetailView: View {
    @Environment(\.designMetrics) private var metrics
    @State private var isSaved = false
    private let match: DestinationMatch
    private let guide: DestinationGuide
    private let budgetCeiling: Money?
    private let estimate: Money?
    private let copy: DestinationCopy
    private let locale: Locale
    private let onBuildTrip: () -> Void

    public init(match: DestinationMatch, guide: DestinationGuide, budgetCeiling: Money?, estimate: Money?, language: Language = .english, locale: Locale = .current, onBuildTrip: @escaping () -> Void) {
        self.match = match
        self.guide = guide
        self.budgetCeiling = budgetCeiling
        self.estimate = estimate
        self.copy = .resolve(language)
        self.locale = locale
        self.onBuildTrip = onBuildTrip
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 0) {
                DestinationHero(match: match)
                VStack(alignment: .leading, spacing: metrics.points(22)) {
                    FactsStrip(facts: DestinationFacts.glance(for: match, estimate: estimate, copy: copy, locale: locale))
                        .padding(.horizontal, -metrics.points(20))
                    WhyPickedCard(
                        title: copy.whyItFits,
                        summary: DestinationFacts.summary(for: match),
                        reasons: DestinationFacts.reasons(for: match, budgetCeiling: budgetCeiling, estimate: estimate, copy: copy, locale: locale)
                    )
                    CategoryGrid(title: "\(copy.everythingAbout) \(match.name)", onSelect: { _ in }).padding(.horizontal, -metrics.points(20))
                    AreasRow(title: copy.whereToStay, seeAll: copy.seeAll, areas: guide.areas).padding(.horizontal, -metrics.points(20))
                    ActivitiesList(title: copy.thingsToDo, seeAll: copy.seeAll, activities: guide.activities)
                    MapPreview(title: copy.onTheMap, openMap: copy.openMap, guide: guide)
                    InlineActions {
                        CallToActionButton(copy.buildTrip, action: onBuildTrip)
                        Button { isSaved.toggle() } label: {
                            Text(isSaved ? copy.saved : copy.save)
                                .typeStyle(Typography.secondaryAction)
                                .foregroundStyle(Palette.sky)
                                .frame(maxWidth: .infinity)
                                .frame(height: metrics.points(34))
                        }
                        .buttonStyle(.plain)
                    }
                }
                .padding(.horizontal, metrics.points(20))
                .padding(.top, metrics.points(6))
                .padding(.bottom, metrics.points(40))
            }
        }
        .scrollIndicators(.hidden)
        .ignoresSafeArea(edges: .top)
        .background(Palette.background.ignoresSafeArea())
        .modifier(PhotoNavigationChrome())
        .toolbar {
            ToolbarItem(placement: .principal) {
                HStack(spacing: metrics.points(6)) {
                    Image(systemName: match.weather == nil ? "cloud" : "sun.max.fill").foregroundStyle(Palette.gold)
                    Text(match.weather.map { "\($0.temperatureCelsius)° · \(weatherDetail)" } ?? weatherDetail)
                }
                .typeStyle(Typography.metaEmphasis)
                .foregroundStyle(.primary)
                .padding(.horizontal, metrics.points(6))
                .fixedSize()
            }
            ToolbarItem(placement: .primaryAction) {
                Button { isSaved.toggle() } label: {
                    Image(systemName: isSaved ? "bookmark.fill" : "bookmark").foregroundStyle(.primary)
                }
                .accessibilityLabel(isSaved ? copy.saved : copy.save)
            }
        }
    }
}

extension DestinationDetailView {
    var weatherDetail: String {
        guard let weather = match.weather else { return "Weather not yet known" }
        return weather.condition == "Warm" ? "Mostly sunny" : weather.condition
    }
}

struct PhotoNavigationChrome: ViewModifier {
    func body(content: Content) -> some View {
        #if os(iOS)
        content
            .toolbarBackground(.hidden, for: .navigationBar)
        #else
        content
        #endif
    }
}
