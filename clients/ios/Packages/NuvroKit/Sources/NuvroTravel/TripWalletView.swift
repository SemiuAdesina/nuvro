import NuvroDesignSystem
import NuvroModels
import SwiftUI

public struct TripWalletView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(TransientFeedback.self) private var feedback: TransientFeedback?
    @Bindable private var model: TripWalletModel
    private let destinationName: String
    private let dateRange: String
    private let onNext: () -> Void

    public init(model: TripWalletModel, destinationName: String, dateRange: String, onNext: @escaping () -> Void) {
        self.model = model
        self.destinationName = destinationName
        self.dateRange = dateRange
        self.onNext = onNext
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                if model.documents.isEmpty {
                    NuvroEmptyState(symbol: "ticket", title: model.copy.emptyTitle, message: model.copy.emptyBody).padding(.top, metrics.points(24))
                }
                if let next = model.neededNext {
                    TicketCard(document: next, model: model) { model.openTicket = next }
                }
                ForEach(model.groups, id: \.category) { group in
                    VStack(alignment: .leading, spacing: metrics.points(8)) {
                        HStack(alignment: .firstTextBaseline, spacing: metrics.points(8)) {
                            Text(group.title).typeStyle(Typography.cardHeading).foregroundStyle(Palette.label)
                            Text("\(group.documents.count) \(group.documents.count == 1 ? model.copy.document : model.copy.documents)").typeStyle(Typography.small).foregroundStyle(Palette.labelAlpha(0.45))
                        }
                        VStack(spacing: 0) {
                            ForEach(group.documents) { document in
                                DocumentRow(document: document, model: model)
                                if document.id != group.documents.last?.id { Divider().overlay(Palette.labelAlpha(0.07)).padding(.leading, metrics.points(58)) }
                            }
                        }
                        .surfaceCard()
                    }
                }
                HStack(alignment: .top, spacing: metrics.points(8)) {
                    OfflineMark().padding(.top, 1)
                    Text(model.copy.footer).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                }
                .padding(.horizontal, metrics.points(4))
                InlineActions {
                    CallToActionButton(model.allDownloaded ? model.copy.allOffline : model.copy.downloadAll, action: model.downloadAll)
                        .disabled(model.allDownloaded)
                    HStack(spacing: metrics.points(24)) {
                        Button(model.copy.addAllToWallet) { model.addAllToWallet() }
                        Button("Trip updates", action: onNext)
                    }
                    .typeStyle(Typography.secondaryAction).foregroundStyle(Palette.sky).frame(height: metrics.points(34))
                    .buttonStyle(.plain)
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(40))
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(model.copy.title, subtitle: "\(destinationName) · \(dateRange)")
        .alert(model.copy.removeDownloadTitle, isPresented: Binding(get: { model.pendingDownloadRemoval != nil }, set: { if !$0 { model.cancelRemoveDownload() } })) {
            Button(model.copy.removeDownload, role: .destructive) { model.confirmRemoveDownload() }
            Button(model.copy.cancel, role: .cancel) { model.cancelRemoveDownload() }
        } message: {
            Text(model.copy.removeDownloadBody)
        }
        .onChange(of: model.lastEvent) { _, event in if let event { feedback?.show(event, symbol: event == model.copy.walletUnavailable ? "wallet.pass" : "checkmark"); model.clearEvent() } }
        .ticketCover(item: $model.openTicket) { document in
            TicketFullScreen(document: document, model: model) { model.openTicket = nil }
        }
    }
}

struct TicketCard: View {
    @Environment(\.designMetrics) private var metrics
    let document: TripDocument
    let model: TripWalletModel
    let onShow: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            HStack(alignment: .top, spacing: metrics.points(12)) {
                VStack(alignment: .leading, spacing: metrics.points(4)) {
                    Text(model.copy.neededNext).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.5)).foregroundStyle(Palette.cloudAlpha(0.7))
                    Text(document.title).typeStyle(TypeStyle(size: 22, weight: .bold, tracking: -0.4)).foregroundStyle(Palette.cloud)
                    Text("\(document.whenLabel) · \(document.travelers) \(model.copy.travelers)").typeStyle(TypeStyle(size: 14, weight: .regular)).foregroundStyle(Palette.cloudAlpha(0.82))
                }
                Spacer(minLength: 0)
                BarcodeTile()
            }
            HStack(spacing: metrics.points(8)) {
                Image(systemName: "arrow.down").font(.system(size: metrics.points(11), weight: .bold))
                Text(document.isAvailableOffline ? model.copy.availableOffline : model.copy.notDownloaded).typeStyle(TypeStyle(size: 12.5, weight: .bold))
            }
            .foregroundStyle(Palette.cloud)
            .padding(.horizontal, metrics.points(12)).padding(.vertical, metrics.points(7))
            .background(Palette.aqua.opacity(0.22), in: Capsule())
            .overlay(Capsule().strokeBorder(Palette.aqua.opacity(0.45), lineWidth: 0.5))
            if document.isExpired {
                HStack(spacing: metrics.points(8)) {
                    Image(systemName: "clock.badge.xmark").font(.system(size: metrics.points(12), weight: .bold))
                    Text("\(model.copy.expired) · \(model.copy.expiredBody)").typeStyle(TypeStyle(size: 12.5, weight: .semibold))
                }
                .foregroundStyle(Palette.cloud)
                .accessibilityElement(children: .combine)
            }
            HStack(spacing: metrics.points(10)) {
                Button(action: onShow) {
                    Text(model.copy.showTicket).typeStyle(TypeStyle(size: 16, weight: .semibold)).foregroundStyle(.white).frame(maxWidth: .infinity).frame(height: metrics.points(48))
                }
                .buttonStyle(GlassCallToActionStyle())
                .disabled(document.isExpired)
                .opacity(document.isExpired ? 0.5 : 1)
                if document.canAddToWallet {
                    Button { model.addToWallet(document) } label: {
                        Text(model.copy.addToWallet).typeStyle(TypeStyle(size: 14.5, weight: .semibold)).foregroundStyle(Palette.cloud)
                            .padding(.horizontal, metrics.points(18)).frame(height: metrics.points(48))
                    }
                    .buttonStyle(GlassPillButtonStyle(style: .control))
                }
            }
        }
        .padding(EdgeInsets(top: metrics.points(16), leading: metrics.points(18), bottom: metrics.points(18), trailing: metrics.points(18)))
        .background(LinearGradient(colors: [Palette.twilightDeep, Palette.twilight, Palette.navy], startPoint: .topLeading, endPoint: .bottomTrailing), in: RoundedRectangle(cornerRadius: metrics.points(Radius.hero), style: .continuous))
        .shadow(color: Palette.labelAlpha(0.06), radius: metrics.points(6), y: metrics.points(2))
    }
}

struct BarcodeTile: View {
    @Environment(\.designMetrics) private var metrics

    var body: some View {
        HStack(alignment: .bottom, spacing: 2) {
            Rectangle().fill(Palette.cloud).frame(width: 2, height: metrics.points(20))
            Rectangle().fill(Palette.cloud).frame(width: 4, height: metrics.points(20))
            Rectangle().fill(Palette.cloudAlpha(0.5)).frame(width: 2, height: metrics.points(20))
            Rectangle().fill(Palette.cloud).frame(width: 3, height: metrics.points(20))
        }
        .frame(width: metrics.points(52), height: metrics.points(52))
        .background(Color.white.opacity(0.14), in: RoundedRectangle(cornerRadius: metrics.points(14), style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: metrics.points(14), style: .continuous).strokeBorder(Color.white.opacity(0.3), lineWidth: 0.5))
        .accessibilityHidden(true)
    }
}

struct DocumentRow: View {
    @Environment(\.designMetrics) private var metrics
    let document: TripDocument
    let model: TripWalletModel

    private var palette: PhotoPalette {
        switch document.category {
        case .flight: .sky
        case .stay: .twilight
        case .activity: document.isAvailableOffline ? .gold : .named("madeira")
        case .transfer: .aqua
        case .other: .named("guanacaste")
        }
    }

    var body: some View {
        HStack(spacing: metrics.points(12)) {
            PhotoStandIn(palette: palette, label: "", labelSize: 6)
                .frame(width: metrics.points(32), height: metrics.points(32))
                .clipShape(RoundedRectangle(cornerRadius: metrics.points(10), style: .continuous))
                .opacity(document.isExpired ? 0.5 : 1)
            VStack(alignment: .leading, spacing: metrics.points(2)) {
                HStack(spacing: metrics.points(8)) {
                    Text(document.title).typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(document.isExpired ? Palette.labelAlpha(0.55) : Palette.label)
                    if document.isExpired { TagChip(model.copy.expired, tone: .twilight) }
                }
                HStack(spacing: metrics.points(8)) {
                    Text(document.subtitle).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                    Button { model.toggleReveal(document) } label: {
                        Text(model.reference(for: document))
                            .font(model.isRevealed(document) ? .system(size: metrics.points(12), weight: .semibold, design: .monospaced) : .system(size: metrics.points(12), weight: .semibold))
                            .foregroundStyle(Palette.sky)
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel(model.isRevealed(document) ? "Hide reference" : model.copy.tapToReveal)
                }
            }
            Spacer(minLength: 0)
            HStack(spacing: metrics.points(7)) {
                let isOffline = model.isAvailableOffline(document)
                Image(systemName: isOffline ? "arrow.down" : "circle.dashed")
                    .font(.system(size: metrics.points(10), weight: .bold))
                    .foregroundStyle(isOffline ? Palette.matchTeal : Palette.labelAlpha(0.45))
                    .frame(width: metrics.points(20), height: metrics.points(20))
                    .background(isOffline ? Palette.aqua.opacity(0.25) : Palette.labelAlpha(0.07), in: RoundedRectangle(cornerRadius: metrics.points(6), style: .continuous))
                    .accessibilityLabel(isOffline ? model.copy.availableOffline : model.copy.notDownloaded)
                Image(systemName: "chevron.forward").font(.system(size: metrics.points(12), weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.3))
            }
        }
        .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(12))
        .contentShape(Rectangle())
        .onTapGesture { if !document.isExpired { model.openTicket = document } }
        .accessibilityAddTraits(.isButton)
        .accessibilityHint(document.isExpired ? model.copy.expiredBody : model.copy.showTicket)
        .contextMenu {
            if model.isAvailableOffline(document) {
                Button(model.copy.removeDownload, systemImage: "arrow.down.circle.dotted", role: .destructive) { model.requestRemoveDownload(document) }
            } else {
                Button(model.copy.download, systemImage: "arrow.down.circle") { model.download(document) }
            }
        }
    }
}

private extension View {
    func ticketCover<Item: Identifiable, Cover: View>(item: Binding<Item?>, @ViewBuilder content: @escaping (Item) -> Cover) -> some View {
        #if os(iOS)
        fullScreenCover(item: item, content: content)
        #else
        sheet(item: item, content: content)
        #endif
    }
}
