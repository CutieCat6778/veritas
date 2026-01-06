import SwiftUI

struct PatternBackground: View {
    var body: some View {
        ZStack {
            // Base gradient
            LinearGradient(
                gradient: Gradient(colors: [
                    Color(.systemBackground),
                    Color(.systemBackground).opacity(0.95)
                ]),
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )

            // Pattern overlay
            GeometryReader { geometry in
                Canvas { context, size in
                    let spacing: CGFloat = 40
                    let circleSize: CGFloat = 3

                    for x in stride(from: 0, through: size.width, by: spacing) {
                        for y in stride(from: 0, through: size.height, by: spacing) {
                            let rect = CGRect(
                                x: x - circleSize / 2,
                                y: y - circleSize / 2,
                                width: circleSize,
                                height: circleSize
                            )
                            context.fill(
                                Path(ellipseIn: rect),
                                with: .color(Color.secondary.opacity(0.15))
                            )
                        }
                    }
                }
            }

            // Subtle gradient overlay for depth
            RadialGradient(
                gradient: Gradient(colors: [
                    Color.clear,
                    Color(.secondarySystemBackground).opacity(0.1)
                ]),
                center: .center,
                startRadius: 100,
                endRadius: 700
            )
        }
        .ignoresSafeArea()
    }
}

extension View {
    func withPatternBackground() -> some View {
        ZStack {
            PatternBackground()
            self
        }
    }
}
