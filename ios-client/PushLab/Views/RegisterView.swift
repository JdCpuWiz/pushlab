import SwiftUI

struct RegisterView: View {
    @Environment(\.dismiss) var dismiss
    @EnvironmentObject var authService: AuthService

    @State private var username = ""
    @State private var email = ""
    @State private var password = ""
    @State private var confirmPassword = ""
    @State private var isLoading = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationView {
            Form {
                Section("Account Information") {
                    TextField("Username", text: $username)
                        .autocapitalization(.none)

                    TextField("Email", text: $email)
                        .autocapitalization(.none)
                        .keyboardType(.emailAddress)

                    SecureField("Password", text: $password)

                    SecureField("Confirm Password", text: $confirmPassword)
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundColor(.red)
                            .font(.caption)
                    }
                }

                Section {
                    Button(action: register) {
                        if isLoading {
                            HStack {
                                Spacer()
                                ProgressView()
                                Spacer()
                            }
                        } else {
                            Text("Create Account")
                                .frame(maxWidth: .infinity)
                        }
                    }
                    .disabled(isLoading || !isValid)
                }
            }
            .navigationTitle("Register")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
            }
        }
    }

    private var isValid: Bool {
        !username.isEmpty && !email.isEmpty && !password.isEmpty &&
        password == confirmPassword && password.count >= 8
    }

    private func register() {
        guard password == confirmPassword else {
            errorMessage = "Passwords do not match"
            return
        }

        guard password.count >= 8 else {
            errorMessage = "Password must be at least 8 characters"
            return
        }

        isLoading = true
        errorMessage = nil

        Task {
            do {
                try await authService.register(username: username, email: email, password: password)
                await MainActor.run {
                    dismiss()
                }
            } catch {
                await MainActor.run {
                    errorMessage = "Registration failed. Please try again."
                }
            }

            await MainActor.run {
                isLoading = false
            }
        }
    }
}
