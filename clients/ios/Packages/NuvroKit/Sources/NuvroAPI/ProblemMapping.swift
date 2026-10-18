import Foundation
import NuvroModels

extension APIFailure {
    init(_ problem: Components.Schemas.Problem) {
        self.init(
            status: problem.status, code: Code(rawValue: problem.code.rawValue) ?? .internalError,
            title: problem.title, detail: problem.detail, requestID: problem.requestId
        )
    }

    static func from(_ response: Components.Responses.ComponentsResponsesProblem, statusCode: Int) -> APIFailure {
        switch response.body {
        case .applicationProblemJson(let problem): APIFailure(problem)
        }
    }
}
