//
//  VeritasApp.swift
//  Veritas
//
//  Created by Thinh Nguyen on 04.01.26.
//

import SwiftUI

@main
struct VeritasApp: App {
    @AppStorage("hasLaunchedBefore") var hasLaunchedBefore: Bool = false

    var body: some Scene {
        WindowGroup {
            if hasLaunchedBefore {
                MainTabView()
            } else {
                WelcomeView()
            }
        }
    }
}
