import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

public struct TravelModeActions {
    public var fullItinerary: () -> Void
    public var reviewPlan: () -> Void
    public var navigate: () -> Void
    public var tickets: () -> Void
    public var translate: () -> Void
    public var updates: () -> Void

    public init(fullItinerary: @escaping () -> Void = {}, reviewPlan: @escaping () -> Void = {}, navigate: @escaping () -> Void = {}, tickets: @escaping () -> Void = {}, translate: @escaping () -> Void = {}, updates: @escaping () -> Void = {}) {
        self.fullItinerary = fullItinerary
        self.reviewPlan = reviewPlan
        self.navigate = navigate
        self.tickets = tickets
        self.translate = translate
        self.updates = updates
    }
}

public struct TravelModeHomeView: View {
    @Environment(\.designMetrics) private var metrics
    private let model: TravelModeModel
    private let actions: TravelModeActions
    private let updateCount: Int

    public init(model: TravelModeModel, actions: TravelModeActions, updateCount: Int = 0) {
        self.model = model
        self.actions = actions
        self.updateCount = updateCount
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                header
                if let next = model.nextActivity {
                    NextActivityHero(activity: next, leaveIn: model.leaveInMinutes, copy: model.copy, onNavigate: actions.navigate, onTicket: actions.tickets)
                }
                statusRow
                quickActions
                todayCard
                if model.hasPendingWeatherChange {
                    Button(action: actions.reviewPlan) {
                        InsightCard(title: "Light rain after 3 PM", showsMark: false) {
                            Text("Your outdoor plan is still fine, but the garden may be better before 2 PM.").typeStyle(Typography.bodyCompact).foregroundStyle(Palette.labelAlpha(0.75))
                            SurfaceLinkButton(model.copy.reviewPlan, action: actions.reviewPlan)
                        }
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.99))
                    .transition(.opacity)
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(40))
            .animation(Motion.fade(0.4), value: model.hasPendingWeatherChange)
        }
        .scrollIndicators(.hidden)
        .screenTitle("Day \(model.dayNumber) \(model.copy.dayIn) \(model.destinationName)", subtitle: model.dateLabel)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button(action: actions.updates) {
                    Image(systemName: "bell.fill").font(.system(size: metrics.points(15), weight: .semibold)).foregroundStyle(Palette.twilightText)
                        .overlay(alignment: .topTrailing) {
                            if updateCount > 0 {
                                Text("\(updateCount)").font(.system(size: metrics.points(9), weight: .bold)).foregroundStyle(.white)
                                    .frame(width: metrics.points(15), height: metrics.points(15)).background(Palette.amber, in: Circle()).offset(x: 6, y: -6)
                            }
                        }
                }
                .accessibilityLabel("Trip updates, \(updateCount) new")
            }
        }
        .background(alignment: .top) {
            VStack(spacing: 0) {
                LinearGradient(stops: [.init(color: Palette.dawnWashTop, location: 0), .init(color: Palette.dawnWashMid, location: 0.26), .init(color: Palette.background, location: 0.76)], startPoint: .top, endPoint: .bottom)
                    .frame(height: metrics.points(330))
                Palette.background
            }
            .ignoresSafeArea()
        }
    }

    private var header: some View {
        HStack(spacing: metrics.points(6)) {
            Text(model.greeting).typeStyle(Typography.greeting).foregroundStyle(Palette.labelAlpha(0.6))
            Text("·").foregroundStyle(Palette.labelAlpha(0.4))
            Image(systemName: "sun.max.fill").font(.system(size: metrics.points(12))).foregroundStyle(Palette.amber)
            Text("\(model.weather.temperatureCelsius)° \(model.weather.condition.lowercased())").typeStyle(Typography.bodyCompact).foregroundStyle(Palette.labelAlpha(0.6))
                if model.weatherFreshness != .fresh {
                    FreshnessLabel(tone: model.weatherFreshness == .stale ? .stale : .cached, text: StateCopy.resolve(.english).cached)
                }
        }
    }

    private var statusRow: some View {
        HStack(spacing: metrics.points(8)) {
            statusTile(model.copy.weather, "Rain 3 PM", Palette.amber)
            statusTile(model.copy.transport, model.copy.normal, Palette.matchTeal)
            statusTile(model.copy.reservations, "\(model.day.activities.filter { $0.status == .reserved }.count) today", Palette.label)
        }
    }

    private func statusTile(_ key: String, _ value: String, _ colour: Color) -> some View {
        VStack(alignment: .leading, spacing: metrics.points(2)) {
            Text(key).typeStyle(Typography.smallEmphasis).foregroundStyle(Palette.labelAlpha(0.55))
            Text(value).typeStyle(Typography.bodyCompactEmphasis).foregroundStyle(colour)
        }
        .padding(.horizontal, metrics.points(12)).padding(.vertical, metrics.points(10))
        .frame(maxWidth: .infinity, alignment: .leading)
        .surfaceCard(radius: 16)
    }

    private var quickActions: some View {
        HStack(spacing: metrics.points(8)) {
            GlassActionTile(title: model.copy.navigate, systemImage: "location.fill", palette: .sky, action: actions.navigate)
            GlassActionTile(title: model.copy.tickets, systemImage: "ticket.fill", palette: .gold, action: actions.tickets)
            GlassActionTile(title: model.copy.translate, systemImage: "character.bubble.fill", palette: .aqua, action: actions.translate)
            GlassActionTile(title: model.copy.askNuvro, systemImage: "sparkles", palette: .twilight) {}
        }
    }

    private var todayCard: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            HStack(alignment: .firstTextBaseline) {
                Text(model.copy.today).typeStyle(Typography.cardHeading).foregroundStyle(Palette.label)
                Spacer()
                LinkButton(model.copy.fullItinerary, action: actions.fullItinerary)
            }
            VStack(spacing: 0) {
                ForEach(Array(model.day.activities.enumerated()), id: \.element.id) { index, activity in
                    TimelineRow(time: activity.time, isFirst: index == 0, isLast: index == model.day.activities.count - 1, isEmphasised: model.isNext(activity)) {
                        HStack {
                            Text(activity.title).typeStyle(Typography.secondaryAction).foregroundStyle(Palette.label)
                            Spacer(minLength: 0)
                            if let tag = tag(for: activity) {
                                Text(tag.text).typeStyle(TypeStyle(size: 11.5, weight: .semibold)).foregroundStyle(tag.colour)
                            }
                        }
                    }
                    .transition(.opacity)
                }
            }
            .animation(Motion.settle(0.6), value: model.day)
        }
        .padding(metrics.points(16))
        .surfaceCard()
    }

    private func tag(for activity: ItineraryActivity) -> (text: String, colour: Color)? {
        if model.wasUpdated(activity) { return (model.copy.updated, Palette.sky) }
        if model.isNext(activity) { return ("Next", Palette.sky) }
        if activity.status == .reserved { return ("Reserved", Palette.matchTeal) }
        return nil
    }
}

struct NextActivityHero: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var pulsing = false
    let activity: ItineraryActivity
    let leaveIn: Int?
    let copy: TravelCopy
    let onNavigate: () -> Void
    let onTicket: () -> Void

    var body: some View {
        ZStack(alignment: .topTrailing) {
            LinearGradient(colors: [Palette.twilightDeep, Palette.twilight, Palette.navy], startPoint: .topLeading, endPoint: .bottomTrailing)
            Circle().fill(RadialGradient(colors: [Palette.sunGlow.opacity(0.35), .clear], center: .center, startRadius: 0, endRadius: metrics.points(110)))
                .frame(width: metrics.points(220), height: metrics.points(220)).offset(x: metrics.points(40), y: -metrics.points(60))
            VStack(alignment: .leading, spacing: metrics.points(14)) {
                HStack(alignment: .top, spacing: metrics.points(12)) {
                    VStack(alignment: .leading, spacing: metrics.points(4)) {
                        Text(copy.next).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.5)).foregroundStyle(Palette.cloudAlpha(0.7))
                        Text(activity.title).typeStyle(TypeStyle(size: 24, weight: .bold, tracking: -0.4)).foregroundStyle(Palette.cloud)
                        Text("\(activity.time) · \(activity.detail)").typeStyle(Typography.bodyCompact).foregroundStyle(Palette.cloudAlpha(0.8))
                    }
                    Spacer(minLength: 0)
                    if let photo = activity.photoReference {
                        PhotoStandIn(palette: .named(photo), label: "", labelSize: 8)
                            .frame(width: metrics.points(66), height: metrics.points(58))
                            .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.chip), style: .continuous))
                    }
                }
                if let leaveIn {
                    HStack(spacing: metrics.points(8)) {
                        ZStack {
                            Circle().fill(Palette.gold)
                                .scaleEffect(pulsing ? 2.3 : 1)
                                .opacity(pulsing ? 0 : 0.85)
                                .animation(reduceMotion ? nil : .easeOut(duration: 2).repeatForever(autoreverses: false), value: pulsing)
                            Circle().fill(Palette.gold)
                        }
                        .frame(width: metrics.points(8), height: metrics.points(8))
                        .task { pulsing = true }
                        Text("\(copy.leaveIn) \(leaveIn) \(copy.minutes)").typeStyle(TypeStyle(size: 13.5, weight: .bold)).foregroundStyle(Palette.cloud)
                    }
                    .padding(.horizontal, metrics.points(13)).padding(.vertical, metrics.points(8))
                    .glassCapsule(.control)
                }
                HStack(spacing: metrics.points(10)) {
                    Button(action: onNavigate) {
                        Text(copy.navigate).typeStyle(TypeStyle(size: 15.5, weight: .semibold)).foregroundStyle(.white).frame(maxWidth: .infinity).frame(height: metrics.points(46))
                    }
                    .buttonStyle(GlassCallToActionStyle())
                    Button(action: onTicket) {
                        Text(copy.ticket).typeStyle(TypeStyle(size: 15.5, weight: .semibold)).foregroundStyle(Palette.cloud).frame(maxWidth: .infinity).frame(height: metrics.points(46))
                    }
                    .buttonStyle(GlassPillButtonStyle(style: .control))
                }
            }
            .padding(EdgeInsets(top: metrics.points(18), leading: metrics.points(18), bottom: metrics.points(16), trailing: metrics.points(18)))
        }
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.hero), style: .continuous))
        .shadow(color: Palette.labelAlpha(0.06), radius: metrics.points(6), y: metrics.points(2))
        .accessibilityElement(children: .contain)
    }
}
