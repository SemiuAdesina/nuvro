import Foundation

public enum LoadFailure: Equatable, Sendable {
    case offline
    case unavailable
    case unauthorized
    case unexpected

    public init(apiFailure: APIFailure) {
        switch apiFailure.code {
        case .authenticationRequired, .forbidden: self = .unauthorized
        case .upstreamUnavailable, .providerTimeout, .rateLimited: self = .unavailable
        case .validationError, .notFound, .methodNotAllowed, .conflict, .invalidState, .idempotencyConflict,
             .internalError, .payloadTooLarge, .unsupportedMediaType: self = .unexpected
        }
    }

    public init(error: any Error) {
        if let apiFailure = error as? APIFailure {
            self.init(apiFailure: apiFailure)
        } else if let urlError = error as? URLError, [.notConnectedToInternet, .networkConnectionLost, .timedOut, .cannotFindHost, .cannotConnectToHost].contains(urlError.code) {
            self = .offline
        } else {
            self = .unexpected
        }
    }

    public var isRetryable: Bool { self != .unauthorized }
}

public enum LoadState<Value: Equatable & Sendable>: Equatable, Sendable {
    case idle
    case loading
    case loaded(Value)
    case refreshing(Value)
    case failed(LoadFailure, cached: Value?)

    public var value: Value? {
        switch self {
        case .loaded(let value), .refreshing(let value): value
        case .failed(_, let cached): cached
        case .idle, .loading: nil
        }
    }

    public var isLoading: Bool {
        switch self {
        case .loading, .refreshing: true
        case .idle, .loaded, .failed: false
        }
    }

    public var failure: LoadFailure? {
        if case .failed(let failure, _) = self { return failure }
        return nil
    }

    public func beginning() -> LoadState {
        if let value { return .refreshing(value) }
        return .loading
    }

    public func failing(_ failure: LoadFailure) -> LoadState {
        .failed(failure, cached: value)
    }
}

public enum Freshness: String, Equatable, Sendable, Codable {
    case fresh
    case cached
    case stale
    case unavailable

    public var isCurrent: Bool { self == .fresh }
}

public enum ConnectivityStatus: Equatable, Sendable {
    case online
    case offline
    case reconnecting
}
