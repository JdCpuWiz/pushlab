import Foundation

struct User: Codable, Identifiable {
    let id: String
    let username: String
    let email: String
    let apiKey: String?
    let createdAt: Date
    let isActive: Bool

    enum CodingKeys: String, CodingKey {
        case id, username, email
        case apiKey = "api_key"
        case createdAt = "created_at"
        case isActive = "is_active"
    }
}
