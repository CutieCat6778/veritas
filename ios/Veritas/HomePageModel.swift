import Apollo
import Combine
import Foundation
import Graphql

@MainActor
protocol HomePageModelProtocol: ObservableObject, AnyObject {
    var articles: [Article] { get }
    var keywords: [Keyword] { get }
    var isLoadingArticles: Bool { get }
    var isLoadingKeywords: Bool { get }
    func getRecentArticles(amount: Int)
    func getKeyWords()
}

final class HomePageModel: HomePageModelProtocol {
    @Published var articles: [Article] = []
    @Published var keywords: [Keyword] = []
    @Published var isLoadingArticles: Bool = false
    @Published var isLoadingKeywords: Bool = false

    private var articlesTask: Task<Void, Never>?
    private var keywordsTask: Task<Void, Never>?

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
        keywordsTask?.cancel()
    }

    func getRecentArticles(amount: Int) {
        articlesTask?.cancel()
        isLoadingArticles = true
        articlesTask = Task {
            do {
                let query = GetRecentArticlesQuery(amount: Int32(amount))
                let result = try await Network.shared.apollo.fetch(query: query)

                guard !Task.isCancelled else { return }

                if let fetchedData = result.data?.recentArticle {
                    self.articles = fetchedData.compactMap { self.mapToArticle(item: $0) }
                }
                self.isLoadingArticles = false
            } catch {
                guard !Task.isCancelled else { return }
                print("Error fetching articles: \(error)")
                self.isLoadingArticles = false
            }
        }
    }

    func getKeyWords() {
        keywordsTask?.cancel()
        isLoadingKeywords = true
        keywordsTask = Task {
            do {
                let query = GetKeywordsQuery()
                let result = try await Network.shared.apollo.fetch(query: query)

                guard !Task.isCancelled else { return }

                if let fetchedData = result.data?.keywords {
                    self.keywords = fetchedData.compactMap { self.mapToKeyword(item: $0) }
                }
                self.isLoadingKeywords = false
            } catch {
                guard !Task.isCancelled else { return }
                print("Error fetching keywords: \(error)")
                self.isLoadingKeywords = false
            }
        }
    }

    private func mapToKeyword(item: GetKeywordsQuery.Data.Keyword?) -> Keyword? {
        guard let item = item else { return nil }

        let fields = item.fragments.responseKeywordFields
        let associatedArticles = fields.articles.compactMap { articleItem -> Article? in
            guard let articleItem = articleItem else { return nil }
            let aFields = articleItem.fragments.articleFields
            return mapArticleMinimal(fields: aFields)
        }

        return Keyword(
            id: fields.id,
            keyword: fields.keyword,
            lastUpdate: parseDateSafe(fields.lastUpdate),
            articles: associatedArticles
        )
    }

    private func mapToArticle(item: GetRecentArticlesQuery.Data.RecentArticle?) -> Article? {
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
