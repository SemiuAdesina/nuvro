import SwiftUI

public struct TimelineRow<Content: View>: View {
    @Environment(\.designMetrics) private var metrics
    let time: String
    let isFirst: Bool
    let isLast: Bool
    let isEmphasised: Bool
    let travelToNext: String?
    let content: Content

    public init(time: String, isFirst: Bool, isLast: Bool, isEmphasised: Bool = false, travelToNext: String? = nil, @ViewBuilder content: () -> Content) {
        self.time = time
        self.isFirst = isFirst
        self.isLast = isLast
        self.isEmphasised = isEmphasised
        self.travelToNext = travelToNext
        self.content = content()
    }

    public var body: some View {
        HStack(alignment: .top, spacing: metrics.points(12)) {
            Text(time)
                .typeStyle(TypeStyle(size: 13, weight: .bold, tracking: -0.1))
                .foregroundStyle(isEmphasised ? Palette.sky : Palette.labelAlpha(0.7))
                .frame(width: metrics.points(44), alignment: .trailing)
                .padding(.top, metrics.points(2))
            VStack(spacing: metrics.points(4)) {
                Circle()
                    .fill(isFirst ? Palette.sky : Palette.aqua)
                    .overlay(Circle().strokeBorder(Palette.background, lineWidth: metrics.points(2.5)))
                    .overlay(Circle().strokeBorder(isFirst ? Palette.sky : Palette.aqua, lineWidth: 1.5).padding(-1.5))
                    .frame(width: metrics.points(11), height: metrics.points(11))
                    .padding(.top, metrics.points(4))
                if !isLast {
                    Rectangle().fill(Palette.sky.opacity(0.18)).frame(width: 2).frame(maxHeight: .infinity)
                }
            }
            .frame(width: metrics.points(14))
            VStack(alignment: .leading, spacing: metrics.points(8)) {
                content
                if let travelToNext, !isLast {
                    TravelTimeConnector(label: travelToNext)
                }
            }
            .padding(.bottom, isLast ? 0 : metrics.points(12))
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .fixedSize(horizontal: false, vertical: true)
    }
}

public struct TravelTimeConnector: View {
    @Environment(\.designMetrics) private var metrics
    let label: String

    public init(label: String) {
        self.label = label
    }

    public var body: some View {
        HStack(spacing: metrics.points(7)) {
            Rectangle().fill(Palette.labelAlpha(0.2)).frame(width: metrics.points(16), height: 1)
            Text(label).typeStyle(TypeStyle(size: 11.5, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.5))
        }
        .accessibilityLabel("Then \(label)")
    }
}

public struct DaySelector: View {
    @Environment(\.designMetrics) private var metrics
    let days: [(number: Int, label: String, date: String)]
    @Binding var selected: Int

    public init(days: [(number: Int, label: String, date: String)], selected: Binding<Int>) {
        self.days = days
        self._selected = selected
    }

    public var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: metrics.points(8)) {
                ForEach(days, id: \.number) { day in
                    let isSelected = day.number == selected
                    Button { selected = day.number } label: {
                        VStack(spacing: metrics.points(2)) {
                            Text(day.label).typeStyle(Typography.smallEmphasis).foregroundStyle(isSelected ? Color.white.opacity(0.8) : Palette.labelAlpha(0.55))
                            Text(day.date).typeStyle(TypeStyle(size: 16, weight: .bold, tracking: -0.2)).foregroundStyle(isSelected ? .white : Palette.label)
                        }
                        .frame(width: metrics.points(62))
                        .padding(.vertical, metrics.points(10))
                        .background(isSelected ? Palette.sky : Palette.surface, in: RoundedRectangle(cornerRadius: metrics.points(16), style: .continuous))
                        .overlay(RoundedRectangle(cornerRadius: metrics.points(16), style: .continuous).strokeBorder(isSelected ? Palette.sky : Palette.labelAlpha(0.08), lineWidth: 0.5))
                        .shadow(color: isSelected ? Palette.sky.opacity(0.14) : .clear, radius: isSelected ? metrics.points(5) : 0, y: isSelected ? metrics.points(2) : 0)
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.96))
                    .accessibilityLabel("\(day.label), \(day.date)")
                    .accessibilityAddTraits(isSelected ? .isSelected : [])
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.vertical, metrics.points(2))
        }
        .scrollClipDisabled()
        .animation(Motion.fade(0.25), value: selected)
    }
}
