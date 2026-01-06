import SwiftUI

struct NewsCard: View, Equatable {
    let article: Article

    static func == (lhs: NewsCard, rhs: NewsCard) -> Bool {
        lhs.article.id == rhs.article.id
    }

    private let cardHeight: CGFloat = 330
    private let cornerRadius: CGFloat = 16
    private let bannerHeightRatio: CGFloat = 0.5
    private let contentPadding: CGFloat = 12
    private let borderOpacity: Double = 0.6
    private let shadowOpacity: Double = 0.1
    private let shadowRadius: CGFloat = 20
    private let cardWidthRatio: CGFloat = 0.9

    var body: some View {
        NavigationLink(destination: ArticleDetailView(article: article, pullData: true)) {
            VStack(spacing: 0) {
                bannerImage

                VStack(alignment: .leading, spacing: 4) {
                    titleText
                    descriptionText
                }
                .padding(contentPadding)
                .frame(maxWidth: .infinity, alignment: .leading)

                Spacer()

                sourceAndDate
                    .padding([.horizontal, .bottom], contentPadding)
            }
            .frame(width: UIScreen.main.bounds.width * cardWidthRatio, height: cardHeight)
            .glassEffect(in: .rect(cornerRadius: 16))
            .clipShape(RoundedRectangle(cornerRadius: cornerRadius))
        }
        .buttonStyle(PlainButtonStyle())
    }

    private var bannerImage: some View {
        ImagePlaceholder(
            urlString: article.banner,
            width: UIScreen.main.bounds.width * cardWidthRatio,
            height: (UIScreen.main.bounds.width * cardWidthRatio) * bannerHeightRatio
        )
    }

    private var titleText: some View {
        Text(article.title)
            .font(.headline)
            .foregroundColor(.primary)
            .lineLimit(2)
    }

    private var descriptionText: some View {
        Text(article.description)
            .font(.subheadline)
            .foregroundColor(.secondary)
            .lineLimit(3)
    }

    private var sourceAndDate: some View {
        HStack {
            Text(article.source)
                .font(.caption)
                .foregroundColor(.secondary)

            Spacer()

            HStack(spacing: 4) {
                Text(article.publishedAt, style: .date)
                    .font(.caption)
                    .foregroundColor(.secondary)
                Text("•")
                    .font(.caption)
                    .foregroundColor(.secondary)
                Text(article.publishedAt, style: .time)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            .lineLimit(1)
        }
    }
}
