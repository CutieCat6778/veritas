import Apollo
import Foundation
import Graphql
import SwiftUI

struct HomePage<ViewModel: HomePageModelProtocol>: View {
    @ObservedObject var viewModel: ViewModel

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 32) {
                SectionView(sectionTitle: "homepage_now", articles: viewModel.articles)
            }
            .padding(.vertical)

            ForEach(viewModel.keywords) { item in
                SectionView(
                    sectionTitle: item.keyword,
                    articles: item.articles
                )
            }
        }
        .background(Color.clear)
        .onAppear {
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
