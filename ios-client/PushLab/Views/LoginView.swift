import SwiftUI

struct LoginView: View {
    @EnvironmentObject var authService: AuthService
    @State private var username = ""
    @State private var password = ""
    @State private var isLoading = false
    @State private var errorMessage: String?
    @State private var showingRegister = false

    var body: some View {
        NavigationView {
            VStack(spacing: 20) {
                Image(systemName: "bell.badge.fill")
                    .font(.system(size: 60))
                    .foregroundColor(.blue)
                    .padding(.bottom, 20)

                Text("PushLab")
                    .font(.largeTitle)
                    .bold()

                TextField("Username", text: $username)
                    .textFieldStyle(.roundedBorder)
                    .autocapitalization(.none)
                    .padding(.horizontal)

                SecureField("Password", text: $password)
                    .textFieldStyle(.roundedBorder)
                    .padding(.horizontal)

                if let error = errorMessage {
                    Text(error)
                        .foregroundColor(.red)
                        .font(.caption)
                }

                Button(action: login) {
                    if isLoading {
                        ProgressView()
                            .progressViewStyle(CircularProgressViewStyle(tint: .white))
                    } else {
                        Text("Login")
                            .bold()
                    }
                }
                .frame(maxWidth: .infinity)
                .padding()
                .background(Color.blue)
                .foregroundColor(.white)
                .cornerRadius(10)
                .padding(.horizontal)
                .disabled(isLoading || username.isEmpty || password.isEmpty)

                Button("Create Account") {
                    showingRegister = true
                }
                .padding()

                Spacer()
            }
            .padding()
            .navigationTitle("Welcome")
            .sheet(isPresented: $showingRegister) {
                RegisterView()
            }
        }
    }

    private func login() {
        isLoading = true
        errorMessage = nil

        Task {
            do {
                try await authService.login(username: username, password: password)
            } catch {
                await MainActor.run {
                    errorMessage = "Login failed. Please check your credentials."
                }
            }

            await MainActor.run {
                isLoading = false
            }
        }
    }
}

struct LoginView_Previews: PreviewProvider {
    static var previews: some View {
        LoginView()
            .environmentObject(AuthService.shared)
    }
}
