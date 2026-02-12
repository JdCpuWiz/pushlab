import SwiftUI

struct HomeView: View {
    @EnvironmentObject var authService: AuthService
    @State private var selectedTab = 0

    var body: some View {
        TabView(selection: $selectedTab) {
            NotificationHistoryView()
                .tabItem {
                    Label("Notifications", systemImage: "bell.fill")
                }
                .tag(0)

            SettingsView()
                .tabItem {
                    Label("Settings", systemImage: "gear")
                }
                .tag(1)
        }
    }
}

struct HomeView_Previews: PreviewProvider {
    static var previews: some View {
        HomeView()
            .environmentObject(AuthService.shared)
    }
}
