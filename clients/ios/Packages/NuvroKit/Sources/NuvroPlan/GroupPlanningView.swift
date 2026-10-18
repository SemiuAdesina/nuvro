import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation
import SwiftUI

public enum GroupPresentation: Equatable, Sendable {
    case memberActions(GroupMember)
    case removeConfirmation(GroupMember)
    case leaveConfirmation
    case invite
}

@Observable
@MainActor
public final class GroupModel {
    public private(set) var members: [GroupMember]
    public private(set) var decisions: [GroupDecision]
    public let changes: [GroupChange]
    public let budgetLabel: String
    public let budgetNote: String
    public var presentation: GroupPresentation?
    public var inviteEmail = ""
    public private(set) var hasLeft = false
    public private(set) var lastEvent: String?

    public func clearEvent() {
        lastEvent = nil
    }
    public let copy: GroupCopy

    public init(members: [GroupMember], decisions: [GroupDecision], changes: [GroupChange], budgetLabel: String, budgetNote: String, language: Language = .english) {
        self.members = members
        self.decisions = decisions
        self.changes = changes
        self.budgetLabel = budgetLabel
        self.budgetNote = budgetNote
        self.copy = .resolve(language)
    }

    public var canInvite: Bool {
        let trimmed = inviteEmail.trimmingCharacters(in: .whitespaces)
        return trimmed.contains("@") && trimmed.contains(".") && !members.contains { $0.subtitle.localizedCaseInsensitiveContains(trimmed) }
    }

    public func canRemove(_ member: GroupMember) -> Bool {
        member.role != .owner
    }

    public func setRole(_ role: GroupMember.Role, for member: GroupMember) {
        guard member.role != .owner else { return }
        members = members.map { $0.id == member.id ? GroupMember(name: $0.name, subtitle: $0.subtitle, role: role, initial: $0.initial, placeholder: $0.placeholder) : $0 }
        presentation = nil
    }

    public func requestRemove(_ member: GroupMember) {
        guard canRemove(member) else { return }
        presentation = .removeConfirmation(member)
    }

    public func confirmRemove() {
        guard case .removeConfirmation(let member) = presentation else { return }
        members.removeAll { $0.id == member.id }
        presentation = nil
        lastEvent = copy.removed
    }

    public func requestLeave() {
        presentation = .leaveConfirmation
    }

    public func confirmLeave() {
        hasLeft = true
        presentation = nil
    }

    public func sendInvite() {
        guard canInvite else { return }
        let email = inviteEmail.trimmingCharacters(in: .whitespaces)
        members.append(GroupMember(name: email, subtitle: copy.invitePending, role: .viewer, initial: String(email.prefix(1)).uppercased(), placeholder: "sky"))
        inviteEmail = ""
        presentation = nil
        lastEvent = copy.inviteSent
    }

    public func dismissPresentation() {
        presentation = nil
    }

    public func vote(_ option: DecisionOption, in decision: GroupDecision) {
        decisions = decisions.map { current in
            guard current.id == decision.id else { return current }
            return GroupDecision(question: current.question, meta: current.meta, options: current.options.map { $0.id == option.id ? $0.toggled() : $0 })
        }
    }

    public func roleLabel(_ role: GroupMember.Role) -> String {
        switch role {
        case .owner: copy.owner
        case .editor: copy.editor
        case .viewer: copy.viewer
        }
    }

    public func votesLabel(_ option: DecisionOption) -> String { option.votes == 1 ? "1 \(copy.vote)" : "\(option.votes) \(copy.votes)" }
}

public struct GroupPlanningView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(TransientFeedback.self) private var feedback: TransientFeedback?
    @Bindable private var model: GroupModel
    private let subtitle: String

    public init(model: GroupModel, subtitle: String) {
        self.model = model
        self.subtitle = subtitle
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                members
                VStack(alignment: .leading, spacing: metrics.points(11)) {
                    Text(model.copy.decideTogether).typeStyle(TypeStyle(size: 18, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                    ForEach(model.decisions) { decision in DecisionCard(decision: decision, model: model) }
                }
                HStack(alignment: .bottom) {
                    VStack(alignment: .leading, spacing: metrics.points(3)) {
                        Text(model.copy.groupBudget).typeStyle(TypeStyle(size: 12, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.55))
                        Text(model.budgetLabel).typeStyle(TypeStyle(size: 28, weight: .bold, tracking: -0.6)).foregroundStyle(Palette.label)
                        Text(model.budgetNote).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                    }
                    Spacer()
                    LinkButton(model.copy.view) {}
                }
                .padding(metrics.points(16))
                .surfaceCard()
                VStack(alignment: .leading, spacing: metrics.points(11)) {
                    Text(model.copy.recentChanges).typeStyle(TypeStyle(size: 18, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                    VStack(spacing: 0) {
                        ForEach(model.changes) { change in
                            HStack(spacing: metrics.points(11)) {
                                Circle().fill(changeColour(change.tone)).frame(width: metrics.points(8), height: metrics.points(8))
                                Text(change.text).typeStyle(TypeStyle(size: 14, weight: .regular)).foregroundStyle(Palette.label)
                                Spacer()
                                Text(change.when).typeStyle(TypeStyle(size: 11.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.5))
                            }
                            .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(11))
                            if change.id != model.changes.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                        }
                    }
                    .surfaceCard()
                }
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
            .animation(Motion.settle(0.4), value: model.decisions)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(model.copy.title, subtitle: subtitle)
        .toolbar { ToolbarItem(placement: .primaryAction) { GlassPillButton(model.copy.invite, systemImage: "plus") { model.presentation = .invite } } }
        .groupPresentations(model: model)
        .onChange(of: model.lastEvent) { _, event in if let event { feedback?.show(event); model.clearEvent() } }
    }

    private func changeColour(_ tone: GroupChange.Tone) -> Color {
        switch tone {
        case .aqua: Palette.aqua
        case .sky: Palette.sky
        case .neutral: Palette.labelAlpha(0.25)
        }
    }

    private var members: some View {
        VStack(spacing: 0) {
            ForEach(model.members) { member in
                HStack(spacing: metrics.points(12)) {
                    Text(member.initial).typeStyle(TypeStyle(size: 14, weight: .bold)).foregroundStyle(Palette.cloud)
                        .frame(width: metrics.points(38), height: metrics.points(38))
                        .background(PhotoStandIn(palette: .named(member.placeholder), label: "", labelSize: 6)).clipShape(Circle())
                    VStack(alignment: .leading, spacing: 1) {
                        Text(member.name).typeStyle(TypeStyle(size: 15.5, weight: .semibold)).foregroundStyle(Palette.label)
                        Text(member.subtitle).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                    }
                    Spacer()
                    TagChip(model.roleLabel(member.role), tone: member.role == .owner ? .sky : member.role == .editor ? .aqua : .twilight)
                    if model.canRemove(member) {
                        Button { model.presentation = .memberActions(member) } label: {
                            Image(systemName: "ellipsis").font(.system(size: metrics.points(14), weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.4))
                                .frame(width: metrics.points(28), height: metrics.points(28)).contentShape(Rectangle())
                        }
                        .buttonStyle(.plain)
                        .accessibilityLabel("\(model.copy.memberActions), \(member.name)")
                    }
                }
                .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(12))
                Divider().overlay(Palette.labelAlpha(0.07))
            }
            Button(model.copy.leaveTrip, action: model.requestLeave)
                .typeStyle(TypeStyle(size: 14, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.6))
                .buttonStyle(.plain)
                .frame(maxWidth: .infinity).padding(.vertical, metrics.points(12))
        }
        .surfaceCard()
    }
}

struct DecisionCard: View {
    @Environment(\.designMetrics) private var metrics
    let decision: GroupDecision
    let model: GroupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            HStack(alignment: .firstTextBaseline) {
                Text(decision.question).typeStyle(TypeStyle(size: 16, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                Spacer()
                Text(decision.meta).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.55))
            }
            ForEach(decision.options) { option in
                let isLeading = option.votes == decision.leadingVotes && option.votes > 0
                Button { model.vote(option, in: decision) } label: {
                    HStack(spacing: metrics.points(11)) {
                        VStack(alignment: .leading, spacing: metrics.points(5)) {
                            HStack {
                                Text(option.name).typeStyle(TypeStyle(size: 14, weight: isLeading ? .bold : .medium)).foregroundStyle(Palette.label)
                                Spacer()
                                Text(model.votesLabel(option)).typeStyle(TypeStyle(size: 12, weight: .semibold)).foregroundStyle(isLeading ? Palette.matchTeal : Palette.labelAlpha(0.55))
                            }
                            GradientProgressBar(fraction: Double(option.votes) / Double(decision.leadingVotes), isMuted: !isLeading)
                        }
                        Text(option.hasMyVote ? "✓" : "+").typeStyle(TypeStyle(size: 13, weight: .bold)).foregroundStyle(option.hasMyVote ? Palette.sky : Palette.labelAlpha(0.45))
                            .frame(width: metrics.points(34), height: metrics.points(34))
                            .background(option.hasMyVote ? Palette.sky.opacity(0.14) : .clear, in: Circle())
                            .overlay(Circle().strokeBorder(option.hasMyVote ? Palette.sky.opacity(0.35) : Palette.labelAlpha(0.15), lineWidth: 0.5))
                    }
                    .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
                .accessibilityAddTraits(option.hasMyVote ? .isSelected : [])
            }
        }
        .padding(EdgeInsets(top: metrics.points(14), leading: metrics.points(15), bottom: metrics.points(14), trailing: metrics.points(15)))
        .surfaceCard()
    }
}
