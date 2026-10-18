public struct PermissionCopy: Equatable, Sendable {
    public let locationTitle: String
    public let locationBody: String
    public let locationDeniedTitle: String
    public let locationDeniedBody: String
    public let enterLocation: String
    public let continueWithout: String
    public let notificationsTitle: String
    public let notificationsBody: String
    public let notificationsDeniedTitle: String
    public let notificationsDeniedBody: String
    public let cameraTitle: String
    public let cameraBody: String
    public let cameraDeniedTitle: String
    public let cameraDeniedBody: String
    public let microphoneTitle: String
    public let microphoneBody: String
    public let microphoneDeniedTitle: String
    public let microphoneDeniedBody: String
    public let allow: String
    public let notNow: String
    public let openSettings: String
    public let typeInstead: String

    public static let english = PermissionCopy(
        locationTitle: "Find what’s around you", locationBody: "Allow location to show nearby places and travel times.",
        locationDeniedTitle: "Location is off", locationDeniedBody: "Turn it on in Settings, or keep exploring from the centre of Funchal.",
        enterLocation: "Use destination centre", continueWithout: "Continue without location",
        notificationsTitle: "Stay ahead of changes", notificationsBody: "Get flight changes, weather shifts, leave reminders and reservation reminders while you travel.",
        notificationsDeniedTitle: "Notifications are off", notificationsDeniedBody: "You’ll still see updates here. Turn notifications on in Settings to be told as things change.",
        cameraTitle: "Read menus and signs", cameraBody: "Allow the camera to translate text you point it at.",
        cameraDeniedTitle: "Camera access is off", cameraDeniedBody: "Turn it on in Settings, or type the text to translate it.",
        microphoneTitle: "Speak to translate", microphoneBody: "Allow the microphone so NUVRO can hear what you say.",
        microphoneDeniedTitle: "Microphone access is off", microphoneDeniedBody: "Turn it on in Settings, or type what you want to say. Phrase cards still work.",
        allow: "Continue", notNow: "Not now", openSettings: "Open Settings", typeInstead: "Type instead"
    )

    public static func resolve(_ language: Language) -> PermissionCopy { english }
}
