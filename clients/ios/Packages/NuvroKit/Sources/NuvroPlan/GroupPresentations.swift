import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct GroupPresentations: ViewModifier {
    @Bindable var model: GroupModel

    private var actionsMember: GroupMember? {
        if case .memberActions(let member) = model.presentation { return member }
        return nil
    }

    private var removalMember: GroupMember? {
        if case .removeConfirmation(let member) = model.presentation { return member }
        return nil
    }

    func body(content: Content) -> some View {
        content
            .confirmationDialog(actionsMember?.name ?? "", isPresented: Binding(get: { actionsMember != nil }, set: { if !$0 { model.dismissPresentation() } }), titleVisibility: .visible) {
                if let member = actionsMember {
                    ForEach([GroupMember.Role.editor, .viewer], id: \.self) { role in
                        if role != member.role { Button("\(model.copy.changeRole): \(model.roleLabel(role))") { model.setRole(role, for: member) } }
                    }
                    Button(model.copy.removeTraveller, role: .destructive) { model.requestRemove(member) }
                    Button(model.copy.cancel, role: .cancel) { model.dismissPresentation() }
                }
            }
            .alert(model.copy.removeTitle, isPresented: Binding(get: { removalMember != nil }, set: { if !$0 { model.dismissPresentation() } })) {
                Button(model.copy.removeTraveller, role: .destructive) { model.confirmRemove() }
                Button(model.copy.cancel, role: .cancel) { model.dismissPresentation() }
            } message: {
                Text(model.copy.removeBody)
            }
            .alert(model.copy.leaveTitle, isPresented: Binding(get: { model.presentation == .leaveConfirmation }, set: { if !$0 { model.dismissPresentation() } })) {
                Button(model.copy.leaveTrip, role: .destructive) { model.confirmLeave() }
                Button(model.copy.cancel, role: .cancel) { model.dismissPresentation() }
            } message: {
                Text(model.copy.leaveBody)
            }
            .sheet(isPresented: Binding(get: { model.presentation == .invite }, set: { if !$0 { model.dismissPresentation() } })) {
                InviteTravellerSheet(model: model)
            }
    }
}

struct InviteTravellerSheet: View {
    @Environment(\.designMetrics) private var metrics
    @FocusState private var isFocused: Bool
    @Bindable var model: GroupModel

    var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: metrics.points(14)) {
                Text(model.copy.inviteBody).typeStyle(Typography.body).foregroundStyle(Palette.labelAlpha(0.65))
                TextField(model.copy.inviteEmail, text: $model.inviteEmail)
                    .typeStyle(Typography.fieldValue)
                    .textContentType(.emailAddress)
                    .autocorrectionDisabled()
                    .submitLabel(.send)
                    .focused($isFocused)
                    .onSubmit(model.sendInvite)
                    .padding(.horizontal, metrics.points(16)).frame(height: metrics.points(52))
                    .surfaceCard(radius: Radius.field)
                    #if os(iOS)
                    .keyboardType(.emailAddress)
                    .textInputAutocapitalization(.never)
                    #endif
                Spacer()
                CallToActionButton(model.copy.sendInvite, action: model.sendInvite)
                    .disabled(!model.canInvite)
                    .opacity(model.canInvite ? 1 : 0.55)
            }
            .padding(metrics.points(24))
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            .background(Palette.background)
            .navigationTitle(model.copy.inviteTitle)
            .toolbar { ToolbarItem(placement: .cancellationAction) { Button(model.copy.cancel) { model.dismissPresentation() } } }
            .toolbarTitleDisplayMode(.inline)
            .onAppear { isFocused = true }
        }
        .presentationDetents([.medium])
    }
}

extension View {
    func groupPresentations(model: GroupModel) -> some View {
        modifier(GroupPresentations(model: model))
    }
}
