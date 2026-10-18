import Foundation

public struct TripAlert: Sendable, Hashable, Codable, Identifiable {
    public enum Priority: String, Sendable, Codable, CaseIterable {
        case now, today, later
    }

    public enum Kind: String, Sendable, Codable {
        case weather, flight, activity, document, summary
    }

    public enum Severity: String, Sendable, Codable {
        case actionNeeded = "action_needed"
        case onTime = "on_time"
        case reminder
        case updated
    }

    public enum Destination: String, Sendable, Codable {
        case replanning, travelMode, document, summary
    }

    public let id: String
    public let priority: Priority
    public let kind: Kind
    public let severity: Severity
    public let title: String
    public let body: String
    public let reason: String
    public let actionLabel: String
    public let destination: Destination

    public init(id: String, priority: Priority, kind: Kind, severity: Severity, title: String, body: String, reason: String, actionLabel: String, destination: Destination) {
        self.id = id
        self.priority = priority
        self.kind = kind
        self.severity = severity
        self.title = title
        self.body = body
        self.reason = reason
        self.actionLabel = actionLabel
        self.destination = destination
    }

    private enum CodingKeys: String, CodingKey {
        case id, priority, kind, severity, title, body, reason, destination
        case actionLabel = "action_label"
    }
}
