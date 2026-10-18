import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import NuvroPermissions
import SwiftUI

public struct TripAlertsView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.openURL) private var openURL
    @Environment(PermissionCenter.self) private var permissions: PermissionCenter?
    @Bindable private var model: TripAlertsModel
    private let permissionCopy = PermissionCopy.resolve(.english)
    private let onOpen: (TripAlert.Destination) -> Void

    public init(model: TripAlertsModel, onOpen: @escaping (TripAlert.Destination) -> Void) {
        self.model = model
        self.onOpen = onOpen
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                notificationsGuidance
                if model.isAllClear {
                    NuvroEmptyState(symbol: "checkmark.circle", title: model.copy.allCaughtUp, message: model.copy.allCaughtUpBody)
                        .padding(.top, metrics.points(24))
                        .transition(.opacity)
                }
                ForEach(model.groups, id: \.priority) { group in
                    VStack(alignment: .leading, spacing: metrics.points(8)) {
                        HStack(spacing: metrics.points(8)) {
                            Text(group.title.uppercased()).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.3)).foregroundStyle(groupColour(group.priority))
                            Rectangle().fill(Palette.labelAlpha(0.1)).frame(height: 0.5)
                            Text(model.descriptor(for: group.priority)).typeStyle(TypeStyle(size: 11.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.5))
                        }
                        ForEach(group.alerts) { alert in
                            TripAlertRow(alert: alert, model: model, onOpen: { model.markRead(alert); onOpen(alert.destination) })
                                .transition(.opacity)
                        }
                    }
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(40))
            .animation(Motion.settle(0.5), value: model.dismissedIDs)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .refreshable { await model.refresh() }
        .screenTitle(model.copy.title, subtitle: model.copy.subtitle)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button { model.isShowingPreferences = true } label: {
                    Text(model.copy.preferences).typeStyle(Typography.control).foregroundStyle(Palette.label).padding(.horizontal, metrics.points(4))
                }
                .accessibilityHint(model.copy.preferencesBody)
            }
        }
        .task { await permissions?.refresh(.notifications) }
        .sheet(isPresented: $model.isShowingPreferences) { AlertPreferencesSheet(model: model) }
    }

    @ViewBuilder
    private var notificationsGuidance: some View {
        if let permissions {
            switch permissions.state(of: .notifications) {
            case .notDetermined:
                NuvroPermissionExplainer(
                    symbol: "bell.badge.fill", title: permissionCopy.notificationsTitle, message: permissionCopy.notificationsBody,
                    primary: (permissionCopy.allow, { Task { if await permissions.prepare(.notifications) == false { _ = await permissions.proceedFromExplainer() } } })
                )
            case .denied, .restricted:
                NuvroPermissionExplainer(
                    symbol: "bell.slash.fill", palette: .gold, title: permissionCopy.notificationsDeniedTitle, message: permissionCopy.notificationsDeniedBody,
                    primary: (permissionCopy.openSettings, { if let url = SystemSettings.url { openURL(url) } })
                )
            case .granted:
                EmptyView()
            }
        }
    }

    private func groupColour(_ priority: TripAlert.Priority) -> Color {
        switch priority {
        case .now: Palette.amber
        case .today: Palette.matchTeal
        case .later: Palette.labelAlpha(0.5)
        }
    }
}

struct TripAlertRow: View {
    @Environment(\.designMetrics) private var metrics
    let alert: TripAlert
    let model: TripAlertsModel
    let onOpen: () -> Void

    private var tint: (fill: Color, text: Color, border: Color) {
        switch alert.severity {
        case .actionNeeded: (Palette.gold.opacity(0.28), Palette.amber, Palette.gold.opacity(0.55))
        case .onTime: (Palette.sky.opacity(0.14), Palette.sky, Palette.sky.opacity(0.35))
        case .updated: (Palette.twilight.opacity(0.12), Palette.twilight, Palette.twilight.opacity(0.28))
        case .reminder: (Palette.aqua.opacity(0.18), Palette.matchTeal, Palette.labelAlpha(0.1))
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(10)) {
            HStack(alignment: .top, spacing: metrics.points(12)) {
                Image(systemName: model.symbol(for: alert.kind))
                    .font(.system(size: metrics.points(14), weight: .semibold))
                    .foregroundStyle(tint.text)
                    .frame(width: metrics.points(32), height: metrics.points(32))
                    .background(tint.fill, in: RoundedRectangle(cornerRadius: metrics.points(10), style: .continuous))
                VStack(alignment: .leading, spacing: metrics.points(3)) {
                    HStack(spacing: metrics.points(7)) {
                        if model.isUnread(alert) {
                            Circle().fill(Palette.sky).frame(width: metrics.points(7), height: metrics.points(7)).accessibilityLabel("Unread")
                        }
                        Text(alert.title).typeStyle(TypeStyle(size: 15.5, weight: .bold, tracking: -0.1)).foregroundStyle(Palette.label)
                        Text(model.severityLabel(alert.severity)).typeStyle(TypeStyle(size: 10, weight: .bold, tracking: 0.6)).foregroundStyle(tint.text)
                            .padding(.horizontal, metrics.points(7)).padding(.vertical, metrics.points(2.5))
                            .background(tint.fill, in: Capsule())
                    }
                    Text(alert.body).typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.72))
                    Text(alert.reason).typeStyle(TypeStyle(size: 11.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.5))
                }
            }
            HStack(spacing: metrics.points(8)) {
                SurfaceLinkButton(alert.actionLabel, action: onOpen)
                Button(model.copy.dismiss) { model.dismiss(alert) }
                    .typeStyle(TypeStyle(size: 14, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.5))
                    .buttonStyle(.plain)
                    .padding(.horizontal, metrics.points(14))
            }
            .padding(.leading, metrics.points(44))
        }
        .padding(EdgeInsets(top: metrics.points(14), leading: metrics.points(15), bottom: metrics.points(14), trailing: metrics.points(15)))
        .background(Palette.surface, in: RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous).strokeBorder(tint.border, lineWidth: 0.5))
        .accessibilityElement(children: .contain)
    }
}

struct AlertPreferencesSheet: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable var model: TripAlertsModel

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: metrics.points(14)) {
                    ForEach(model.preferences) { group in
                        VStack(alignment: .leading, spacing: 0) {
                            Text(group.name).typeStyle(TypeStyle(size: 14, weight: .bold)).foregroundStyle(Palette.label).padding(EdgeInsets(top: metrics.points(12), leading: metrics.points(15), bottom: metrics.points(6), trailing: metrics.points(15)))
                            ForEach(group.settings) { setting in
                                Toggle(isOn: Binding(get: { model.isOn(setting) }, set: { _ in model.toggle(setting) })) {
                                    Text(setting.name).typeStyle(TypeStyle(size: 15, weight: .medium)).foregroundStyle(Palette.label)
                                }
                                .tint(Palette.sky)
                                .padding(.horizontal, metrics.points(15)).padding(.vertical, metrics.points(10))
                                if setting.id != group.settings.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                            }
                        }
                        .surfaceCard()
                    }
                }
                .padding(metrics.points(20))
            }
            .background(Palette.background)
            .navigationTitle(model.copy.preferencesTitle)
            .toolbar { ToolbarItem(placement: .confirmationAction) { Button(model.copy.done) { model.isShowingPreferences = false } } }
            .toolbarTitleDisplayMode(.inline)
        }
        .presentationDetents([.medium, .large])
    }
}
