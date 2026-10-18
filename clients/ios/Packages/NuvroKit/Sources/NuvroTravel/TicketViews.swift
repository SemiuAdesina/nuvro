import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct QRPlaceholder: View {
    @Environment(\.designMetrics) private var metrics
    let size: CGFloat
    var seed: String = "nuvro"

    private static let modules = 21

    var body: some View {
        Canvas { context, canvasSize in
            let cell = canvasSize.width / CGFloat(Self.modules)
            var state = seed.utf8.reduce(UInt32(2_166_136_261)) { ($0 ^ UInt32($1)) &* 16_777_619 }
            for row in 0..<Self.modules {
                for column in 0..<Self.modules {
                    state = state &* 1_103_515_245 &+ 12_345
                    let isDark = isFinder(row, column) ? isFinderDark(row, column) : (state >> 16) & 1 == 1
                    if isDark {
                        context.fill(Path(CGRect(x: CGFloat(column) * cell, y: CGFloat(row) * cell, width: cell, height: cell)), with: .color(Palette.navy))
                    }
                }
            }
        }
        .padding(metrics.points(size > 100 ? 18 : 6))
        .frame(width: metrics.points(size), height: metrics.points(size))
        .background(Color.white, in: RoundedRectangle(cornerRadius: metrics.points(size > 100 ? 22 : 10), style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: metrics.points(size > 100 ? 22 : 10), style: .continuous).strokeBorder(Palette.labelAlpha(0.1), lineWidth: 0.5))
        .shadow(color: Palette.navy.opacity(size > 100 ? 0.16 : 0), radius: metrics.points(20), y: metrics.points(14))
        .accessibilityHidden(true)
    }

    private func isFinder(_ row: Int, _ column: Int) -> Bool {
        let last = Self.modules - 7
        return (row < 7 && column < 7) || (row < 7 && column >= last) || (row >= last && column < 7)
    }

    private func isFinderDark(_ row: Int, _ column: Int) -> Bool {
        let last = Self.modules - 7
        let localRow = row < 7 ? row : row - last
        let localColumn = column < 7 ? column : column - last
        let ring = min(localRow, localColumn, 6 - localRow, 6 - localColumn)
        return ring == 0 || ring >= 2
    }
}

struct TicketFullScreen: View {
    @Environment(\.designMetrics) private var metrics
    let document: TripDocument
    let model: TripWalletModel
    let dismiss: () -> Void

    var body: some View {
        Button(action: dismiss) {
            VStack(spacing: metrics.points(22)) {
                Spacer()
                VStack(spacing: metrics.points(5)) {
                    Text(document.title).typeStyle(TypeStyle(size: 22, weight: .bold, tracking: -0.4)).foregroundStyle(Palette.label)
                    Text("\(document.whenLabel) · \(document.travelers) \(model.copy.travelers)").typeStyle(TypeStyle(size: 14, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                }
                if document.hasBarcode {
                    QRPlaceholder(size: 246, seed: document.reference)
                } else {
                    NuvroErrorState(symbol: "qrcode", title: model.copy.noBarcodeTitle, message: model.copy.noBarcodeBody)
                }
                VStack(spacing: metrics.points(4)) {
                    Text("TKT · \(document.reference)").font(.system(size: metrics.points(13), design: .monospaced)).tracking(1.3).foregroundStyle(Palette.labelAlpha(0.7))
                    Text("✓ \(model.copy.availableOffline)").typeStyle(TypeStyle(size: 12.5, weight: .semibold)).foregroundStyle(Palette.matchTeal)
                }
                Spacer()
                Text(model.copy.tapToClose).typeStyle(Typography.metaEmphasis).foregroundStyle(Palette.labelAlpha(0.45)).padding(.bottom, metrics.points(40))
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .padding(metrics.points(28))
            .background(Palette.surface.ignoresSafeArea())
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }
}
