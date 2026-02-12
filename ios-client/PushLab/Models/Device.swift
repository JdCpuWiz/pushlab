import Foundation

struct Device: Codable, Identifiable {
    let id: String
    let userId: String
    let deviceName: String
    let deviceIdentifier: String
    let tags: [String]
    let createdAt: Date
    let updatedAt: Date
    let lastSeenAt: Date?

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case deviceName = "device_name"
        case deviceIdentifier = "device_identifier"
        case tags
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case lastSeenAt = "last_seen_at"
    }
}
