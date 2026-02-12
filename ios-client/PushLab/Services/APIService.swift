import Foundation

class APIService {
    static let shared = APIService()

    private let baseURL = Constants.apiBaseURL

    private init() {}

    // MARK: - Authentication

    func register(username: String, email: String, password: String) async throws -> (token: String, user: User) {
        let url = URL(string: "\(baseURL)/api/v1/auth/register")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let body = [
            "username": username,
            "email": email,
            "password": password
        ]
        request.httpBody = try JSONEncoder().encode(body)

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 201 else {
            throw APIError.invalidResponse
        }

        let loginResponse = try JSONDecoder().decode(LoginResponse.self, from: data)
        return (loginResponse.token, loginResponse.user)
    }

    func login(username: String, password: String) async throws -> (token: String, user: User) {
        let url = URL(string: "\(baseURL)/api/v1/auth/login")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let body = [
            "username": username,
            "password": password
        ]
        request.httpBody = try JSONEncoder().encode(body)

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw APIError.unauthorized
        }

        let loginResponse = try JSONDecoder().decode(LoginResponse.self, from: data)
        return (loginResponse.token, loginResponse.user)
    }

    // MARK: - Device Management

    func registerDevice(deviceToken: String) async {
        guard let token = KeychainHelper.getToken() else { return }

        let url = URL(string: "\(baseURL)/api/v1/devices")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        let deviceName = UIDevice.current.name
        let deviceIdentifier = UIDevice.current.identifierForVendor?.uuidString ?? UUID().uuidString
        let tags = UserDefaults.standard.stringArray(forKey: "deviceTags") ?? []

        let body: [String: Any] = [
            "device_name": deviceName,
            "device_identifier": deviceIdentifier,
            "device_token": deviceToken,
            "bundle_id": Constants.bundleID,
            "environment": Constants.apnsEnvironment,
            "tags": tags
        ]

        request.httpBody = try? JSONSerialization.data(withJSONObject: body)

        do {
            let (_, response) = try await URLSession.shared.data(for: request)
            if let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 || httpResponse.statusCode == 201 {
                print("Device registered successfully")
            }
        } catch {
            print("Failed to register device: \(error.localizedDescription)")
        }
    }

    func fetchDevices() async throws -> [Device] {
        guard let token = KeychainHelper.getToken() else {
            throw APIError.unauthorized
        }

        let url = URL(string: "\(baseURL)/api/v1/devices")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        let (data, _) = try await URLSession.shared.data(for: request)
        return try JSONDecoder().decode([Device].self, from: data)
    }

    func updateDevice(id: String, name: String?, tags: [String]?) async throws {
        guard let token = KeychainHelper.getToken() else {
            throw APIError.unauthorized
        }

        let url = URL(string: "\(baseURL)/api/v1/devices/\(id)")!
        var request = URLRequest(url: url)
        request.httpMethod = "PUT"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        var body: [String: Any] = [:]
        if let name = name { body["device_name"] = name }
        if let tags = tags { body["tags"] = tags }

        request.httpBody = try? JSONSerialization.data(withJSONObject: body)

        let (_, _) = try await URLSession.shared.data(for: request)
    }

    // MARK: - Notifications

    func fetchNotifications(limit: Int = 50, offset: Int = 0) async throws -> [PushNotification] {
        guard let token = KeychainHelper.getToken() else {
            throw APIError.unauthorized
        }

        let url = URL(string: "\(baseURL)/api/v1/notifications?limit=\(limit)&offset=\(offset)")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        let (data, _) = try await URLSession.shared.data(for: request)
        return try JSONDecoder().decode([PushNotification].self, from: data)
    }
}

// MARK: - Supporting Types

enum APIError: Error {
    case invalidResponse
    case unauthorized
    case networkError
}

struct LoginResponse: Codable {
    let token: String
    let user: User
}
