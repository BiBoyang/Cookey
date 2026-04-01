import Foundation

enum NotificationPromptStore {
    static func response(for serverURL: URL) -> NotificationPromptResponse? {
        guard let rawValue = UserDefaults.standard.string(forKey: key(for: serverURL)) else {
            return nil
        }
        return NotificationPromptResponse(rawValue: rawValue)
    }

    static func store(_ response: NotificationPromptResponse, for serverURL: URL) {
        UserDefaults.standard.set(response.rawValue, forKey: key(for: serverURL))
    }

    private static func key(for serverURL: URL) -> String {
        "apn_prompt_state::\(serverURL.absoluteString)"
    }
}
