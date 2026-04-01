@testable import Cookey
import Foundation
import Testing

struct DeepLinkTests {
    @Test("DeepLink defaults request_type to login")
    func defaultsRequestTypeToLogin() throws {
        let url = try #require(
            URL(string: "cookey://login?rid=r_default&server=https%3A%2F%2Fapi.cookey.sh&target=https%3A%2F%2Fexample.com&pubkey=abc123&device_id=device-default")
        )

        let deepLink = try #require(DeepLink(url: url))
        #expect(deepLink.requestType == .login)
    }

    @Test("DeepLink parses request_type refresh")
    func parsesRefreshRequestType() throws {
        let url = try #require(
            URL(string: "cookey://login?rid=r_refresh&server=https%3A%2F%2Fapi.cookey.sh&target=https%3A%2F%2Fexample.com&pubkey=abc123&device_id=device-refresh&request_type=refresh")
        )

        let deepLink = try #require(DeepLink(url: url))
        #expect(deepLink.requestType == .refresh)
    }
}
