import Apollo
import Foundation

class Network {
    static let shared = Network()

    let apollo: ApolloClient

    init() {
        apollo = ApolloClient(url: URL(string: "http://192.168.178.135:3000/query")!)
    }
}
