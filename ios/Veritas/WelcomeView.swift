//  WelcomeView.swift
//  Veritas
//
//  Created by Thinh Nguyen on 04.01.26.
//

import SwiftUI

struct WelcomeView: View {
    @AppStorage("hasLaunchedBefore") var hasLaunchedBefore: Bool = false

    var body: some View {
        ZStack {
            Color(UIColor.systemBackground)
                .ignoresSafeArea()

            VStack(alignment: .leading, spacing: 10) {
                Spacer()

                Text("Welcome to Veritas!")
                    .font(.largeTitle)
                    .fontWeight(.bold)
                    .foregroundColor(Color(UIColor.label))

                Text("Your journey with the truth of media, starts here.")
                    .foregroundColor(Color(UIColor.secondaryLabel))

                Spacer()

                HStack {
                    Spacer()
                    Button(action: {
                        hasLaunchedBefore = true
                    }) {
                        Text("Continue")
                            .foregroundColor(Color(UIColor.systemBackground))
                            .padding()
                            .background(Color(UIColor.label))
                            .cornerRadius(10)
                    }
                }
            }
        }
        .padding()
    }
}
