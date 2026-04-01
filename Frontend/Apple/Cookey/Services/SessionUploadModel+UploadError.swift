import Foundation

extension SessionUploadModel {
    enum UploadError: LocalizedError {
        case invalidRecipientPublicKey
        case emptySessionPayload
        case invalidSessionPayload

        var errorDescription: String? {
            switch self {
            case .invalidRecipientPublicKey:
                String(localized: "The login request contains an invalid recipient key.")
            case .emptySessionPayload:
                String(localized: "The captured browser session was empty. Reload the page, complete login, and try sending again.")
            case .invalidSessionPayload:
                String(localized: "The captured browser session was invalid. Reload the page and try sending again.")
            }
        }
    }
}
