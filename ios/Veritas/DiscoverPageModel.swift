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
    var isLoadingMore: Bool { get }
    var hasError: Bool { get }
    var errorMessage: String { get }
    func getAllArticles()
    func refreshArticles() async
    func loadMoreArticles()
}

final class DiscoverPageModel: DiscoverPageModelProtocol {
    @Published var articles: [Article] = []
    @Published var searchText: String = ""
    @Published var isLoading: Bool = false
    @Published var isLoadingMore: Bool = false
    @Published var hasError: Bool = false
    @Published var errorMessage: String = ""

    private var hasFetchedArticles: Bool = false
    private var cancellables = Set<AnyCancellable>()
    @Published private var debouncedSearchText: String = ""

    // Pagination state
    private var currentOffset: Int = 0
    private let pageSize: Int = 20
    private var hasMoreArticles: Bool = true

    var filteredArticles: [Article] {
        guard !debouncedSearchText.isEmpty else {
            return articles
        }

        let trimmedSearch = debouncedSearchText.trimmingCharacters(in: .whitespaces)
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

    init() {
        // Debounce search text changes by 0.3 seconds
        $searchText
            .debounce(for: .milliseconds(300), scheduler: DispatchQueue.main)
            .assign(to: &$debouncedSearchText)
    }

    deinit {
        articlesTask?.cancel()
    }

    func getAllArticles() {
        // Only fetch if we haven't already
        guard !hasFetchedArticles && !isLoading else { return }

        articlesTask?.cancel()
        isLoading = true
        hasError = false
        currentOffset = 0
        hasMoreArticles = true

        articlesTask = Task {
            do {
                let query = GetNextRecentArticlesQuery(start: 0, stop: Int32(pageSize))
                let result = try await Network.shared.apollo.fetch(query: query)

                guard !Task.isCancelled else { return }

                if let fetchedData = result.data?.nextRecentArticle {
                    self.articles = fetchedData
                        .compactMap { self.mapToArticleWithLinks(item: $0) }
                    self.currentOffset = self.articles.count
                    self.hasFetchedArticles = true
                    self.hasMoreArticles = fetchedData.count == self.pageSize
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

    func loadMoreArticles() {
        // Don't load more if already loading, no more articles, or user is searching
        guard !isLoadingMore && !isLoading && hasMoreArticles && debouncedSearchText.isEmpty else { return }

        isLoadingMore = true

        Task {
            do {
                let query = GetNextRecentArticlesQuery(start: Int32(currentOffset), stop: Int32(currentOffset + pageSize))
                let result = try await Network.shared.apollo.fetch(query: query)

                guard !Task.isCancelled else { return }

                if let fetchedData = result.data?.nextRecentArticle {
                    let newArticles = fetchedData.compactMap { self.mapToArticleWithLinks(item: $0) }
                    self.articles.append(contentsOf: newArticles)
                    self.currentOffset = self.articles.count
                    self.hasMoreArticles = newArticles.count == self.pageSize
                }

                self.isLoadingMore = false
            } catch {
                guard !Task.isCancelled else { return }
                print("Error loading more articles: \(error)")
                self.isLoadingMore = false
            }
        }
    }

    func refreshArticles() async {
        articlesTask?.cancel()
        hasError = false
        hasFetchedArticles = false
        currentOffset = 0
        hasMoreArticles = true

        articlesTask = Task {
            do {
                let query = GetNextRecentArticlesQuery(start: 0, stop: Int32(pageSize))
                let result = try await Network.shared.apollo.fetch(query: query)

                guard !Task.isCancelled else { return }

                if let fetchedData = result.data?.nextRecentArticle {
                    self.articles = fetchedData
                        .compactMap { self.mapToArticleWithLinks(item: $0) }
                    self.currentOffset = self.articles.count
                    self.hasFetchedArticles = true
                    self.hasError = false
                    self.hasMoreArticles = fetchedData.count == self.pageSize
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

    private func mapToArticleWithLinks(item: GetNextRecentArticlesQuery.Data.NextRecentArticle?) -> Article? {
        guard let item = item else { return nil }
        let fields = item.fragments.articleFields

        // Map linked articles
        let linkedArticles = item.fragments.articleWithLinks.linkedTo?.compactMap { linkedItem -> Article? in
            guard let linkedItem = linkedItem else { return nil }
            return Article(
                id: linkedItem.id,
                title: linkedItem.title,
                source: linkedItem.source.rawValue,
                publishedAt: parseDateSafe(linkedItem.publishedAt),
                uri: "",
                views: 0,
                description: "",
                banner: linkedItem.banner,
                linkedTo: [],
                category: []
            )
        } ?? []

        return Article(
            id: fields.id,
            title: fields.title,
            source: fields.source.rawValue,
            publishedAt: parseDateSafe(fields.publishedAt),
            uri: fields.uri,
            views: fields.views,
            description: fields.description,
            banner: fields.banner,
            linkedTo: linkedArticles,
            category: fields.category?.compactMap { $0 } ?? []
        )
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
