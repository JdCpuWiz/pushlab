import Foundation

struct Constants {
    // Update these values for your setup
    static let apiBaseURL = "http://localhost:8080"
    static let bundleID = Bundle.main.bundleIdentifier ?? "com.pushlab.app"

    #if DEBUG
    static let apnsEnvironment = "sandbox"
    #else
    static let apnsEnvironment = "production"
    #endif
}
