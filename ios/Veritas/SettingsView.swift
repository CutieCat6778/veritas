import SwiftUI

struct SettingsView: View {
    @StateObject private var settingsManager = SettingsManager.shared
    @State private var showingResetAlert = false

    var body: some View {
        List {
            Section {
                Picker("settings_language", selection: $settingsManager.selectedLanguage) {
                    Text("settings_german").tag(Language.german)
                    Text("settings_english").tag(Language.english)
                }
                .pickerStyle(.menu)
            } header: {
                Text("settings_language_section")
            } footer: {
                Text("settings_language_footer")
                    .font(.caption)
            }

            Section {
                ForEach(ArticleSource.allCases, id: \.self) { source in
                    Toggle(isOn: Binding(
                        get: { settingsManager.selectedSources.contains(source) },
                        set: { isSelected in
                            if isSelected {
                                settingsManager.selectedSources.insert(source)
                            } else {
                                if settingsManager.selectedSources.count > 1 {
                                    settingsManager.selectedSources.remove(source)
                                }
                            }
                        }
                    )) {
                        HStack {
                            Image(systemName: "newspaper")
                                .foregroundColor(.blue)
                                .frame(width: 25)
                            Text(source.rawValue)
                        }
                    }
                }
            } header: {
                Text("settings_sources_section")
            } footer: {
                Text("settings_sources_footer")
                    .font(.caption)
            }

            Section {
                Picker("settings_appearance", selection: $settingsManager.colorScheme) {
                    Text("settings_system").tag(nil as ColorScheme?)
                    Text("settings_light").tag(ColorScheme.light as ColorScheme?)
                    Text("settings_dark").tag(ColorScheme.dark as ColorScheme?)
                }
                .pickerStyle(.segmented)
            } header: {
                Text("settings_appearance_section")
            } footer: {
                Text("settings_appearance_footer")
                    .font(.caption)
            }

            Section {
                Button(role: .destructive, action: {
                    showingResetAlert = true
                }) {
                    HStack {
                        Spacer()
                        Text("settings_reset")
                        Spacer()
                    }
                }
            }
        }
        .navigationTitle("menu_settings")
        .navigationBarTitleDisplayMode(.inline)
        .alert("settings_reset_alert_title", isPresented: $showingResetAlert) {
            Button("settings_cancel", role: .cancel) {}
            Button("settings_reset", role: .destructive) {
                settingsManager.resetToDefaults()
            }
        } message: {
            Text("settings_reset_alert_message")
        }
    }
}
