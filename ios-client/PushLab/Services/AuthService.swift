import Foundation
import Combine

class AuthService: ObservableObject {
    static let shared = AuthService()

    @Published var isAuthenticated = false
    @Published var currentUser: User?

    private init() {
        // Check if token exists in keychain
        if KeychainHelper.getToken() != nil {
            isAuthenticated = true
        }
    }

    func login(username: String, password: String) async throws {
        let (token, user) = try await APIService.shared.login(username: username, password: password)

        KeychainHelper.saveToken(token)

        await MainActor.run {
            self.currentUser = user
            self.isAuthenticated = true
        }

        // Register device if we have a token
        if let deviceToken = UserDefaults.standard.string(forKey: "deviceToken") {
            await APIService.shared.registerDevice(deviceToken: deviceToken)
        }
    }

    func register(username: String, email: String, password: String) async throws {
        let (token, user) = try await APIService.shared.register(username: username, email: email, password: password)

        KeychainHelper.saveToken(token)

        await MainActor.run {
            self.currentUser = user
            self.isAuthenticated = true
        }

        // Register device if we have a token
        if let deviceToken = UserDefaults.standard.string(forKey: "deviceToken") {
            await APIService.shared.registerDevice(deviceToken: deviceToken)
        }
    }

    func logout() {
        KeychainHelper.deleteToken()
        currentUser = nil
        isAuthenticated = false
    }
}
