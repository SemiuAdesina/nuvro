import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

public struct TripsLibraryView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(TransientFeedback.self) private var feedback: TransientFeedback?
    @Bindable private var model: TripsLibraryModel
    private let onContinueTrip: () -> Void
    private let onOpenPlan: () -> Void
    private let onOpenPast: () -> Void
    private let onPlanTrip: () -> Void

    public init(model: TripsLibraryModel, onContinueTrip: @escaping () -> Void, onOpenPlan: @escaping () -> Void, onOpenPast: @escaping () -> Void, onPlanTrip: @escaping () -> Void) {
        self.model = model
        self.onContinueTrip = onContinueTrip
        self.onOpenPlan = onOpenPlan
        self.onOpenPast = onOpenPast
        self.onPlanTrip = onPlanTrip
    }

    private var copy: LibraryCopy { model.copy }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(20)) {
                if model.isEmpty {
                    NuvroEmptyState(symbol: "suitcase", title: copy.emptyTitle, message: copy.emptyBody, action: (copy.findVacation, onPlanTrip))
                        .padding(.top, metrics.points(40))
                }
                if let active = model.entries(in: .active).first {
                    section(copy.happeningNow, entry: active, colour: Palette.matchTeal, dot: true) { ActiveTripCard(entry: active, copy: copy, onContinue: onContinueTrip) }
                }
                if let upcoming = model.entries(in: .upcoming).first {
                    section(copy.comingUp, entry: upcoming) { UpcomingTripCard(entry: upcoming, copy: copy, onOpen: onOpenPlan) }
                }
                if let planning = model.entries(in: .planning).first {
                    section(copy.planning, entry: planning) { PlanningTripRow(entry: planning, copy: copy, onOpen: onOpenPlan) }
                }
                let past = model.entries(in: .past)
                if !past.isEmpty {
                    section(copy.past, entry: nil) {
                        LazyVGrid(columns: [GridItem(.flexible(), spacing: metrics.points(10)), GridItem(.flexible())], spacing: metrics.points(10)) {
                            ForEach(past) { entry in PastTripTile(entry: entry, onOpen: onOpenPast) }
                        }
                    }
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(40))
            .animation(Motion.settle(0.4), value: model.entries)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .toolbar { ToolbarItem(placement: .primaryAction) { GlassPillButton(copy.planTrip, systemImage: "plus", action: onPlanTrip) } }
        .screenTitle(copy.title)
        .confirmationDialog(model.overflow?.destinationName ?? "", isPresented: Binding(get: { model.overflow != nil }, set: { if !$0 { model.overflow = nil } }), titleVisibility: .visible) {
            if let entry = model.overflow {
                if !model.isDownloaded(entry) { Button(copy.downloadOffline) { model.download(entry) } }
                Button(copy.archive) { model.archive(entry) }
                Button(copy.delete, role: .destructive) { model.requestDelete(entry) }
                Button(copy.cancel, role: .cancel) { model.overflow = nil }
            }
        }
        .alert(copy.deleteTitle, isPresented: Binding(get: { model.pendingDelete != nil }, set: { if !$0 { model.cancelDelete() } })) {
            Button(copy.delete, role: .destructive) { model.confirmDelete() }
            Button(copy.cancel, role: .cancel) { model.cancelDelete() }
        } message: {
            Text(copy.deleteBody)
        }
        .onChange(of: model.lastEvent) { _, event in if let event { feedback?.show(event); model.clearEvent() } }
    }

    private func section<Content: View>(_ title: String, entry: TripLibraryEntry?, colour: Color = Palette.labelAlpha(0.5), dot: Bool = false, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: metrics.points(11)) {
            HStack(spacing: metrics.points(8)) {
                if dot { Circle().fill(Palette.aqua).frame(width: metrics.points(7), height: metrics.points(7)).shadow(color: Palette.aqua, radius: 4) }
                EyebrowLabel(title, colour: colour)
                Spacer()
                if let entry {
                    Button { model.overflow = entry } label: {
                        Image(systemName: "ellipsis").font(.system(size: metrics.points(15), weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.45))
                            .frame(width: metrics.points(32), height: metrics.points(28)).contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel("\(copy.more), \(entry.destinationName)")
                }
            }
            content()
                .contextMenu {
                    if let entry {
                        if !model.isDownloaded(entry) { Button(copy.downloadOffline, systemImage: "arrow.down.circle") { model.download(entry) } }
                        Button(copy.archive, systemImage: "archivebox") { model.archive(entry) }
                        Button(copy.delete, systemImage: "trash", role: .destructive) { model.requestDelete(entry) }
                    }
                }
        }
    }
}

struct ActiveTripCard: View {
    @Environment(\.designMetrics) private var metrics
    let entry: TripLibraryEntry
    let copy: LibraryCopy
    let onContinue: () -> Void

    var body: some View {
        VStack(spacing: 0) {
            DynamicDestinationImage(media: entry.media.descriptor, label: "photo · \(entry.destinationName.lowercased()), real destination photo", labelSize: 7)
                .frame(height: metrics.points(190))
                .photoScrim(LinearGradient(colors: [.clear, Palette.navy.opacity(0.6)], startPoint: UnitPoint(x: 0.5, y: 0.42), endPoint: .bottom))
                .overlay(alignment: .bottomLeading) {
                    VStack(alignment: .leading, spacing: metrics.points(3)) {
                        Text(entry.destinationName).typeStyle(TypeStyle(size: 23, weight: .bold, tracking: -0.4)).foregroundStyle(Palette.cloud)
                        HStack(spacing: metrics.points(4)) {
                            Text(entry.statusLabel ?? "").typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.cloudAlpha(0.88))
                            if let weather = entry.weatherLabel {
                                Text("·").foregroundStyle(Palette.cloudAlpha(0.88))
                                Image(systemName: "sun.max.fill").font(.system(size: metrics.points(11))).foregroundStyle(Palette.sunGlow)
                                Text(weather).typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.cloudAlpha(0.88))
                            }
                        }
                    }
                    .padding(EdgeInsets(top: 0, leading: metrics.points(16), bottom: metrics.points(14), trailing: metrics.points(16)))
                }
                .overlay(alignment: .topTrailing) {
                    Text(copy.active).typeStyle(TypeStyle(size: 12, weight: .bold)).foregroundStyle(Palette.cloud)
                        .padding(.horizontal, metrics.points(12)).frame(height: metrics.points(30)).glassCapsule(.control).padding(metrics.points(14))
                }
            VStack(alignment: .leading, spacing: metrics.points(12)) {
                GradientProgressBar(fraction: entry.progress)
                HStack(spacing: metrics.points(11)) {
                    PhotoStandIn(palette: .sky, label: "", labelSize: 6).frame(width: metrics.points(38), height: metrics.points(38)).clipShape(RoundedRectangle(cornerRadius: metrics.points(12), style: .continuous))
                    VStack(alignment: .leading, spacing: metrics.points(1)) {
                        EyebrowLabel(copy.next)
                        Text(entry.nextLabel ?? "").typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(Palette.label)
                    }
                }
                CallToActionButton(copy.continueTrip, action: onContinue)
            }
            .padding(EdgeInsets(top: metrics.points(13), leading: metrics.points(16), bottom: metrics.points(16), trailing: metrics.points(16)))
        }
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.hero), style: .continuous))
        .surfaceCard(radius: Radius.hero, elevated: true)
    }
}

struct UpcomingTripCard: View {
    @Environment(\.designMetrics) private var metrics
    let entry: TripLibraryEntry
    let copy: LibraryCopy
    let onOpen: () -> Void

    var body: some View {
        HStack(spacing: 0) {
            DynamicDestinationImage(media: entry.media.descriptor, labelSize: 6).frame(width: metrics.points(118)).frame(maxHeight: .infinity)
            VStack(alignment: .leading, spacing: metrics.points(7)) {
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    Text(entry.destinationName).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                    Text(entry.dateLabel).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                }
                HStack(spacing: metrics.points(7)) {
                    GradientProgressBar(fraction: entry.progress)
                    Text(entry.readinessLabel ?? "").typeStyle(TypeStyle(size: 11.5, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.6))
                }
                LinkButton(copy.continuePlanning, action: onOpen)
            }
            .padding(EdgeInsets(top: metrics.points(13), leading: metrics.points(14), bottom: metrics.points(13), trailing: metrics.points(14)))
        }
        .fixedSize(horizontal: false, vertical: true)
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(22), style: .continuous))
        .surfaceCard(radius: 22)
    }
}

struct PlanningTripRow: View {
    @Environment(\.designMetrics) private var metrics
    let entry: TripLibraryEntry
    let copy: LibraryCopy
    let onOpen: () -> Void

    var body: some View {
        Button(action: onOpen) {
            HStack(spacing: metrics.points(12)) {
                DynamicDestinationImage(media: entry.media.descriptor, labelSize: 6).frame(width: metrics.points(70), height: metrics.points(62)).clipShape(RoundedRectangle(cornerRadius: metrics.points(14), style: .continuous))
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    Text(entry.destinationName).typeStyle(TypeStyle(size: 16, weight: .semibold)).foregroundStyle(Palette.label)
                    Text(entry.dateLabel).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                }
                Spacer(minLength: 0)
                TagChip(copy.planningTag, tone: .gold)
            }
            .padding(metrics.points(10))
            .surfaceCard(radius: Radius.cardCompact)
        }
        .buttonStyle(PressScaleButtonStyle())
    }
}

struct PastTripTile: View {
    @Environment(\.designMetrics) private var metrics
    let entry: TripLibraryEntry
    let onOpen: () -> Void

    var body: some View {
        Button(action: onOpen) {
            VStack(alignment: .leading, spacing: 0) {
                DynamicDestinationImage(media: entry.media.descriptor, labelSize: 6).frame(height: metrics.points(82))
                VStack(alignment: .leading, spacing: metrics.points(1)) {
                    Text(entry.destinationName).typeStyle(TypeStyle(size: 14, weight: .semibold)).foregroundStyle(Palette.label)
                    Text(entry.dateLabel).typeStyle(TypeStyle(size: 11.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.58))
                }
                .padding(EdgeInsets(top: metrics.points(9), leading: metrics.points(11), bottom: metrics.points(11), trailing: metrics.points(11)))
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            .clipShape(RoundedRectangle(cornerRadius: metrics.points(18), style: .continuous))
            .surfaceCard(radius: 18)
        }
        .buttonStyle(PressScaleButtonStyle(scale: 0.985))
    }
}
