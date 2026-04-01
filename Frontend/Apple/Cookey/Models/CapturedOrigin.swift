import Foundation

struct CapturedOrigin: Codable, Equatable {
    let origin: String
    let localStorage: [CapturedStorageItem]
}
