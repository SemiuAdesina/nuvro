import NuvroDesignSystem
import SwiftUI

struct SetupQuestion: View {
    @Environment(\.designMetrics) private var metrics
    let question: String
    let supporting: String?
    var privacy: String? = nil

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(8)) {
            Text(question).typeStyle(Typography.headline).foregroundStyle(Palette.label).fixedSize(horizontal: false, vertical: true)
            if let supporting { Text(supporting).typeStyle(Typography.body).foregroundStyle(Palette.labelAlpha(0.62)).fixedSize(horizontal: false, vertical: true) }
            if let privacy {
                HStack(alignment: .top, spacing: metrics.points(6)) {
                    Image(systemName: "lock.fill").font(.system(size: metrics.points(11), weight: .semibold)).foregroundStyle(Palette.matchTeal).accessibilityHidden(true)
                    Text(privacy).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.55))
                }
                .padding(.top, metrics.points(2))
            }
        }
        .accessibilityElement(children: .combine)
    }
}

struct ChoiceRow: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    var detail: String? = nil
    let isSelected: Bool
    var allowsMultiple = false
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: metrics.points(12)) {
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    Text(title).typeStyle(TypeStyle(size: 16, weight: .semibold)).foregroundStyle(Palette.label)
                    if let detail { Text(detail).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.6)) }
                }
                .fixedSize(horizontal: false, vertical: true)
                Spacer(minLength: 0)
                Image(systemName: isSelected ? "checkmark.circle.fill" : (allowsMultiple ? "circle" : "circle"))
                    .font(.system(size: metrics.points(20), weight: .semibold))
                    .foregroundStyle(isSelected ? Palette.sky : Palette.labelAlpha(0.25))
                    .contentTransition(.symbolEffect(.replace))
                    .accessibilityHidden(true)
            }
            .padding(.horizontal, metrics.points(16)).padding(.vertical, metrics.points(14))
            .frame(maxWidth: .infinity)
            .background(isSelected ? Palette.sky.opacity(0.10) : Palette.surface, in: RoundedRectangle(cornerRadius: metrics.points(Radius.cardCompact), style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.cardCompact), style: .continuous).strokeBorder(isSelected ? Palette.sky.opacity(0.45) : Palette.labelAlpha(0.07), lineWidth: isSelected ? 1 : 0.5))
            .contentShape(Rectangle())
        }
        .buttonStyle(PressScaleButtonStyle(scale: 0.985))
        .accessibilityAddTraits(isSelected ? .isSelected : [])
        .animation(Motion.fade(0.2), value: isSelected)
    }
}

struct ChoiceChip: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Text(title)
                .typeStyle(Typography.control)
                .foregroundStyle(isSelected ? Palette.background : Palette.label)
                .padding(.horizontal, metrics.points(14)).frame(minHeight: metrics.points(38))
                .background(isSelected ? Palette.label : Palette.surface, in: Capsule())
                .overlay(Capsule().strokeBorder(isSelected ? Palette.label : Palette.labelAlpha(0.1), lineWidth: 0.5))
        }
        .buttonStyle(PressScaleButtonStyle(scale: 0.96))
        .accessibilityAddTraits(isSelected ? .isSelected : [])
        .animation(Motion.fade(0.2), value: isSelected)
    }
}

struct SelectionCard: View {
    @Environment(\.designMetrics) private var metrics
    let leading: String
    let title: String
    var detail: String? = nil
    let actionLabel: String
    let action: () -> Void
    var onRemove: (() -> Void)? = nil

    var body: some View {
        HStack(spacing: metrics.points(12)) {
            CodeTile(leading)
            VStack(alignment: .leading, spacing: 2) {
                Text(title).typeStyle(TypeStyle(size: 16, weight: .semibold)).foregroundStyle(Palette.label)
                if let detail { Text(detail).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.6)) }
            }
            Spacer(minLength: 0)
            if let onRemove {
                Button(action: onRemove) {
                    Image(systemName: "xmark.circle.fill").font(.system(size: metrics.points(18))).foregroundStyle(Palette.labelAlpha(0.35))
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Remove \(title)")
            } else {
                Button(actionLabel, action: action).typeStyle(Typography.link).foregroundStyle(Palette.sky).buttonStyle(.plain)
            }
        }
        .padding(metrics.points(14))
        .surfaceCard(radius: Radius.cardCompact)
        .accessibilityElement(children: .combine)
    }
}

struct AddRow: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: metrics.points(12)) {
                Image(systemName: "plus.circle.fill").font(.system(size: metrics.points(20))).foregroundStyle(Palette.sky).accessibilityHidden(true)
                Text(title).typeStyle(TypeStyle(size: 16, weight: .semibold)).foregroundStyle(Palette.sky)
                Spacer(minLength: 0)
            }
            .padding(.horizontal, metrics.points(16)).padding(.vertical, metrics.points(14))
            .frame(maxWidth: .infinity)
            .background(Palette.surface, in: RoundedRectangle(cornerRadius: metrics.points(Radius.cardCompact), style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.cardCompact), style: .continuous).strokeBorder(Palette.sky.opacity(0.35), style: StrokeStyle(lineWidth: 1, dash: [6, 4])))
            .contentShape(Rectangle())
        }
        .buttonStyle(PressScaleButtonStyle(scale: 0.985))
    }
}
