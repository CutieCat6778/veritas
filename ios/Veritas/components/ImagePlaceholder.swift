import SwiftUI

struct ImagePlaceholder: View {
    let urlString: String
    let width: CGFloat
    let height: CGFloat

    @State private var isVisible: Bool = false

    var body: some View {
        GeometryReader { geometry in
            AsyncImage(url: isVisible ? validURL : nil) { phase in
                switch phase {
                case .empty:
                    placeholderView
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
            .onAppear {
                checkVisibility(geometry: geometry)
            }
            .onChange(of: geometry.frame(in: .global)) { _, _ in
                checkVisibility(geometry: geometry)
            }
        }
        .frame(width: width, height: height)
        .clipped()
    }

    private func checkVisibility(geometry: GeometryProxy) {
        let frame = geometry.frame(in: .global)
        let screenHeight = UIScreen.main.bounds.height
        let screenWidth = UIScreen.main.bounds.width

        // Check if the image is within the viewport (with some buffer for preloading)
        let buffer: CGFloat = 200
        let isInViewport = frame.minY < screenHeight + buffer &&
                          frame.maxY > -buffer &&
                          frame.minX < screenWidth + buffer &&
                          frame.maxX > -buffer

        if isInViewport && !isVisible {
            isVisible = true
        }
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
