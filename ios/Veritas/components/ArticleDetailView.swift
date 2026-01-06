import Apollo
import SwiftUI

struct ArticleDetailView: View {
    let article: Article
    let pullData: Bool
    @Environment(\.dismiss) private var dismiss
    @StateObject private var viewModel = ArticleDetailViewModel()
    @StateObject private var savedArticlesManager = SavedArticlesManager.shared

    private let bannerHeight: CGFloat = 250

    init(article: Article, pullData: Bool = false) {
        self.article = article
        self.pullData = pullData
    }

    var body: some View {
        ZStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    // Banner Image
                    ImagePlaceholder(
                        urlString: displayArticle.banner,
                        width: UIScreen.main.bounds.width,
                        height: bannerHeight
                    )

                VStack(alignment: .leading, spacing: 16) {
                    // Categories
                    if !displayArticle.category.isEmpty {
                        ScrollView(.horizontal, showsIndicators: false) {
                            HStack(spacing: 8) {
                                ForEach(displayArticle.category, id: \.self) { category in
                                    Text(category)
                                        .font(.caption)
                                        .padding(.horizontal, 12)
                                        .padding(.vertical, 6)
                                        .background(Color.blue.opacity(0.2))
                                        .foregroundColor(.blue)
                                        .clipShape(Capsule())
                                }
                            }
                        }
                    }

                    // Title
                    Text(displayArticle.title)
                        .font(.title)
                        .fontWeight(.bold)
                        .fixedSize(horizontal: false, vertical: true)
                        .multilineTextAlignment(.leading)

                    // Source and Date
                    HStack {
                        Text(displayArticle.source)
                            .font(.subheadline)
                            .foregroundColor(.secondary)
                            .lineLimit(1)
                            .truncationMode(.tail)

                        Spacer()

                        HStack(spacing: 4) {
                            Text(displayArticle.publishedAt, style: .date)
                                .font(.subheadline)
                                .foregroundColor(.secondary)
                            Text("•")
                                .foregroundColor(.secondary)
                            Text(displayArticle.publishedAt, style: .time)
                                .font(.subheadline)
                                .foregroundColor(.secondary)
                        }
                        .lineLimit(1)
                    }

                    // Views
                    HStack(spacing: 4) {
                        Image(systemName: "eye.fill")
                            .font(.caption)
                        Text("\(displayArticle.views) views")
                            .font(.caption)
                    }
                    .foregroundColor(.secondary)

                    Divider()

                    // Description
                    Text(displayArticle.description)
                        .font(.body)
                        .foregroundColor(.primary)
                        .fixedSize(horizontal: false, vertical: true)
                        .multilineTextAlignment(.leading)

                    // Read Full Article Button
                    if let url = URL(string: displayArticle.uri) {
                        Link(destination: url) {
                            HStack {
                                Text("articledetailview_read_full_article")
                                    .fontWeight(.semibold)
                                Spacer()
                                Image(systemName: "arrow.up.right")
                            }
                            .padding()
                            .frame(maxWidth: .infinity)
                            .background(Color.blue)
                            .foregroundColor(.white)
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                        }
                    }

                    // Linked Articles Section
                    if !displayArticle.linkedTo.isEmpty {
                        VStack(alignment: .leading, spacing: 12) {
                            Text("articledetailview_related_articles")
                                .font(.title2)
                                .fontWeight(.bold)

                            ForEach(displayArticle.linkedTo) { linkedArticle in
                                NavigationLink(destination: ArticleDetailView(article: linkedArticle, pullData: true)) {
                                    LinkedArticleCard(article: linkedArticle)
                                }
                                .buttonStyle(PlainButtonStyle())
                            }
                        }
                        .padding(.top, 8)
                    }
                }
                .padding(.horizontal)
            }
        }
        .opacity(viewModel.isLoading && pullData ? 0.5 : 1.0)

        // Loading Overlay
        if viewModel.isLoading && pullData {
            VStack(spacing: 16) {
                ProgressView()
                    .scaleEffect(1.5)
                    .tint(.blue)
                Text("Loading article details...")
                    .font(.subheadline)
                    .foregroundColor(.secondary)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(Color(.systemBackground).opacity(0.9))
        }
        }
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                Button(action: {
                    withAnimation(.spring(response: 0.3, dampingFraction: 0.6)) {
                        savedArticlesManager.toggleSaved(displayArticle.id)
                    }
                }) {
                    Image(systemName: savedArticlesManager.isSaved(displayArticle.id) ? "bookmark.fill" : "bookmark")
                        .font(.title3)
                        .foregroundColor(savedArticlesManager.isSaved(displayArticle.id) ? .blue : .secondary)
                        .symbolEffect(.bounce, value: savedArticlesManager.isSaved(displayArticle.id))
                }
            }
        }
        .task {
            if pullData {
                viewModel.fetchArticle(id: article.id)
            }
        }
    }

    private var displayArticle: Article {
        viewModel.fetchedArticle ?? article
    }
}
