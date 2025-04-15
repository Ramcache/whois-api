package config

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(cfg Config, logger *zap.Logger) *pgxpool.Pool {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Fatal("Failed to create DB connection pool", zap.Error(err))
	}

	if err = pool.Ping(ctx); err != nil {
		logger.Fatal("Database ping failed", zap.Error(err))
	}

	logger.Info("Database connection established successfully")
	return pool
}
