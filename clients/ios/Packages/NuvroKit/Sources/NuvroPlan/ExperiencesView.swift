import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation
import SwiftUI

@Observable
@MainActor
public final class ExperiencesModel {
    public let hero: HeroExperience
    public let experiences: [Experience]
    public var selectedTab: String
    public private(set) var isHeroAdded = false
    public let copy: ExperiencesCopy

    public init(hero: HeroExperience, experiences: [Experience], language: Language = .english) {
        self.hero = hero
        self.experiences = experiences
        self.copy = .resolve(language)
        self.selectedTab = ExperiencesCopy.resolve(language).tabs.first ?? ""
    }

    public func toggleHeroAdded() { isHeroAdded.toggle() }
}

public struct ExperiencesView: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable private var model: ExperiencesModel
    private let subtitle: String

    public init(model: ExperiencesModel, subtitle: String) {
        self.model = model
        self.subtitle = subtitle
    }

    private var copy: ExperiencesCopy { model.copy }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                VStack(alignment: .leading, spacing: metrics.points(8)) {
                    HStack(spacing: metrics.points(7)) {
                        Circle().fill(Palette.aqua).frame(width: metrics.points(6), height: metrics.points(6))
                        Text(copy.rankedFor).typeStyle(TypeStyle(size: 12.5, weight: .semibold)).foregroundStyle(Palette.matchTeal)
                    }
                    .padding(.horizontal, metrics.points(12)).frame(height: metrics.points(32)).background(Palette.aqua.opacity(0.16), in: Capsule())
                }
                tabs
                heroCard
                VStack(alignment: .leading, spacing: metrics.points(10)) {
                    HStack(alignment: .firstTextBaseline) {
                        Text(copy.moreForDates).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                        Spacer()
                        LinkButton(copy.seeAll) {}
                    }
                    if model.experiences.isEmpty {
                        NuvroEmptyState(symbol: "figure.hiking", title: copy.emptyTitle, message: copy.emptyBody)
                    }
                    ForEach(model.experiences) { experience in
                        HStack(spacing: metrics.points(12)) {
                            DynamicDestinationImage(media: experience.media.descriptor, labelSize: 6).frame(width: metrics.points(78), height: metrics.points(70)).clipShape(RoundedRectangle(cornerRadius: metrics.points(14), style: .continuous))
                            VStack(alignment: .leading, spacing: metrics.points(3)) {
                                Text(experience.name).typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(Palette.label)
                                Text(experience.meta).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                                Text(experience.fit).typeStyle(TypeStyle(size: 11.5, weight: .semibold)).foregroundStyle(Palette.matchTeal)
                            }
                            Spacer(minLength: 0)
                            ChevronGlyph()
                        }
                        .padding(metrics.points(10))
                        .surfaceCard(radius: Radius.cardCompact)
                    }
                    Text(copy.disclaimer).typeStyle(TypeStyle(size: 11, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.45)).multilineTextAlignment(.center).frame(maxWidth: .infinity)
                }
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(copy.title, subtitle: subtitle)
    }

    private var tabs: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: metrics.points(8)) {
                ForEach(copy.tabs, id: \.self) { tab in
                    let selected = model.selectedTab == tab
                    Button { model.selectedTab = tab } label: {
                        Text(tab).typeStyle(Typography.control).foregroundStyle(selected ? Palette.background : Palette.label)
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

    private var heroCard: some View {
        VStack(spacing: 0) {
            DynamicDestinationImage(media: model.hero.media.descriptor, label: "photo · west madeira coastline", labelSize: 7)
                .frame(height: metrics.points(200))
                .photoScrim(LinearGradient(colors: [.clear, Palette.navy.opacity(0.5)], startPoint: UnitPoint(x: 0.5, y: 0.52), endPoint: .bottom))
                .overlay(alignment: .bottomLeading) {
                    VStack(alignment: .leading, spacing: metrics.points(2)) {
                        Text(model.hero.name).typeStyle(TypeStyle(size: 20, weight: .bold, tracking: -0.3)).foregroundStyle(Palette.cloud)
                        Text(model.hero.subtitle).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.cloudAlpha(0.85))
                    }
                    .padding(EdgeInsets(top: 0, leading: metrics.points(14), bottom: metrics.points(13), trailing: metrics.points(14)))
                }
            VStack(alignment: .leading, spacing: metrics.points(13)) {
                HStack(spacing: metrics.points(8)) {
                    FactTile(key: copy.duration, value: model.hero.durationLabel)
                    FactTile(key: copy.from, value: model.hero.fromPriceLabel)
                    FactTile(key: copy.rating, value: "\(model.hero.rating.formatted(.number.precision(.fractionLength(1)))) ★")
                }
                InsightCard(title: model.hero.fitTitle) {
                    Text(model.hero.fitBody).typeStyle(TypeStyle(size: 13, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.72))
                }
                HStack(spacing: metrics.points(10)) {
                    CallToActionButton(copy.viewExperience) {}
                    Button(action: model.toggleHeroAdded) {
                        Text(model.isHeroAdded ? copy.added : copy.addToTrip).typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(model.isHeroAdded ? Palette.matchTeal : Palette.label)
                            .padding(.horizontal, metrics.points(18)).frame(height: metrics.points(48))
                            .background(model.isHeroAdded ? Palette.aqua.opacity(0.22) : Palette.labelAlpha(0.06), in: Capsule())
                            .overlay(Capsule().strokeBorder(Palette.labelAlpha(0.1), lineWidth: 0.5))
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.97))
                }
            }
            .padding(EdgeInsets(top: metrics.points(14), leading: metrics.points(16), bottom: metrics.points(16), trailing: metrics.points(16)))
        }
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.hero), style: .continuous))
        .surfaceCard(radius: Radius.hero, elevated: true)
    }
}

extension DestinationMedia {
    var descriptor: MediaDescriptor { MediaDescriptor(url: url, placeholder: placeholder, altText: altText, attribution: attribution) }
}
