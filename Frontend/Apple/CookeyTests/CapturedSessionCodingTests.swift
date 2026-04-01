@testable import Cookey
import Foundation
import Testing

@MainActor
struct CapturedSessionCodingTests {
    @Test("CapturedSession coding preserves device_info when present")
    func encodesAndDecodesWithDeviceInfo() throws {
        let session = CapturedSession(
            cookies: [
                CapturedCookie(
                    name: "session",
                    value: "abc",
                    domain: "example.com",
                    path: "/",
                    expires: -1,
                    httpOnly: true,
                    secure: true,
                    sameSite: "Lax"
                ),
            ],
            origins: [
                CapturedOrigin(
                    origin: "https://example.com",
                    localStorage: [CapturedStorageItem(name: "token", value: "value")]
                ),
            ],
            deviceInfo: DeviceInfo(
                deviceID: "device-123",
                apnToken: "token-123",
                apnEnvironment: "sandbox",
                publicKey: "public-key"
            )
        )

        let data = try JSONEncoder().encode(session)
        let json = String(decoding: data, as: UTF8.self)
        #expect(json.contains("\"device_info\""))
        #expect(json.contains("\"apn_token\""))
        #expect(json.contains("\"public_key\""))

        let decoded = try JSONDecoder().decode(CapturedSession.self, from: data)
        #expect(decoded == session)
    }

    @Test("CapturedSession coding omits device_info when absent")
    func encodesAndDecodesWithoutDeviceInfo() throws {
        let session = CapturedSession(
            cookies: [],
            origins: [],
            deviceInfo: nil
        )

        let data = try JSONEncoder().encode(session)
        let json = String(decoding: data, as: UTF8.self)
        #expect(!json.contains("\"device_info\""))

        let decoded = try JSONDecoder().decode(CapturedSession.self, from: data)
        #expect(decoded.deviceInfo == nil)
        #expect(decoded.cookies.isEmpty)
        #expect(decoded.origins.isEmpty)
    }
}
