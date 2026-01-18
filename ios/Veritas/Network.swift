import Apollo
import Foundation

class Network {
    static let shared = Network()

    let apollo: ApolloClient

    init() {
        apollo = ApolloClient(url: URL(string: "https://veritas-server.thinis.de/query")!)
    }
}
