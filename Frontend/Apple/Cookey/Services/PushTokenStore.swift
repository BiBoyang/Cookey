import Foundation

@MainActor
enum PushTokenStore {
    private static let tokenKey = "wiki.qaq.cookey.push-token"
    private static let environmentKey = "wiki.qaq.cookey.push-environment"

    static var currentToken: String? {
        get {
            let value = UserDefaults.standard.string(forKey: tokenKey)?.trimmingCharacters(in: .whitespacesAndNewlines)
            return value?.isEmpty == false ? value : nil
        }
        set {
            if let newValue, !newValue.isEmpty {
                UserDefaults.standard.set(newValue, forKey: tokenKey)
            } else {
                UserDefaults.standard.removeObject(forKey: tokenKey)
            }
        }
    }

    static var currentEnvironment: String? {
        get {
            let value = UserDefaults.standard.string(forKey: environmentKey)?.trimmingCharacters(in: .whitespacesAndNewlines)
            return value?.isEmpty == false ? value : nil
        }
        set {
            if let newValue, !newValue.isEmpty {
                UserDefaults.standard.set(newValue, forKey: environmentKey)
            } else {
                UserDefaults.standard.removeObject(forKey: environmentKey)
            }
        }
    }

    static var tokenDescription: String {
        currentToken ?? String(localized: "Unavailable")
    }
}
