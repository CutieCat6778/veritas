import Apollo
import Combine
import Foundation
import Graphql

@MainActor
final class SavedArticlesViewModel: ObservableObject {
    @Published var articles: [Article] = []
    @Published var isLoading: Bool = false
    @Published var hasError: Bool = false
    @Published var errorMessage: String = ""

    private var articlesTask: Task<Void, Never>?

    private let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ssZ"
        formatter.timeZone = TimeZone(secondsFromGMT: 0)
        return formatter
    }()

    private let iso8601Formatter = ISO8601DateFormatter()

    deinit {
        articlesTask?.cancel()
    }

    func fetchSavedArticles(ids: [String]) {
        guard !ids.isEmpty else {
            articles = []
            return
        }

        articlesTask?.cancel()
        isLoading = true
        hasError = false

        articlesTask = Task {
            do {
                let query = BatchFindArticlesQuery(ids: ids)
                let result = try await Network.shared.apollo.fetch(query: query)

                guard !Task.isCancelled else { return }

                if let fetchedData = result.data?.batchFindArticles {
                    self.articles = fetchedData.compactMap { self.mapToArticle(item: $0) }
                }
                self.isLoading = false
            } catch {
                guard !Task.isCancelled else { return }
                print("Error fetching saved articles: \(error)")
                self.hasError = true
                self.errorMessage = error.localizedDescription
                self.isLoading = false
            }
        }
    }

    private func mapToArticle(item: BatchFindArticlesQuery.Data.BatchFindArticle?) -> Article? {
        guard let item = item else { return nil }
        let fields = item.fragments.articleFields
        return mapArticleMinimal(fields: fields)
    }

    private func mapArticleMinimal(fields: ArticleFields) -> Article {
        return Article(
            id: fields.id,
            title: fields.title,
            source: fields.source.rawValue,
            publishedAt: parseDateSafe(fields.publishedAt),
            uri: fields.uri,
            views: fields.views,
            description: fields.description,
            banner: fields.banner,
            linkedTo: [],
            category: fields.category?.compactMap { $0 } ?? []
        )
    }

    private func parseDateSafe(_ dateString: String) -> Date {
        if let date = dateFormatter.date(from: dateString) {
            return date
        }
        if let date = iso8601Formatter.date(from: dateString) {
            return date
        }
        print("Warning: Could not parse date string: '\(dateString)', using current date")
        return Date()
    }
}
