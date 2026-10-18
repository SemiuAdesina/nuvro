import Foundation
#if canImport(UIKit)
import UIKit
#endif

public enum SystemSettings {
    public static var url: URL? {
        #if canImport(UIKit)
        URL(string: UIApplication.openSettingsURLString)
        #else
        nil
        #endif
    }
}
