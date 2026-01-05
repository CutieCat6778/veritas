package graph

import (
	"context"
	"log"
	"net/http"
	"news-swipe/backend/utils"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/landrade/gqlgen-cache-control-plugin/cache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vektah/gqlparser/v2/ast"
	"gorm.io/gorm"
)

func InitGraphQL(ctx context.Context, port string, db *gorm.DB) error {
	resolver := &Resolver{DB: db}
	c := Config{Resolvers: resolver}

	router := chi.NewRouter()

	router.Use(FilterMiddleware)
	srv := handler.New(NewExecutableSchema(c))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.Use(cache.Extension{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", cache.Middleware(srv))
	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr: ":" + port,
	}

	// Handle shutdown
	go func() {
		<-ctx.Done()
		log.Println("Shutting down HTTP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Println("HTTP server shutdown error:", err)
		}
	}()

	// Start the server
	utils.Log(utils.Server, "GraphQL server starting", "port", port, "playground", "http://localhost:"+port+"/")
	return server.ListenAndServe()
}
