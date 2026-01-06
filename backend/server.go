package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"news-swipe/backend/cron"
	"news-swipe/backend/graph"
	"news-swipe/backend/graph/model"
	"news-swipe/backend/utils"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const defaultPort = "8080"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	host := os.Getenv("POSTGRES_HOST")
	pPort := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	sslmode := os.Getenv("POSTGRES_SSLMODE")

	if pPort == "" {
		pPort = "5432"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	if host == "" || user == "" || password == "" || dbname == "" {
		log.Fatal("Missing required Postgres environment variables")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, pPort, user, password, dbname, sslmode,
	)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&model.Article{}, &model.KeyWords{})

	// Initialize Redis
	if err := utils.InitRedis(); err != nil {
		log.Printf("Warning: Redis initialization failed: %v", err)
		log.Println("Continuing without Redis caching and rate limiting")
	}
	defer utils.CloseRedis()

	go cron.CreateCron(ctx, db)

	go func() {
		if err := graph.InitGraphQL(ctx, port, db); err != nil && err != http.ErrServerClosed {
			utils.Log(utils.GraphQL, "HTTP server error: ", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	utils.Log(utils.System, "Program started")
	<-sigCh

	utils.Log(utils.System, "Initiating shutdown")
	cancel()

	time.Sleep(time.Second)
	utils.Log(utils.System, "Program exited")
}
