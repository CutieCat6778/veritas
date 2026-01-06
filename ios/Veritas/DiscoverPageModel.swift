import Apollo
import Combine
import Foundation
import Graphql

@MainActor
protocol DiscoverPageModelProtocol: ObservableObject, AnyObject {
    var articles: [Article] { get }
    var filteredArticles: [Article] { get }
    var searchText: String { get set }
    var isLoading: Bool { get }
    var hasError: Bool { get }
    var errorMessage: String { get }
    func getAllArticles()
    func refreshArticles() async
}

final class DiscoverPageModel: DiscoverPageModelProtocol {
    @Published var articles: [Article] = []
    @Published var searchText: String = ""
    @Published var isLoading: Bool = false
    @Published var hasError: Bool = false
    @Published var errorMessage: String = ""

    private var hasFetchedArticles: Bool = false

    var filteredArticles: [Article] {
        guard !searchText.isEmpty else {
            return articles
        }

        let trimmedSearch = searchText.trimmingCharacters(in: .whitespaces)
        guard !trimmedSearch.isEmpty else {
            return articles
        }

        return articles.filter { article in
            article.title.localizedCaseInsensitiveContains(trimmedSearch) ||
            article.description.localizedCaseInsensitiveContains(trimmedSearch) ||
            article.source.localizedCaseInsensitiveContains(trimmedSearch) ||
            article.category.contains { $0.localizedCaseInsensitiveContains(trimmedSearch) }
        }
    }

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

    func getAllArticles() {
        // Only fetch if we haven't already
        guard !hasFetchedArticles && !isLoading else { return }

        articlesTask?.cancel()
        isLoading = true
        hasError = false

        articlesTask = Task {
            do {
                let query = GetArticlesQuery()
                let result = try await Network.shared.apollo.fetch(query: query)

                guard !Task.isCancelled else { return }

                if let fetchedData = result.data?.articles {
                    self.articles = fetchedData
                        .compactMap { self.mapToArticle(item: $0) }
                        .sorted { $0.publishedAt > $1.publishedAt } // Sort by most recent
                    self.hasFetchedArticles = true
                }

                self.isLoading = false
            } catch {
                guard !Task.isCancelled else { return }
                print("Error fetching articles: \(error)")
                self.hasError = true
                self.errorMessage = error.localizedDescription
                self.isLoading = false
            }
        }
    }

    func refreshArticles() async {
        articlesTask?.cancel()
        hasError = false
        hasFetchedArticles = false

        articlesTask = Task {
            do {
                let query = GetArticlesQuery()
                let result = try await Network.shared.apollo.fetch(query: query)

                guard !Task.isCancelled else { return }

                if let fetchedData = result.data?.articles {
                    self.articles = fetchedData
                        .compactMap { self.mapToArticle(item: $0) }
                        .sorted { $0.publishedAt > $1.publishedAt }
                    self.hasFetchedArticles = true
                    self.hasError = false
                }
            } catch {
                guard !Task.isCancelled else { return }
                print("Error refreshing articles: \(error)")
                self.hasError = true
                self.errorMessage = error.localizedDescription
            }
        }

        await articlesTask?.value
    }

    private func mapToArticle(item: GetArticlesQuery.Data.Article?) -> Article? {
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
