import Apollo
import SwiftUI

struct MainTabView: View {
    @StateObject private var settingsManager = SettingsManager.shared

    var body: some View {
        TabView {
            NavigationStack {
                HomePage(viewModel: HomePageModel())
            }
            .tabItem {
                Label("home_tab", systemImage: "house")
            }

            NavigationStack {
                DiscoverPage(viewModel: DiscoverPageModel())
            }
            .tabItem {
                Label("discover_tab", systemImage: "magnifyingglass")
            }

            MenuPage()
                .tabItem {
                    Label("menu_tab", systemImage: "line.3.horizontal")
                }
        }
        .withPatternBackground()
        .preferredColorScheme(settingsManager.colorScheme)
    }
}
