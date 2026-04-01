import Foundation

struct RelayClient {
    let baseURL: URL
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder
    private let session: URLSession
    private let requestExecutor: (@Sendable (URLRequest) throws -> (Data, URLResponse))?

    init(
        baseURL: URL,
        session: URLSession = .shared
    ) {
        self.baseURL = baseURL
        self.session = session
        requestExecutor = nil
        decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
    }

    init(
        baseURL: URL,
        session: URLSession = .shared,
        requestExecutor: @escaping @Sendable (URLRequest) throws -> (Data, URLResponse)
    ) {
        self.baseURL = baseURL
        self.session = session
        self.requestExecutor = requestExecutor
        decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
    }

    func healthCheck() async throws -> HealthCheckResult {
        let endpoint = baseURL.appending(path: "health")
        let (data, response) = try await session.data(from: endpoint)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw URLError(.badServerResponse)
        }

        guard (200 ..< 300).contains(httpResponse.statusCode) else {
            throw NSError(
                domain: "Cookey.RelayClient",
                code: httpResponse.statusCode,
                userInfo: [NSLocalizedDescriptionKey: "Unexpected status code \(httpResponse.statusCode)"]
            )
        }

        return HealthCheckResult(
            body: String(decoding: data, as: UTF8.self),
            serverName: httpResponse.value(forHTTPHeaderField: "Server") ?? "unknown",
            checkedAt: Date()
        )
    }

    func uploadSession(rid: String, envelope: EncryptedSessionEnvelope) async throws {
        let endpoint = baseURL.appending(path: "v1/requests/\(rid)/session")
        _ = try await sendRequest(to: endpoint, method: "POST", body: envelope)
    }

    func fetchRequestStatus(rid: String) async throws -> RequestStatusResponse {
        let endpoint = baseURL.appending(path: "v1/requests/\(rid)")
        var request = URLRequest(url: endpoint)
        request.httpMethod = "GET"

        let (data, response) = try await perform(request)
        guard let httpResponse = response as? HTTPURLResponse else {
            throw URLError(.badServerResponse)
        }

        guard (200 ..< 300).contains(httpResponse.statusCode) else {
            let body = String(decoding: data, as: UTF8.self)
            throw NSError(
                domain: "Cookey.RelayClient",
                code: httpResponse.statusCode,
                userInfo: [
                    NSLocalizedDescriptionKey: "Unexpected status code \(httpResponse.statusCode): \(body)",
                ]
            )
        }

        return try decoder.decode(RequestStatusResponse.self, from: data)
    }

    func fetchSeedSession(rid: String) async throws -> EncryptedSessionEnvelope? {
        let endpoint = baseURL.appending(path: "v1/requests/\(rid)/seed-session")
        var request = URLRequest(url: endpoint)
        request.httpMethod = "GET"

        let (data, response) = try await perform(request)
        guard let httpResponse = response as? HTTPURLResponse else {
            throw URLError(.badServerResponse)
        }

        switch httpResponse.statusCode {
        case 404:
            return nil
        case 200 ..< 300:
            return try decoder.decode(EncryptedSessionEnvelope.self, from: data)
        default:
            let body = String(decoding: data, as: UTF8.self)
            throw NSError(
                domain: "Cookey.RelayClient",
                code: httpResponse.statusCode,
                userInfo: [
                    NSLocalizedDescriptionKey: "Unexpected status code \(httpResponse.statusCode): \(body)",
                ]
            )
        }
    }

    @discardableResult
    private func sendRequest(
        to url: URL,
        method: String,
        body: (some Encodable)?
    ) async throws -> HTTPURLResponse {
        var request = URLRequest(url: url)
        request.httpMethod = method
        if let body {
            request.httpBody = try encoder.encode(body)
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        }

        let (data, response) = try await perform(request)
        guard let httpResponse = response as? HTTPURLResponse else {
            throw URLError(.badServerResponse)
        }

        guard (200 ..< 300).contains(httpResponse.statusCode) else {
            let body = String(decoding: data, as: UTF8.self)
            throw NSError(
                domain: "Cookey.RelayClient",
                code: httpResponse.statusCode,
                userInfo: [
                    NSLocalizedDescriptionKey: "Unexpected status code \(httpResponse.statusCode): \(body)",
                ]
            )
        }

        return httpResponse
    }

    private func perform(_ request: URLRequest) async throws -> (Data, URLResponse) {
        if let requestExecutor {
            return try requestExecutor(request)
        }
        return try await RelayClient.performRequest(request, with: session)
    }

    private static func performRequest(
        _ request: URLRequest,
        with session: URLSession
    ) async throws -> (Data, URLResponse) {
        try await session.data(for: request)
    }
}
