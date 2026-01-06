import Foundation
import SwiftUI
import Combine

@MainActor
class SettingsManager: ObservableObject {
    static let shared = SettingsManager()

    @Published var selectedLanguage: Language = .german
    @Published var selectedSources: Set<ArticleSource> = Set(ArticleSource.allCases)
    @Published var colorScheme: ColorScheme? = nil

    private let languageKey = "selectedLanguage"
    private let sourcesKey = "selectedSources"
    private let colorSchemeKey = "colorScheme"

    private var cancellables = Set<AnyCancellable>()

    private init() {
        loadFromDisk()
        setupObservers()
    }

    private func loadFromDisk() {
        if let savedLanguageRaw = UserDefaults.standard.string(forKey: languageKey),
           let savedLanguage = Language(rawValue: savedLanguageRaw) {
            self.selectedLanguage = savedLanguage
        }

        if let savedSourcesData = UserDefaults.standard.data(forKey: sourcesKey),
           let savedSourcesRaw = try? JSONDecoder().decode([String].self, from: savedSourcesData) {
            self.selectedSources = Set(savedSourcesRaw.compactMap { ArticleSource(rawValue: $0) })
        }

        if let savedSchemeRaw = UserDefaults.standard.string(forKey: colorSchemeKey) {
            switch savedSchemeRaw {
            case "light":
                self.colorScheme = .light
            case "dark":
                self.colorScheme = .dark
            default:
                self.colorScheme = nil
            }
        }
    }

    private func setupObservers() {
        $selectedLanguage
            .dropFirst()
            .sink { [weak self] _ in
                self?.saveToDisk()
            }
            .store(in: &cancellables)

        $selectedSources
            .dropFirst()
            .sink { [weak self] _ in
                self?.saveToDisk()
            }
            .store(in: &cancellables)

        $colorScheme
            .dropFirst()
            .sink { [weak self] _ in
                self?.saveToDisk()
            }
            .store(in: &cancellables)
    }

    private func saveToDisk() {
        UserDefaults.standard.set(selectedLanguage.rawValue, forKey: languageKey)

        let sourcesRaw = selectedSources.map { $0.rawValue }
        if let encoded = try? JSONEncoder().encode(sourcesRaw) {
            UserDefaults.standard.set(encoded, forKey: sourcesKey)
        }

        if let scheme = colorScheme {
            let schemeString = scheme == .light ? "light" : "dark"
            UserDefaults.standard.set(schemeString, forKey: colorSchemeKey)
        } else {
            UserDefaults.standard.set("system", forKey: colorSchemeKey)
        }
    }

    func resetToDefaults() {
        selectedLanguage = .german
        selectedSources = Set(ArticleSource.allCases)
        colorScheme = nil
    }
}
