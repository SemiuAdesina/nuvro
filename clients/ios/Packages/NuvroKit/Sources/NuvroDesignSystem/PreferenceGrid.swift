import NuvroModels
import SwiftUI

public struct PreferenceGrid: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let countLabel: String
    let isSelected: (TravelPreference) -> Bool
    let onToggle: (TravelPreference) -> Void
    let options: [TravelPreference]

    public init(title: String, countLabel: String, options: [TravelPreference] = Array(TravelPreference.allCases.prefix(8)), isSelected: @escaping (TravelPreference) -> Bool, onToggle: @escaping (TravelPreference) -> Void) {
        self.title = title
        self.countLabel = countLabel
        self.options = options
        self.isSelected = isSelected
        self.onToggle = onToggle
    }

    private var columns: [GridItem] {
        Array(repeating: GridItem(.flexible(), spacing: metrics.points(10)), count: 4)
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            HStack(alignment: .firstTextBaseline) {
                Text(title).typeStyle(Typography.cardHeading).foregroundStyle(Palette.label)
                Spacer()
                Text(countLabel).typeStyle(Typography.meta).foregroundStyle(Palette.labelAlpha(0.5))
                    .contentTransition(.numericText())
            }
            LazyVGrid(columns: columns, spacing: metrics.points(12)) {
                ForEach(options) { preference in
                    PreferenceTile(preference: preference, isSelected: isSelected(preference)) { onToggle(preference) }
                }
            }
        }
    }
}

struct PreferenceTile: View {
    @Environment(\.designMetrics) private var metrics
    let preference: TravelPreference
    let isSelected: Bool
    let onToggle: () -> Void

    private var palette: PhotoPalette {
        switch preference {
        case .beach: .aqua
        case .food: .gold
        case .culture: .twilight
        case .nature: .rainforest
        case .nightlife: .sky
        case .adventure: .adventure
        case .relaxation: .lisbon
        case .family: .family
        case .shopping: .gold
        case .history: .twilight
        }
    }

    var body: some View {
        Button(action: onToggle) {
            VStack(spacing: metrics.points(7)) {
                ZStack(alignment: .topTrailing) {
                    RoundedRectangle(cornerRadius: metrics.points(Radius.tile), style: .continuous)
                        .fill(palette.gradient)
                        .overlay(DiagonalHatch().fill(Color.white.opacity(0.08)))
                        .overlay {
                            Image(systemName: preference.systemImage)
                                .font(.system(size: metrics.points(22), weight: .semibold))
                                .foregroundStyle(Palette.cloud)
                                .shadow(color: Palette.navy.opacity(0.25), radius: 3, y: 2)
                        }
                        .aspectRatio(1, contentMode: .fit)
                        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.tile), style: .continuous))
                        .overlay {
                            if isSelected {
                                RoundedRectangle(cornerRadius: metrics.points(Radius.tile), style: .continuous)
                                    .strokeBorder(Palette.cloud, lineWidth: metrics.points(3))
                                    .padding(-metrics.points(3))
                                    .overlay(
                                        RoundedRectangle(cornerRadius: metrics.points(Radius.tile) + metrics.points(5.5), style: .continuous)
                                            .strokeBorder(Palette.sky, lineWidth: metrics.points(2.5))
                                            .padding(-metrics.points(5.5))
                                    )
                            }
                        }
                        .shadow(color: .clear, radius: 0, y: metrics.points(6))
                    Image(systemName: "checkmark")
                        .font(.system(size: metrics.points(9), weight: .bold))
                        .foregroundStyle(.white)
                        .frame(width: metrics.points(18), height: metrics.points(18))
                        .background(Palette.sky, in: Circle())
                        .overlay(Circle().strokeBorder(.white, lineWidth: 2))
                        .padding(metrics.points(6))
                        .opacity(isSelected ? 1 : 0)
                }
                Text(preference.title)
                    .typeStyle(Typography.captionEmphasis)
                    .foregroundStyle(isSelected ? Palette.label : Palette.labelAlpha(0.6))
                    .lineLimit(1)
                    .minimumScaleFactor(0.8)
            }
        }
        .buttonStyle(PressScaleButtonStyle(scale: 0.96))
        .animation(Motion.fade(0.2), value: isSelected)
        .accessibilityLabel(preference.title)
        .accessibilityAddTraits(isSelected ? .isSelected : [])
    }
}
