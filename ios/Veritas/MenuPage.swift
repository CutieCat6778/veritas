import SwiftUI

struct MenuPage: View {
    @StateObject private var savedArticlesManager = SavedArticlesManager.shared
    @State private var showingAbout = false

    var body: some View {
        NavigationStack {
            List {
                Section {
                    NavigationLink(destination: SavedArticlesView()) {
                        HStack {
                            Image(systemName: "bookmark.fill")
                                .foregroundColor(.blue)
                                .frame(width: 30)
                            Text("menu_saved_articles")
                            Spacer()
                            Text("\(savedArticlesManager.savedArticleIDs.count)")
                                .foregroundColor(.secondary)
                                .font(.subheadline)
                        }
                    }

                    NavigationLink(destination: SettingsView()) {
                        HStack {
                            Image(systemName: "gearshape.fill")
                                .foregroundColor(.blue)
                                .frame(width: 30)
                            Text("menu_settings")
                            Spacer()
                            Image(systemName: "chevron.right")
                                .foregroundColor(.secondary)
                                .font(.caption)
                        }
                    }
                }

                Section {
                    Button(action: {
                        showingAbout = true
                    }) {
                        HStack {
                            Image(systemName: "info.circle")
                                .foregroundColor(.blue)
                                .frame(width: 30)
                            Text("menu_about")
                            Spacer()
                            Image(systemName: "chevron.right")
                                .foregroundColor(.secondary)
                                .font(.caption)
                        }
                    }
                    .foregroundColor(.primary)
                }

                Section {
                    HStack {
                        Text("menu_version")
                            .foregroundColor(.secondary)
                        Spacer()
                        Text("1.0.0")
                            .foregroundColor(.secondary)
                    }
                    .font(.caption)
                }
            }
            .navigationTitle("menu_title")
            .sheet(isPresented: $showingAbout) {
                AboutView()
            }
        }
    }
}

struct SavedArticlesView: View {
    @StateObject private var savedArticlesManager = SavedArticlesManager.shared
    @StateObject private var viewModel = SavedArticlesViewModel()

    var body: some View {
        ScrollView {
            VStack {
                if viewModel.isLoading {
                    ProgressView()
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                        .padding(.top, 100)
                } else if savedArticlesManager.savedArticleIDs.isEmpty {
                    VStack(spacing: 16) {
                        Image(systemName: "bookmark.slash")
                            .font(.system(size: 60))
                            .foregroundColor(.secondary)
                        Text("menu_no_saved_articles")
                            .font(.headline)
                            .foregroundColor(.secondary)
                    }
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .padding(.top, 100)
                } else if viewModel.hasError {
                    VStack(spacing: 16) {
                        Image(systemName: "exclamationmark.triangle")
                            .font(.system(size: 60))
                            .foregroundColor(.red)
                        Text("Error loading articles")
                            .font(.headline)
                        Text(viewModel.errorMessage)
                            .font(.caption)
                            .foregroundColor(.secondary)
                            .multilineTextAlignment(.center)
                            .padding(.horizontal)
                    }
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .padding(.top, 100)
                } else {
                    LazyVStack(spacing: 16) {
                        ForEach(viewModel.articles) { article in
                            NewsCard(article: article)
                        }
                    }
                    .padding(.vertical)
                }
            }
        }
        .navigationTitle("menu_saved_articles")
        .navigationBarTitleDisplayMode(.inline)
        .onAppear {
            viewModel.fetchSavedArticles(ids: savedArticlesManager.getSavedArticleIDs())
        }
        .onChange(of: savedArticlesManager.savedArticleIDs) { _, _ in
            viewModel.fetchSavedArticles(ids: savedArticlesManager.getSavedArticleIDs())
        }
    }
}

struct AboutView: View {
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 24) {
                    Image(systemName: "newspaper.fill")
                        .font(.system(size: 80))
                        .foregroundColor(.blue)
                        .padding(.top, 40)

                    VStack(spacing: 8) {
                        Text("Veritas")
                            .font(.title)
                            .fontWeight(.bold)
                        Text("menu_version_number")
                            .font(.subheadline)
                            .foregroundColor(.secondary)
                    }

                    VStack(alignment: .leading, spacing: 16) {
                        Text("menu_about_description")
                            .font(.body)
                            .foregroundColor(.primary)
                            .multilineTextAlignment(.center)
                    }
                    .padding(.horizontal)

                    Spacer()
                }
                .padding()
            }
            .navigationTitle("menu_about")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }
}
