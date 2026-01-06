// @generated
// This file was automatically generated and should not be edited.

import ApolloTestSupport
@testable import Graphql

public final class Query: MockObject {
  public static let objectType: ApolloAPI.Object = Graphql.Objects.Query
  public static let _mockFields = MockFields()
  public typealias MockValueCollectionType = Array<Mock<Query>>

  public struct MockFields: Sendable {
    @Field<Article>("article") public var article
    @Field<[Article?]>("articles") public var articles
    @Field<[Article?]>("batchFindArticles") public var batchFindArticles
    @Field<[ResponseKeyWords?]>("keywords") public var keywords
    @Field<[Article?]>("linkedArticles") public var linkedArticles
    @Field<[Article?]>("nextRecentArticle") public var nextRecentArticle
    @Field<[Article?]>("recentArticle") public var recentArticle
    @Field<[Article?]>("topArticles") public var topArticles
  }
}

public extension Mock where O == Query {
  convenience init(
    article: Mock<Article>? = nil,
    articles: [Mock<Article>?] = [],
    batchFindArticles: [Mock<Article>?] = [],
    keywords: [Mock<ResponseKeyWords>?] = [],
    linkedArticles: [Mock<Article>?] = [],
    nextRecentArticle: [Mock<Article>?] = [],
    recentArticle: [Mock<Article>?] = [],
    topArticles: [Mock<Article>?] = []
  ) {
    self.init()
    _setEntity(article, for: \.article)
    _setList(articles, for: \.articles)
    _setList(batchFindArticles, for: \.batchFindArticles)
    _setList(keywords, for: \.keywords)
    _setList(linkedArticles, for: \.linkedArticles)
    _setList(nextRecentArticle, for: \.nextRecentArticle)
    _setList(recentArticle, for: \.recentArticle)
    _setList(topArticles, for: \.topArticles)
  }
}
