import Foundation
import NuvroLocalization
import NuvroModels
import Observation

@Observable
@MainActor
public final class TripAlertsModel {
    public private(set) var alerts: [TripAlert]
    public private(set) var dismissedIDs: Set<String> = []
    public private(set) var readIDs: Set<String> = []
    public private(set) var preferences: [NotificationGroup]
    public var isShowingPreferences = false
    public private(set) var isRefreshing = false
    public let copy: AlertsCopy
    private let source: (@Sendable () async throws -> [TripAlert])?

    public init(alerts: [TripAlert], preferences: [NotificationGroup] = [], source: (@Sendable () async throws -> [TripAlert])? = nil, language: Language = .english) {
        self.alerts = alerts
        self.preferences = preferences
        self.source = source
        self.copy = .resolve(language)
    }

    public func refresh() async {
        guard let source, !isRefreshing else { return }
        isRefreshing = true
        if let fresh = try? await source() { alerts = fresh }
        isRefreshing = false
    }

    public func isOn(_ setting: NotificationSetting) -> Bool {
        preferences.flatMap(\.settings).first { $0.id == setting.id }?.isOn ?? setting.isOn
    }

    public func toggle(_ setting: NotificationSetting) {
        preferences = preferences.map { group in
            NotificationGroup(name: group.name, settings: group.settings.map { $0.id == setting.id ? NotificationSetting(name: $0.name, isOn: !$0.isOn) : $0 })
        }
    }

    public var groups: [(priority: TripAlert.Priority, title: String, alerts: [TripAlert])] {
        TripAlert.Priority.allCases.compactMap { priority in
            let items = alerts.filter { $0.priority == priority && !dismissedIDs.contains($0.id) }
            return items.isEmpty ? nil : (priority, title(for: priority), items)
        }
    }

    public var isAllClear: Bool { groups.isEmpty }

    public var activeCount: Int { alerts.count - dismissedIDs.count }

    public func dismiss(_ alert: TripAlert) {
        dismissedIDs.insert(alert.id)
    }

    public func markRead(_ alert: TripAlert) {
        readIDs.insert(alert.id)
    }

    public func isUnread(_ alert: TripAlert) -> Bool {
        !readIDs.contains(alert.id) && !dismissedIDs.contains(alert.id)
    }

    public var unreadCount: Int {
        alerts.filter(isUnread).count
    }

    public func title(for priority: TripAlert.Priority) -> String {
        switch priority {
        case .now: copy.now
        case .today: copy.today
        case .later: copy.later
        }
    }

    public func severityLabel(_ severity: TripAlert.Severity) -> String {
        switch severity {
        case .actionNeeded: copy.actionNeeded
        case .onTime: copy.onTime
        case .reminder: copy.reminder
        case .updated: copy.updated
        }
    }

    public func descriptor(for priority: TripAlert.Priority) -> String {
        switch priority {
        case .now: copy.needsDecision
        case .today: copy.goodToKnow
        case .later: copy.noActionYet
        }
    }

    public func symbol(for kind: TripAlert.Kind) -> String {
        switch kind {
        case .weather: "cloud.rain.fill"
        case .flight: "airplane"
        case .activity: "clock.fill"
        case .document: "doc.text.fill"
        case .summary: "sparkles"
        }
    }
}
