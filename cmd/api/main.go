package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ifaisalabid1/url-shortener/internal/config"
	"github.com/ifaisalabid1/url-shortener/internal/repository"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))

	slog.SetDefault(logger)

	db, err := connectDB(cfg)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	defer db.Close()

	redisClient, err := connectRedis(cfg)
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	defer redisClient.Close()

	urlRepo := repository.NewURLRepository(db)
	cacheRepo := repository.NewClientRepository(redisClient)
}

func connectDB(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.DBName,
		cfg.Postgres.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to open dsn: %w", err)
	}

	db.SetMaxOpenConns(cfg.Postgres.MaxConns)
	db.SetMaxIdleConns(cfg.Postgres.MinConns)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to ping database: %w", err)
	}

	return db, nil

}

func connectRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: 10,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}
