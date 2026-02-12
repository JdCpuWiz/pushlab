import SwiftUI

@main
struct PushLabApp: App {
    @UIApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    @StateObject private var authService = AuthService.shared

    var body: some Scene {
        WindowGroup {
            if authService.isAuthenticated {
                HomeView()
                    .environmentObject(authService)
            } else {
                LoginView()
                    .environmentObject(authService)
            }
        }
    }
}
