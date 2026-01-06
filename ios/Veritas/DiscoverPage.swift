import Apollo
import Foundation
import Graphql
import SwiftUI

struct DiscoverPage<ViewModel: DiscoverPageModelProtocol>: View {
    @ObservedObject var viewModel: ViewModel

    var body: some View {
        VStack(spacing: 0) { // Search Bar
            SearchBar(text: $viewModel.searchText)
                .padding(.horizontal)
                .padding(.vertical, 8)

            // Articles List
            if viewModel.isLoading && viewModel.articles.isEmpty {
                Spacer()
                ProgressView()
                    .scaleEffect(1.5)
                Spacer()
            } else if viewModel.hasError && viewModel.articles.isEmpty {
                Spacer()
                VStack(spacing: 16) {
                    Image(systemName: "exclamationmark.triangle")
                        .font(.system(size: 48))
                        .foregroundColor(.red)
                    Text("Error loading articles")
                        .font(.headline)
                        .foregroundColor(.primary)
                    Text(viewModel.errorMessage)
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                        .multilineTextAlignment(.center)
                        .padding(.horizontal)
                    Button("Retry") {
                        viewModel.getAllArticles()
                    }
                    .buttonStyle(.bordered)
                }
                Spacer()
            } else if viewModel.filteredArticles.isEmpty {
                Spacer()
                VStack(spacing: 12) {
                    Image(systemName: "magnifyingglass")
                        .font(.system(size: 48))
                        .foregroundColor(.secondary)
                    Text(viewModel.searchText.isEmpty ? "No articles found" : "No search results")
                        .font(.headline)
                        .foregroundColor(.secondary)
                    if !viewModel.searchText.isEmpty {
                        Text("Try searching with different keywords")
                            .font(.subheadline)
                            .foregroundColor(.secondary)
                    }
                }
                Spacer()
            } else {
                ScrollView {
                    LazyVStack(spacing: 12) {
                        ForEach(Array(viewModel.filteredArticles.enumerated()), id: \.element.id) { index, article in
                            CompactArticleCard(article: article)
                                .equatable()
                                .onAppear {
                                    // Load more when approaching the end (3 items before the end)
                                    if index == viewModel.filteredArticles.count - 3 {
                                        viewModel.loadMoreArticles()
                                    }
                                }
                        }

                        // Loading indicator for pagination
                        if viewModel.isLoadingMore {
                            HStack {
                                Spacer()
                                ProgressView()
                                    .padding()
                                Spacer()
                            }
                        }
                    }
                    .padding(.horizontal)
                    .padding(.top, 8)
                }
                .refreshable {
                    await viewModel.refreshArticles()
                }
            }
        }
        .navigationTitle("discover_tab")
        .navigationBarTitleDisplayMode(.large)
        .task {
            viewModel.getAllArticles()
        }
    }
}

// MARK: - Search Bar Component

struct SearchBar: View {
    @Binding var text: String
    @FocusState private var isFocused: Bool

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: "magnifyingglass")
                .foregroundColor(.secondary)
                .imageScale(.medium)

            TextField(LocalizedStringKey("discover_searchbar_placeholder"), text: $text)
                .textFieldStyle(PlainTextFieldStyle())
                .autocorrectionDisabled()
                .textInputAutocapitalization(.never)
                .focused($isFocused)

            if !text.isEmpty {
                Button(action: {
                    text = ""
                    isFocused = false
                }) {
                    Image(systemName: "xmark.circle.fill")
                        .foregroundColor(.secondary)
                        .imageScale(.medium)
                }
                .transition(.scale.combined(with: .opacity))
            }
        }
        .padding(12)
        .background(Color(.systemGray6))
        .cornerRadius(12)
        .animation(.easeInOut(duration: 0.2), value: text.isEmpty)
    }
}

// MARK: - Compact Article Card Component

struct CompactArticleCard: View, Equatable {
    let article: Article

    static func == (lhs: CompactArticleCard, rhs: CompactArticleCard) -> Bool {
        lhs.article.id == rhs.article.id
    }

    private let cardHeight: CGFloat = 140
    private let cornerRadius: CGFloat = 12
    private let imageWidth: CGFloat = 120

    var body: some View {
        NavigationLink(destination: ArticleDetailView(article: article, pullData: true)) {
            HStack(spacing: 12) {
                // Banner Image
                ImagePlaceholder(
                    urlString: article.banner.isEmpty ? "https://picsum.photos/400/300" : article.banner,
                    width: imageWidth,
                    height: cardHeight
                )
                .clipShape(RoundedRectangle(cornerRadius: cornerRadius))

                // Content
                VStack(alignment: .leading, spacing: 6) {
                    Text(article.title)
                        .font(.subheadline)
                        .fontWeight(.semibold)
                        .foregroundColor(.primary)
                        .lineLimit(3)
                        .multilineTextAlignment(.leading)

                    Spacer()

                    Text(article.description)
                        .font(.caption)
                        .foregroundColor(.secondary)
                        .lineLimit(2)

                    Spacer()

                    // Source and Date
                    HStack {
                        Text(article.source)
                            .font(.caption2)
                            .foregroundColor(.secondary)

                        Spacer()

                        Text(article.publishedAt, style: .date)
                            .font(.caption2)
                            .foregroundColor(.secondary)
                    }
                }
                .frame(maxHeight: cardHeight)
                .padding(.vertical, 4)
            }
            .frame(height: cardHeight)
            .padding(12)
            .background(Color(.systemBackground))
            .glassEffect(in: .rect(cornerRadius: cornerRadius))
            .clipShape(RoundedRectangle(cornerRadius: cornerRadius))
        }
        .buttonStyle(PlainButtonStyle())
    }
}

// MARK: - Preview with Mock ViewModel

extension DiscoverPage where ViewModel == DiscoverPageModel {
    init() {
        self.init(viewModel: DiscoverPageModel())
    }
}
