import Foundation

struct PushNotification: Codable, Identifiable {
    let id: String
    let userId: String
    let title: String?
    let body: String
    let badge: Int?
    let sound: String
    let priority: String
    let tags: [String]?
    let createdAt: Date
    let status: String

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case title, body, badge, sound, priority, tags
        case createdAt = "created_at"
        case status
    }
}
