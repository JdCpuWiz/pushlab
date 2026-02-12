import SwiftUI

struct SettingsView: View {
    @EnvironmentObject var authService: AuthService
    @State private var devices: [Device] = []
    @State private var deviceTags: [String] = UserDefaults.standard.stringArray(forKey: "deviceTags") ?? []
    @State private var newTag = ""
    @State private var showingAddTag = false

    var body: some View {
        NavigationView {
            Form {
                Section("Account") {
                    if let user = authService.currentUser {
                        HStack {
                            Text("Username")
                            Spacer()
                            Text(user.username)
                                .foregroundColor(.secondary)
                        }

                        HStack {
                            Text("Email")
                            Spacer()
                            Text(user.email)
                                .foregroundColor(.secondary)
                        }
                    }

                    Button("Logout", role: .destructive) {
                        authService.logout()
                    }
                }

                Section("Device Tags") {
                    ForEach(deviceTags, id: \.self) { tag in
                        HStack {
                            Text(tag)
                            Spacer()
                            Button(action: {
                                removeTag(tag)
                            }) {
                                Image(systemName: "trash")
                                    .foregroundColor(.red)
                            }
                        }
                    }

                    Button(action: {
                        showingAddTag = true
                    }) {
                        Label("Add Tag", systemImage: "plus.circle")
                    }
                }

                Section("Devices") {
                    ForEach(devices) { device in
                        VStack(alignment: .leading) {
                            Text(device.deviceName)
                                .font(.headline)
                            Text("Tags: \(device.tags.joined(separator: ", "))")
                                .font(.caption)
                                .foregroundColor(.secondary)
                        }
                    }
                }

                Section("About") {
                    HStack {
                        Text("Version")
                        Spacer()
                        Text("1.0.0")
                            .foregroundColor(.secondary)
                    }
                }
            }
            .navigationTitle("Settings")
            .alert("Add Tag", isPresented: $showingAddTag) {
                TextField("Tag name", text: $newTag)
                Button("Add") {
                    addTag()
                }
                Button("Cancel", role: .cancel) {}
            }
        }
        .onAppear(perform: loadDevices)
    }

    private func addTag() {
        guard !newTag.isEmpty else { return }
        deviceTags.append(newTag)
        UserDefaults.standard.set(deviceTags, forKey: "deviceTags")
        newTag = ""

        // Update device on backend
        if let deviceToken = UserDefaults.standard.string(forKey: "deviceToken") {
            Task {
                await APIService.shared.registerDevice(deviceToken: deviceToken)
            }
        }
    }

    private func removeTag(_ tag: String) {
        deviceTags.removeAll { $0 == tag }
        UserDefaults.standard.set(deviceTags, forKey: "deviceTags")

        // Update device on backend
        if let deviceToken = UserDefaults.standard.string(forKey: "deviceToken") {
            Task {
                await APIService.shared.registerDevice(deviceToken: deviceToken)
            }
        }
    }

    private func loadDevices() {
        Task {
            do {
                let fetched = try await APIService.shared.fetchDevices()
                await MainActor.run {
                    devices = fetched
                }
            } catch {
                print("Failed to load devices: \(error)")
            }
        }
    }
}
