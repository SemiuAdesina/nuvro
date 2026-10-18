import NuvroDesignSystem
import SwiftUI

enum BudgetCategoryTone {
    static func color(for category: String) -> Color {
        switch category.lowercased() {
        case "flight", "flights": Palette.sky
        case "stay": Palette.twilight
        case "food": Palette.gold
        case "activities": Palette.aqua
        default: Palette.lavender
        }
    }
}
