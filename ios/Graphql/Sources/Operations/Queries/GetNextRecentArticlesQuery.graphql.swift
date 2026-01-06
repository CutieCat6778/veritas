// @generated
// This file was automatically generated and should not be edited.

@_exported import ApolloAPI
@_spi(Execution) @_spi(Unsafe) import ApolloAPI

public struct GetNextRecentArticlesQuery: GraphQLQuery {
  public static let operationName: String = "GetNextRecentArticles"
  public static let operationDocument: ApolloAPI.OperationDocument = .init(
    definition: .init(
      #"query GetNextRecentArticles($start: Int!, $stop: Int!) { nextRecentArticle(start: $start, stop: $stop) { __typename ...ArticleFields ...ArticleWithLinks } }"#,
      fragments: [ArticleFields.self, ArticleWithLinks.self]
    ))

  public var start: Int32
  public var stop: Int32

  public init(
    start: Int32,
    stop: Int32
  ) {
    self.start = start
    self.stop = stop
  }

  @_spi(Unsafe) public var __variables: Variables? { [
    "start": start,
    "stop": stop
  ] }

  public struct Data: Graphql.SelectionSet {
    @_spi(Unsafe) public let __data: DataDict
    @_spi(Unsafe) public init(_dataDict: DataDict) { __data = _dataDict }

    @_spi(Execution) public static var __parentType: any ApolloAPI.ParentType { Graphql.Objects.Query }
    @_spi(Execution) public static var __selections: [ApolloAPI.Selection] { [
      .field("nextRecentArticle", [NextRecentArticle?].self, arguments: [
        "start": .variable("start"),
        "stop": .variable("stop")
      ]),
    ] }
    @_spi(Execution) public static var __fulfilledFragments: [any ApolloAPI.SelectionSet.Type] { [
      GetNextRecentArticlesQuery.Data.self
    ] }

    public var nextRecentArticle: [NextRecentArticle?] { __data["nextRecentArticle"] }

    public init(
      nextRecentArticle: [NextRecentArticle?]
    ) {
      self.init(unsafelyWithData: [
        "__typename": Graphql.Objects.Query.typename,
        "nextRecentArticle": nextRecentArticle._fieldData,
      ])
    }

    /// NextRecentArticle
    ///
    /// Parent Type: `Article`
    public struct NextRecentArticle: Graphql.SelectionSet {
      @_spi(Unsafe) public let __data: DataDict
      @_spi(Unsafe) public init(_dataDict: DataDict) { __data = _dataDict }

      @_spi(Execution) public static var __parentType: any ApolloAPI.ParentType { Graphql.Objects.Article }
      @_spi(Execution) public static var __selections: [ApolloAPI.Selection] { [
        .field("__typename", String.self),
        .fragment(ArticleFields.self),
        .fragment(ArticleWithLinks.self),
      ] }
      @_spi(Execution) public static var __fulfilledFragments: [any ApolloAPI.SelectionSet.Type] { [
        GetNextRecentArticlesQuery.Data.NextRecentArticle.self,
        ArticleFields.self,
        ArticleWithLinks.self
      ] }

      public var id: Graphql.ID { __data["id"] }
      public var title: String { __data["title"] }
      public var source: GraphQLEnum<Graphql.Source> { __data["source"] }
      public var publishedAt: String { __data["publishedAt"] }
      public var uri: String { __data["uri"] }
      public var views: Int { __data["views"] }
      public var description: String { __data["description"] }
      public var banner: String { __data["banner"] }
      public var category: [String?]? { __data["category"] }
      public var linkedTo: [LinkedTo?]? { __data["linkedTo"] }

      public struct Fragments: FragmentContainer {
        @_spi(Unsafe) public let __data: DataDict
        @_spi(Unsafe) public init(_dataDict: DataDict) { __data = _dataDict }

        public var articleFields: ArticleFields { _toFragment() }
        public var articleWithLinks: ArticleWithLinks { _toFragment() }
      }

      public init(
        id: Graphql.ID,
        title: String,
        source: GraphQLEnum<Graphql.Source>,
        publishedAt: String,
        uri: String,
        views: Int,
        description: String,
        banner: String,
        category: [String?]? = nil,
        linkedTo: [LinkedTo?]? = nil
      ) {
        self.init(unsafelyWithData: [
          "__typename": Graphql.Objects.Article.typename,
          "id": id,
          "title": title,
          "source": source,
          "publishedAt": publishedAt,
          "uri": uri,
          "views": views,
          "description": description,
          "banner": banner,
          "category": category,
          "linkedTo": linkedTo._fieldData,
        ])
      }

      public typealias LinkedTo = ArticleWithLinks.LinkedTo
    }
  }
}
