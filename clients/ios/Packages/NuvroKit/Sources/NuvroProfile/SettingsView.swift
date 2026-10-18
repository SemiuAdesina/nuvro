import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import NuvroPermissions
import Observation
import SwiftUI

public struct AccountActions: Sendable {
    public let signOut: @MainActor () async -> Void
    public let deleteAccount: @MainActor () async throws -> Void

    public init(signOut: @escaping @MainActor () async -> Void, deleteAccount: @escaping @MainActor () async throws -> Void) {
        self.signOut = signOut
        self.deleteAccount = deleteAccount
    }
}

public enum SettingsPresentation: Equatable, Sendable {
    case signOutConfirmation
    case deleteExplainer
    case deleteConfirmation
}

@Observable
@MainActor
public final class SettingsModel {
    public let groups: [SettingsGroup]
    public private(set) var notifications: [NotificationGroup]
    public var permissions: [PermissionStatus]
    public var presentation: SettingsPresentation?
    public private(set) var isDeleting = false
    public private(set) var deleteFailed = false
    public let copy: SettingsCopy
    public var actions: AccountActions?

    public init(groups: [SettingsGroup], notifications: [NotificationGroup], permissions: [PermissionStatus], actions: AccountActions? = nil, language: Language = .english) {
        self.groups = groups
        self.notifications = notifications
        self.permissions = permissions
        self.actions = actions
        self.copy = .resolve(language)
    }

    public var canManageAccount: Bool { actions != nil }

    public func select(_ row: SettingsRow) {
        guard row.name == copy.deleteAccount, canManageAccount else { return }
        deleteFailed = false
        presentation = .deleteExplainer
    }

    public func requestSignOut() {
        guard canManageAccount else { return }
        presentation = .signOutConfirmation
    }

    public func confirmSignOut() async {
        presentation = nil
        await actions?.signOut()
    }

    public func proceedToDeleteConfirmation() {
        presentation = .deleteConfirmation
    }

    public func confirmDelete() async {
        guard let actions, !isDeleting else { return }
        presentation = .deleteExplainer
        isDeleting = true
        deleteFailed = false
        do {
            try await actions.deleteAccount()
            presentation = nil
        } catch {
            deleteFailed = true
        }
        isDeleting = false
    }

    public func dismissPresentation() {
        presentation = nil
    }

    public func toggle(_ setting: NotificationSetting, in group: NotificationGroup) {
        notifications = notifications.map { current in
            guard current.id == group.id else { return current }
            return NotificationGroup(name: current.name, settings: current.settings.map { $0.id == setting.id ? NotificationSetting(name: $0.name, isOn: !$0.isOn) : $0 })
        }
    }

    public func isOn(_ setting: NotificationSetting) -> Bool {
        notifications.flatMap(\.settings).first { $0.id == setting.id }?.isOn ?? setting.isOn
    }
}

public struct SettingsView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.openURL) private var openURL
    @Environment(PermissionCenter.self) private var permissions: PermissionCenter?
    @AppStorage(ThemePreference.storageKey) private var theme: ThemePreference = .system
    @Bindable private var model: SettingsModel

    public init(model: SettingsModel) {
        self.model = model
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(18)) {
                ForEach(model.groups) { group in
                    VStack(alignment: .leading, spacing: metrics.points(9)) {
                        EyebrowLabel(group.name.uppercased())
                        VStack(spacing: 0) {
                            ForEach(group.rows) { row in
                                Button { model.select(row) } label: {
                                    HStack {
                                        Text(row.name).typeStyle(TypeStyle(size: 15.5, weight: .medium)).foregroundStyle(row.isDestructive ? Palette.destructive : Palette.label)
                                        Spacer()
                                        Text(row.value).typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.55))
                                        ChevronGlyph()
                                    }
                                    .padding(.horizontal, metrics.points(15)).padding(.vertical, metrics.points(13))
                                    .contentShape(Rectangle())
                                }
                                .buttonStyle(.plain)
                                if row.id != group.rows.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                            }
                        }
                        .surfaceCard()
                    }
                }
                VStack(alignment: .leading, spacing: metrics.points(9)) {
                    EyebrowLabel(model.copy.appearance)
                    VStack(alignment: .leading, spacing: metrics.points(10)) {
                        SegmentedChoice(options: ThemePreference.allCases, selection: $theme, isPill: true) { appearanceLabel($0) }
                        Text(model.copy.appearanceNote).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.55))
                    }
                    .padding(metrics.points(14))
                    .surfaceCard()
                }
                if model.canManageAccount {
                    Button(action: model.requestSignOut) {
                        Text(model.copy.signOut).typeStyle(TypeStyle(size: 15.5, weight: .semibold)).foregroundStyle(Palette.sky)
                            .frame(maxWidth: .infinity).padding(.vertical, metrics.points(13))
                    }
                    .buttonStyle(.plain)
                    .surfaceCard()
                }
                VStack(alignment: .leading, spacing: metrics.points(9)) {
                    EyebrowLabel(model.copy.notifications)
                    ForEach(model.notifications) { group in
                        VStack(alignment: .leading, spacing: 0) {
                            Text(group.name).typeStyle(TypeStyle(size: 14, weight: .bold)).foregroundStyle(Palette.label).padding(EdgeInsets(top: metrics.points(12), leading: metrics.points(15), bottom: metrics.points(6), trailing: metrics.points(15)))
                            ForEach(group.settings) { setting in
                                Toggle(isOn: Binding(get: { model.isOn(setting) }, set: { _ in model.toggle(setting, in: group) })) {
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
                VStack(alignment: .leading, spacing: metrics.points(9)) {
                    EyebrowLabel(model.copy.permissions)
                    VStack(spacing: 0) {
                        ForEach(model.permissions) { permission in
                            HStack {
                                Text(permission.name).typeStyle(TypeStyle(size: 15, weight: .medium)).foregroundStyle(Palette.label)
                                Spacer()
                                TagChip(permission.value, tone: permission.tone == .granted ? .aqua : permission.tone == .ask ? .gold : .twilight)
                            }
                            .padding(.horizontal, metrics.points(15)).padding(.vertical, metrics.points(12))
                            Divider().overlay(Palette.labelAlpha(0.07))
                        }
                        Button { if let url = SystemSettings.url { openURL(url) } } label: {
                            HStack {
                                Text(model.copy.openSettings).typeStyle(TypeStyle(size: 14, weight: .semibold)).foregroundStyle(Palette.sky)
                                Spacer()
                                Image(systemName: "arrow.up.right").font(.system(size: metrics.points(12), weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.3))
                            }
                            .padding(.horizontal, metrics.points(15)).padding(.vertical, metrics.points(12))
                        }
                        .buttonStyle(.plain)
                    }
                    .surfaceCard()
                }
                HStack(spacing: metrics.points(16)) {
                    ForEach(model.copy.legal, id: \.self) { item in
                        Button(item) {}.typeStyle(TypeStyle(size: 12.5, weight: .medium)).foregroundStyle(Palette.labelAlpha(0.5)).buttonStyle(.plain)
                    }
                }
                .frame(maxWidth: .infinity)
                Text(model.copy.version).typeStyle(TypeStyle(size: 11, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.4)).frame(maxWidth: .infinity)
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(model.copy.title)
        .task {
            guard let permissions else { return }
            await permissions.refreshAll()
            model.permissions = model.permissions.map { status in
                guard let kind = PermissionKind(rawValue: status.name.lowercased()) else { return status }
                let live = permissions.state(of: kind)
                let value: String = switch live {
                case .granted: model.copy.granted
                case .denied: model.copy.denied
                case .restricted: model.copy.restricted
                case .notDetermined: model.copy.notRequested
                }
                return PermissionStatus(name: status.name, value: value, tone: live == .granted ? .granted : live == .notDetermined ? .ask : .none)
            }
        }
        .confirmationDialog(model.copy.signOutTitle, isPresented: presenting(.signOutConfirmation), titleVisibility: .visible) {
            Button(model.copy.signOut, role: .destructive) { Task { await model.confirmSignOut() } }
            Button(model.copy.cancel, role: .cancel) { model.dismissPresentation() }
        } message: {
            Text(model.copy.signOutBody)
        }
        .sheet(isPresented: presenting(.deleteExplainer, .deleteConfirmation)) {
            AccountDeletionSheet(model: model)
        }
    }

    private func appearanceLabel(_ preference: ThemePreference) -> String {
        switch preference {
        case .system: model.copy.appearanceSystem
        case .light: model.copy.appearanceLight
        case .dark: model.copy.appearanceDark
        }
    }

    private func presenting(_ states: SettingsPresentation...) -> Binding<Bool> {
        Binding(
            get: { states.contains { $0 == model.presentation } },
            set: { if !$0 && !model.isDeleting { model.dismissPresentation() } }
        )
    }
}
