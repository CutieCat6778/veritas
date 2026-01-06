import Apollo
import Foundation
import Graphql
import SwiftUI

struct HomePage<ViewModel: HomePageModelProtocol>: View {
    @ObservedObject var viewModel: ViewModel

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 32) {
                // Recent Articles Section
                if viewModel.isLoadingArticles && viewModel.articles.isEmpty {
                    LoadingSectionView(sectionTitle: "homepage_now")
                } else {
                    SectionView(sectionTitle: "homepage_now", articles: viewModel.articles)
                }
            }
            .padding(.vertical)
            .padding(.bottom, 20)

            // Keywords Sections
            if viewModel.isLoadingKeywords && viewModel.keywords.isEmpty {
                LoadingKeywordsSections()
            } else {
                ForEach(viewModel.keywords) { item in
                    SectionView(
                        sectionTitle: item.keyword,
                        articles: item.articles
                    )
                    .padding(.bottom, 20)
                }
                .padding(.bottom, 20)
            }
        }
        .refreshable {
            await viewModel.refreshContent()
        }
        .background(Color.clear)
        .task {
            viewModel.getRecentArticles(amount: 5)
            viewModel.getKeyWords()
        }
    }
}

struct SectionView: View {
    let sectionTitle: String
    let articles: [Article]

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(LocalizedStringKey(sectionTitle))
                .font(.title2)
                .fontWeight(.bold)
                .multilineTextAlignment(.leading)
                .padding(.horizontal, 16)

            ScrollView(.horizontal, showsIndicators: false) {
                LazyHStack(spacing: 16) {
                    ForEach(articles) { article in
                        NewsCard(article: article)
                            .equatable()
                    }
                }
                .scrollTargetLayout()
                .padding(.horizontal, 16)
            }
            .scrollTargetBehavior(.viewAligned)
        }
        .background(Color.clear)
    }
}

// MARK: - Loading Views

struct LoadingSectionView: View {
    let sectionTitle: String

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text(LocalizedStringKey(sectionTitle))
                .font(.title2)
                .fontWeight(.bold)
                .multilineTextAlignment(.leading)
                .padding(.horizontal, 16)

            ScrollView(.horizontal, showsIndicators: false) {
                LazyHStack(spacing: 16) {
                    ForEach(0 ..< 3, id: \.self) { _ in
                        LoadingNewsCard()
                    }
                }
                .padding(.horizontal, 16)
            }
        }
        .background(Color.clear)
    }
}

struct LoadingKeywordsSections: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 32) {
            ForEach(0 ..< 2, id: \.self) { _ in
                VStack(alignment: .leading, spacing: 12) {
                    // Loading title placeholder
                    RoundedRectangle(cornerRadius: 8)
                        .fill(Color.gray.opacity(0.2))
                        .frame(width: 150, height: 28)
                        .padding(.horizontal, 16)
                        .shimmer()

                    ScrollView(.horizontal, showsIndicators: false) {
                        LazyHStack(spacing: 16) {
                            ForEach(0 ..< 3, id: \.self) { _ in
                                LoadingNewsCard()
                            }
                        }
                        .padding(.horizontal, 16)
                    }
                }
            }
        }
    }
}

struct LoadingNewsCard: View {
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            // Image placeholder
            RoundedRectangle(cornerRadius: 12)
                .fill(Color.gray.opacity(0.2))
                .frame(width: 280, height: 180)
                .shimmer()

            // Title placeholder
            VStack(alignment: .leading, spacing: 6) {
                RoundedRectangle(cornerRadius: 4)
                    .fill(Color.gray.opacity(0.2))
                    .frame(height: 16)
                    .shimmer()

                RoundedRectangle(cornerRadius: 4)
                    .fill(Color.gray.opacity(0.2))
                    .frame(width: 200, height: 16)
                    .shimmer()

                // Source and date placeholders
                HStack {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.gray.opacity(0.2))
                        .frame(width: 80, height: 12)
                        .shimmer()

                    Spacer()

                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.gray.opacity(0.2))
                        .frame(width: 60, height: 12)
                        .shimmer()
                }
            }
            .padding(.horizontal, 8)
        }
        .frame(width: 280)
    }
}

// MARK: - Shimmer Effect

extension View {
    func shimmer() -> some View {
        modifier(ShimmerEffect())
    }
}

struct ShimmerEffect: ViewModifier {
    @State private var phase: CGFloat = 0

    func body(content: Content) -> some View {
        content
            .overlay(
                LinearGradient(
                    gradient: Gradient(colors: [
                        Color.clear,
                        Color.white.opacity(0.3),
                        Color.clear,
                    ]),
                    startPoint: .leading,
                    endPoint: .trailing
                )
                .offset(x: phase)
                .mask(content)
            )
            .onAppear {
                withAnimation(
                    Animation.linear(duration: 1.5)
                        .repeatForever(autoreverses: false)
                ) {
                    phase = 400
                }
            }
    }
}
