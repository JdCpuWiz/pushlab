import SwiftUI

struct NotificationHistoryView: View {
    @State private var notifications: [PushNotification] = []
    @State private var isLoading = false

    var body: some View {
        NavigationView {
            Group {
                if isLoading {
                    ProgressView()
                } else if notifications.isEmpty {
                    VStack {
                        Image(systemName: "bell.slash")
                            .font(.system(size: 60))
                            .foregroundColor(.gray)
                        Text("No notifications yet")
                            .foregroundColor(.gray)
                            .padding()
                    }
                } else {
                    List(notifications) { notification in
                        NotificationRow(notification: notification)
                    }
                }
            }
            .navigationTitle("Notifications")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button(action: loadNotifications) {
                        Image(systemName: "arrow.clockwise")
                    }
                }
            }
        }
        .onAppear(perform: loadNotifications)
    }

    private func loadNotifications() {
        isLoading = true

        Task {
            do {
                let fetched = try await APIService.shared.fetchNotifications()
                await MainActor.run {
                    notifications = fetched
                    isLoading = false
                }
            } catch {
                await MainActor.run {
                    isLoading = false
                }
            }
        }
    }
}

struct NotificationRow: View {
    let notification: PushNotification

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            if let title = notification.title {
                Text(title)
                    .font(.headline)
            }

            Text(notification.body)
                .font(.body)
                .foregroundColor(.secondary)

            HStack {
                Text(notification.createdAt, style: .relative)
                    .font(.caption)
                    .foregroundColor(.gray)

                Spacer()

                StatusBadge(status: notification.status)
            }
        }
        .padding(.vertical, 4)
    }
}

struct StatusBadge: View {
    let status: String

    var body: some View {
        Text(status)
            .font(.caption)
            .padding(.horizontal, 8)
            .padding(.vertical, 4)
            .background(backgroundColor)
            .foregroundColor(.white)
            .cornerRadius(4)
    }

    private var backgroundColor: Color {
        switch status {
        case "delivered": return .green
        case "sent": return .blue
        case "failed": return .red
        default: return .gray
        }
    }
}
