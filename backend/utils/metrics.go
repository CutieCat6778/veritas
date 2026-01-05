package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Scraper metrics
	ScraperRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veritas_scraper_requests_total",
			Help: "Total number of scraper requests by source",
		},
		[]string{"source"},
	)

	ScraperErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veritas_scraper_errors_total",
			Help: "Total number of scraper errors by source",
		},
		[]string{"source"},
	)

	ScraperDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "veritas_scraper_duration_seconds",
			Help:    "Duration of scraper requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"source"},
	)

	ArticlesScraped = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veritas_articles_scraped_total",
			Help: "Total number of articles scraped by source",
		},
		[]string{"source"},
	)

	// GraphQL metrics
	GraphQLRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veritas_graphql_requests_total",
			Help: "Total number of GraphQL requests by operation",
		},
		[]string{"operation"},
	)

	GraphQLErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veritas_graphql_errors_total",
			Help: "Total number of GraphQL errors by operation",
		},
		[]string{"operation"},
	)

	GraphQLDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "veritas_graphql_duration_seconds",
			Help:    "Duration of GraphQL operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Database metrics
	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veritas_database_queries_total",
			Help: "Total number of database queries by type",
		},
		[]string{"type"},
	)

	DatabaseErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "veritas_database_errors_total",
			Help: "Total number of database errors",
		},
	)

	// Article metrics
	TotalArticlesGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "veritas_articles_total",
			Help: "Total number of articles in database",
		},
	)

	ArticlesPerSourceGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "veritas_articles_per_source",
			Help: "Number of articles per source",
		},
		[]string{"source"},
	)

	// Cron job metrics
	CronJobRunsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veritas_cron_job_runs_total",
			Help: "Total number of cron job runs",
		},
		[]string{"job"},
	)

	CronJobErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "veritas_cron_job_errors_total",
			Help: "Total number of cron job errors",
		},
		[]string{"job"},
	)

	CronJobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "veritas_cron_job_duration_seconds",
			Help:    "Duration of cron jobs in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{"job"},
	)
)
