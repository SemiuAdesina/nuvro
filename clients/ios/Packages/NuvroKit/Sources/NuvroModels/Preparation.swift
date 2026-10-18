import Foundation

public struct ChecklistItem: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let note: String?
    public let isDone: Bool

    public init(name: String, note: String? = nil, isDone: Bool = false) {
        self.name = name
        self.note = note
        self.isDone = isDone
    }

    public var id: String { name }

    public func toggled() -> ChecklistItem {
        ChecklistItem(name: name, note: note, isDone: !isDone)
    }

    private enum CodingKeys: String, CodingKey {
        case name, note
        case isDone = "is_done"
    }
}

public struct ChecklistGroup: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let items: [ChecklistItem]

    public init(name: String, items: [ChecklistItem]) {
        self.name = name
        self.items = items
    }

    public var id: String { name }
    public var doneCount: Int { items.filter(\.isDone).count }
}

public struct GroupMember: Sendable, Hashable, Codable, Identifiable {
    public enum Role: String, Sendable, Codable {
        case owner, editor, viewer
    }

    public let name: String
    public let subtitle: String
    public let role: Role
    public let initial: String
    public let placeholder: String

    public init(name: String, subtitle: String, role: Role, initial: String, placeholder: String) {
        self.name = name
        self.subtitle = subtitle
        self.role = role
        self.initial = initial
        self.placeholder = placeholder
    }

    public var id: String { name }
}

public struct DecisionOption: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let votes: Int
    public let hasMyVote: Bool

    public init(name: String, votes: Int, hasMyVote: Bool) {
        self.name = name
        self.votes = votes
        self.hasMyVote = hasMyVote
    }

    public var id: String { name }

    public func toggled() -> DecisionOption {
        DecisionOption(name: name, votes: max(0, votes + (hasMyVote ? -1 : 1)), hasMyVote: !hasMyVote)
    }

    private enum CodingKeys: String, CodingKey {
        case name, votes
        case hasMyVote = "has_my_vote"
    }
}

public struct GroupDecision: Sendable, Hashable, Codable, Identifiable {
    public let question: String
    public let meta: String
    public let options: [DecisionOption]

    public init(question: String, meta: String, options: [DecisionOption]) {
        self.question = question
        self.meta = meta
        self.options = options
    }

    public var id: String { question }
    public var leadingVotes: Int { max(options.map(\.votes).max() ?? 0, 1) }
}

public struct GroupChange: Sendable, Hashable, Codable, Identifiable {
    public enum Tone: String, Sendable, Codable {
        case aqua, sky, neutral
    }

    public let text: String
    public let when: String
    public let tone: Tone

    public init(text: String, when: String, tone: Tone) {
        self.text = text
        self.when = when
        self.tone = tone
    }

    public var id: String { text }
}

public struct EntryRequirementRow: Sendable, Hashable, Codable, Identifiable {
    public enum State: String, Sendable, Codable {
        case met, info
    }

    public let key: String
    public let value: String
    public let state: State

    public init(key: String, value: String, state: State) {
        self.key = key
        self.value = value
        self.state = state
    }

    public var id: String { key }
}

public struct InfoRow: Sendable, Hashable, Codable, Identifiable {
    public let key: String
    public let value: String

    public init(key: String, value: String) {
        self.key = key
        self.value = value
    }

    public var id: String { key }
}

public struct InfoSection: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let gist: String
    public let rows: [InfoRow]
    public let source: String
    public let placeholder: String

    public init(name: String, gist: String, rows: [InfoRow], source: String, placeholder: String) {
        self.name = name
        self.gist = gist
        self.rows = rows
        self.source = source
        self.placeholder = placeholder
    }

    public var id: String { name }
}

public struct EmergencyContact: Sendable, Hashable, Codable, Identifiable {
    public let key: String
    public let value: String

    public init(key: String, value: String) {
        self.key = key
        self.value = value
    }

    public var id: String { key }
}
