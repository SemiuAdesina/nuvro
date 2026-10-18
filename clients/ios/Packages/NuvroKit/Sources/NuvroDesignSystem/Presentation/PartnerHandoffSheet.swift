import SwiftUI

public struct PartnerHandoffCopy: Sendable {
    public let title: String
    public let body: String
    public let continueLabel: String
    public let cancel: String
    public let pendingLink: String
    public let disclosure: String

    public init(title: String, body: String, continueLabel: String, cancel: String, pendingLink: String, disclosure: String) {
        self.title = title
        self.body = body
        self.continueLabel = continueLabel
        self.cancel = cancel
        self.pendingLink = pendingLink
        self.disclosure = disclosure
    }
}

public struct PartnerHandoffSheet: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.dismiss) private var dismiss
    @Environment(\.openURL) private var openURL
    let partnerName: String
    let url: URL?
    let copy: PartnerHandoffCopy

    public init(partnerName: String, url: URL?, copy: PartnerHandoffCopy) {
        self.partnerName = partnerName
        self.url = url
        self.copy = copy
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            IconTile(palette: .sky, systemImage: "arrow.up.forward.app", size: 34)
            Text("\(copy.title) \(partnerName)").typeStyle(Typography.cardHeading).foregroundStyle(Palette.label)
            Text(copy.body).typeStyle(Typography.bodyCompact).foregroundStyle(Palette.labelAlpha(0.65))
            Text(copy.disclosure).typeStyle(Typography.small).foregroundStyle(Palette.labelAlpha(0.5))
            if url == nil {
                NuvroInlineError(copy.pendingLink)
            }
            Spacer()
            CallToActionButton(copy.continueLabel) {
                if let url { openURL(url) }
                dismiss()
            }
            .disabled(url == nil)
            .opacity(url == nil ? 0.5 : 1)
            Button(copy.cancel) { dismiss() }
                .typeStyle(Typography.secondaryAction)
                .foregroundStyle(Palette.labelAlpha(0.7))
                .buttonStyle(.plain)
                .frame(maxWidth: .infinity)
        }
        .padding(metrics.points(24))
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(Palette.background)
        .presentationDetents([.medium])
        .presentationDragIndicator(.visible)
    }
}
