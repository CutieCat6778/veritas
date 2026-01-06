import Foundation

struct Article: Identifiable, Codable, Hashable {
    let id: String
    let title: String
    let source: String
    let publishedAt: Date
    let uri: String
    let views: Int
    let description: String
    let banner: String
    let linkedTo: [Article]
    let category: [String]

    func hash(into hasher: inout Hasher) {
        hasher.combine(id)
    }

    static func == (lhs: Article, rhs: Article) -> Bool {
        lhs.id == rhs.id
    }
}

struct Keyword: Identifiable, Codable, Hashable {
    let id: String
    let keyword: String
    let lastUpdate: Date
    let articles: [Article]

    func hash(into hasher: inout Hasher) {
        hasher.combine(id)
    }

    static func == (lhs: Keyword, rhs: Keyword) -> Bool {
        lhs.id == rhs.id
    }
}

enum ArticleSource: String, Codable, CaseIterable {
    case tagesschau = "Tagesschau"
    case sueddeutsche = "Sueddeutsche"
    case dieZeit = "DieZeit"
    case faz = "FAZ"
    case welt = "Welt"
    case taz = "TAZ"
    case handelsblatt = "Handelsblatt"
}

enum Language: String, Codable {
    case english = "EN"
    case german = "DE"
    case unknown = "UNKNOWN"
}
