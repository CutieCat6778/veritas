import SwiftUI

struct ImagePlaceholder: View {
    let urlString: String
    let width: CGFloat
    let height: CGFloat

    var body: some View {
        AsyncImage(url: validURL) { phase in
            switch phase {
            case .empty:
                placeholderView
            // .overlay(ProgressView())
            case let .success(image):
                image
                    .resizable()
                    .scaledToFill()
            case .failure:
                placeholderView
            @unknown default:
                placeholderView
            }
        }
        .frame(width: width, height: height)
        .clipped()
    }

    private var validURL: URL? {
        guard !urlString.isEmpty else { return nil }
        return URL(string: urlString)
    }

    private var placeholderView: some View {
        ZStack {
            Color.gray.opacity(0.2)

            VStack(spacing: 8) {
                Image(systemName: "photo")
                    .font(.system(size: min(width, height) * 0.2))
                    .foregroundColor(.gray.opacity(0.5))
            }
        }
    }
}
