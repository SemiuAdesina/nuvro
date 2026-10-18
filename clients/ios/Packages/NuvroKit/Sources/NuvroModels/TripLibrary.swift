import Foundation

public struct TripLibraryEntry: Sendable, Hashable, Codable, Identifiable {
    public enum Phase: String, Sendable, Codable {
        case active, upcoming, planning, past
    }

    public let id: String
    public let destinationName: String
    public let dateLabel: String
    public let phase: Phase
    public let progress: Double
    public let readinessLabel: String?
    public let media: DestinationMedia
    public let statusLabel: String?
    public let weatherLabel: String?
    public let nextLabel: String?

    public init(id: String, destinationName: String, dateLabel: String, phase: Phase, progress: Double = 0, readinessLabel: String? = nil, media: DestinationMedia, statusLabel: String? = nil, weatherLabel: String? = nil, nextLabel: String? = nil) {
        self.id = id
        self.destinationName = destinationName
        self.dateLabel = dateLabel
        self.phase = phase
        self.progress = progress
        self.readinessLabel = readinessLabel
        self.media = media
        self.statusLabel = statusLabel
        self.weatherLabel = weatherLabel
        self.nextLabel = nextLabel
    }

    private enum CodingKeys: String, CodingKey {
        case id, phase, progress, media
        case destinationName = "destination_name"
        case dateLabel = "date_label"
        case readinessLabel = "readiness_label"
        case statusLabel = "status_label"
        case weatherLabel = "weather_label"
        case nextLabel = "next_label"
    }
}
