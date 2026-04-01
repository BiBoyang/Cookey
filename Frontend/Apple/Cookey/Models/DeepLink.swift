import Foundation

struct DeepLink: Equatable {
    enum RequestType: String, Equatable {
        case login
        case refresh
    }

    let rid: String
    let serverURL: URL
    let targetURL: URL
    let recipientPublicKeyBase64: String
    let deviceID: String
    let requestType: RequestType

    init?(url: URL) {
        guard
            let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
            components.scheme?.lowercased() == "cookey",
            components.host?.lowercased() == "login"
        else {
            return nil
        }

        var values: [String: String] = [:]
        for item in components.queryItems ?? [] {
            guard let value = item.value else { continue }
            values[item.name] = value.removingPercentEncoding ?? value
        }

        guard
            let rid = values["rid"], !rid.isEmpty,
            let serverValue = values["server"], let serverURL = URL(string: serverValue),
            let targetValue = values["target"], let targetURL = URL(string: targetValue),
            let publicKey = values["pubkey"], !publicKey.isEmpty,
            let deviceID = values["device_id"], !deviceID.isEmpty
        else {
            return nil
        }

        guard
            let serverScheme = serverURL.scheme?.lowercased(),
            serverScheme == "https" || serverScheme == "http",
            let targetScheme = targetURL.scheme?.lowercased(),
            targetScheme == "https" || targetScheme == "http"
        else {
            return nil
        }

        self.rid = rid
        self.serverURL = serverURL
        self.targetURL = targetURL
        recipientPublicKeyBase64 = publicKey
        self.deviceID = deviceID
        requestType = RequestType(rawValue: values["request_type"]?.lowercased() ?? "") ?? .login
    }
}
