import SwiftUI

struct NewsCard: View {
    let article: Article

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
        AsyncImage(url: URL(string: article.banner.isEmpty ? "https://picsum.photos/400/200" : article.banner)) { phase in
            switch phase {
            case .empty:
                Color.gray.opacity(0.3)
            case let .success(image):
                image
                    .resizable()
                    .scaledToFill()
            case .failure:
                Color.red.opacity(0.3)
            @unknown default:
                EmptyView()
            }
        }
        .frame(width: UIScreen.main.bounds.width * cardWidthRatio, height: (UIScreen.main.bounds.width * cardWidthRatio) * bannerHeightRatio)
        .clipped()
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
