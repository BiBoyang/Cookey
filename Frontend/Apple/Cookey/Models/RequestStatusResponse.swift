import Foundation

struct RequestStatusResponse: Codable, Equatable {
    let rid: String
    let status: String
    let targetURL: String
    let requestType: String?
    let createdAt: Date
    let expiresAt: Date

    enum CodingKeys: String, CodingKey {
        case rid, status
        case targetURL = "target_url"
        case requestType = "request_type"
        case createdAt = "created_at"
        case expiresAt = "expires_at"
    }

    var isExpired: Bool {
        status == "expired" || expiresAt < Date()
    }
}
