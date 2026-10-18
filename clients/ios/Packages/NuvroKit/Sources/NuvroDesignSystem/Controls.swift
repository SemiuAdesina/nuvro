import SwiftUI

public struct SegmentedChoice<Option: Hashable>: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @Namespace private var selectionSpace
    let options: [Option]
    let label: (Option) -> String
    @Binding var selection: Option
    let compact: Bool
    let isPill: Bool

    public init(options: [Option], selection: Binding<Option>, compact: Bool = false, isPill: Bool = true, label: @escaping (Option) -> String) {
        self.options = options
        self._selection = selection
        self.compact = compact
        self.isPill = isPill
        self.label = label
    }

    private var selectionRadius: CGFloat {
        isPill ? 999 : metrics.points(compact ? Radius.fineSegmentSelection : Radius.segmentSelection)
    }

    private var trackRadius: CGFloat {
        isPill ? 999 : metrics.points(compact ? Radius.fineSegment : Radius.segment)
    }

    private func showsSeparator(after index: Int) -> Bool {
        guard index < options.count - 1 else { return false }
        return options[index] != selection && options[index + 1] != selection
    }

    public var body: some View {
        HStack(spacing: 0) {
            ForEach(Array(options.enumerated()), id: \.element) { index, option in
                let isSelected = option == selection
                Button { selection = option } label: {
                    Text(label(option))
                        .typeStyle(TypeStyle(size: compact ? 13 : 15, weight: isSelected ? .semibold : .medium))
                        .foregroundStyle(Palette.label)
                        .lineLimit(1)
                        .fixedSize(horizontal: true, vertical: false)
                        .padding(.horizontal, metrics.points(compact ? 8 : 12))
                        .frame(maxWidth: .infinity)
                        .frame(height: metrics.points(compact ? 28 : 44))
                        .background {
                            if isSelected {
                                RoundedRectangle(cornerRadius: selectionRadius, style: .continuous)
                                    .fill(Palette.segmentSelection)
                                    .shadow(color: Palette.navy.opacity(0.10), radius: metrics.points(4), y: metrics.points(2))
                                    .matchedGeometryEffect(id: "selection", in: selectionSpace)
                            }
                        }
                        .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
                .accessibilityAddTraits(isSelected ? .isSelected : [])
                .overlay(alignment: .trailing) {
                    Rectangle()
                        .fill(Palette.labelAlpha(0.12))
                        .frame(width: 1, height: metrics.points(compact ? 14 : 20))
                        .opacity(showsSeparator(after: index) ? 1 : 0)
                        .accessibilityHidden(true)
                }
            }
        }
        .padding(metrics.points(2))
        .background(Palette.segmentTrack, in: RoundedRectangle(cornerRadius: trackRadius, style: .continuous))
        .animation(reduceMotion ? Motion.fade(0.2) : Motion.settle(0.32), value: selection)
        .accessibilityElement(children: .contain)
    }
}

public struct CountStepper: View {
    @Environment(\.designMetrics) private var metrics
    @Binding var value: Int
    let range: ClosedRange<Int>
    let unit: String

    public init(value: Binding<Int>, in range: ClosedRange<Int>, unit: String) {
        self._value = value
        self.range = range
        self.unit = unit
    }

    public var body: some View {
        HStack(spacing: metrics.points(10)) {
            stepButton("minus", enabled: value > range.lowerBound, tinted: false) { value -= 1 }
            stepButton("plus", enabled: value < range.upperBound, tinted: true) { value += 1 }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(unit)
        .accessibilityValue("\(value)")
    }

    private func stepButton(_ symbol: String, enabled: Bool, tinted: Bool, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            Image(systemName: symbol)
                .font(.system(size: metrics.points(14), weight: .semibold))
                .foregroundStyle(tinted ? Palette.sky : Palette.label)
                .frame(width: metrics.points(34), height: metrics.points(34))
                .background(tinted ? Palette.sky.opacity(0.12) : Palette.labelAlpha(0.07), in: Circle())
        }
        .buttonStyle(PressScaleButtonStyle(scale: 0.92))
        .disabled(!enabled)
        .opacity(enabled ? 1 : 0.4)
    }
}

public struct IconTile: View {
    @Environment(\.designMetrics) private var metrics
    let palette: PhotoPalette
    let systemImage: String
    let size: CGFloat

    public init(palette: PhotoPalette, systemImage: String, size: CGFloat = 28) {
        self.palette = palette
        self.systemImage = systemImage
        self.size = size
    }

    public var body: some View {
        RoundedRectangle(cornerRadius: metrics.points(Radius.iconTile), style: .continuous)
            .fill(palette.gradient)
            .frame(width: metrics.points(size), height: metrics.points(size))
            .overlay {
                Image(systemName: systemImage)
                    .font(.system(size: metrics.points(size * 0.5), weight: .semibold))
                    .foregroundStyle(Palette.cloud)
            }
            .accessibilityHidden(true)
    }
}

public struct StatusPill: View {
    @Environment(\.designMetrics) private var metrics
    let text: String

    public init(_ text: String) {
        self.text = text
    }

    public var body: some View {
        HStack(spacing: metrics.points(6)) {
            Circle().fill(Palette.gold).frame(width: metrics.points(6), height: metrics.points(6))
            Text(text).typeStyle(Typography.statusPill).foregroundStyle(Palette.amber)
        }
        .padding(.horizontal, metrics.points(12))
        .frame(height: metrics.points(30))
        .background(Palette.gold.opacity(0.25), in: Capsule())
    }
}
