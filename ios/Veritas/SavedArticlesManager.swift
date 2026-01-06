import Foundation
import Combine

@MainActor
class SavedArticlesManager: ObservableObject {
    static let shared = SavedArticlesManager()

    @Published private(set) var savedArticleIDs: Set<String> = []

    private let savedArticlesKey = "savedArticles"

    private init() {
        loadSavedArticles()
    }

    func isSaved(_ articleID: String) -> Bool {
        return savedArticleIDs.contains(articleID)
    }

    func toggleSaved(_ articleID: String) {
        if savedArticleIDs.contains(articleID) {
            savedArticleIDs.remove(articleID)
        } else {
            savedArticleIDs.insert(articleID)
        }
        saveToDisk()
    }

    func saveArticle(_ articleID: String) {
        savedArticleIDs.insert(articleID)
        saveToDisk()
    }

    func unsaveArticle(_ articleID: String) {
        savedArticleIDs.remove(articleID)
        saveToDisk()
    }

    private func loadSavedArticles() {
        if let data = UserDefaults.standard.data(forKey: savedArticlesKey),
           let decoded = try? JSONDecoder().decode(Set<String>.self, from: data)
        {
            savedArticleIDs = decoded
        }
    }

    private func saveToDisk() {
        if let encoded = try? JSONEncoder().encode(savedArticleIDs) {
            UserDefaults.standard.set(encoded, forKey: savedArticlesKey)
        }
    }

    func getSavedArticleIDs() -> [String] {
        return Array(savedArticleIDs)
    }

    func clearAllSaved() {
        savedArticleIDs.removeAll()
        saveToDisk()
    }
}
