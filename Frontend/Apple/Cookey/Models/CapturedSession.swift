import Foundation

struct DeviceInfo: Codable, Equatable {
    let deviceID: String
    let apnToken: String
    let apnEnvironment: String
    let publicKey: String

    enum CodingKeys: String, CodingKey {
        case deviceID = "device_id"
        case apnEnvironment = "apn_environment"
        case apnToken = "apn_token"
        case publicKey = "public_key"
    }
}

struct CapturedSession: Codable, Equatable {
    let cookies: [CapturedCookie]
    let origins: [CapturedOrigin]
    let deviceInfo: DeviceInfo?

    enum CodingKeys: String, CodingKey {
        case cookies
        case origins
        case deviceInfo = "device_info"
    }
}
