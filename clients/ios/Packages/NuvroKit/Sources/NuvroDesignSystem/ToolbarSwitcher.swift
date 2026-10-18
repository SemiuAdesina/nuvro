import SwiftUI

public struct ToolbarSwitcher<Option: Hashable>: ToolbarContent {
    private let options: [Option]
    @Binding private var selection: Option
    private let width: CGFloat
    private let label: (Option) -> String

    public init(options: [Option], selection: Binding<Option>, width: CGFloat = 136, label: @escaping (Option) -> String) {
        self.options = options
        self._selection = selection
        self.width = width
        self.label = label
    }

    public var body: some ToolbarContent {
        if #available(iOS 26, macOS 26, *) {
            ToolbarItem(placement: .primaryAction) { control }.sharedBackgroundVisibility(.hidden)
        } else {
            ToolbarItem(placement: .primaryAction) { control }
        }
    }

    private var control: some View {
        SegmentedChoice(options: options, selection: $selection, compact: true, label: label)
            .frame(width: width)
    }
}
