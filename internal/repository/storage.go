package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type Storage struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewStorage(ctx context.Context, user string, password string, host string, port string, dbname string, sslmode string, log *zap.Logger) (*Storage, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbname, sslmode)
	log = log.With(zap.String("type", "Storage"))

	log.Info("Connecting to PostgreSQL database",
		zap.String("dbname", dbname),
		zap.String("user", user),
		zap.String("sslmode", sslmode))

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Error("Error parsing connection string", zap.Error(err))
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2

	db, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		log.Error("Error connecting to PostgreSQL database", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	log.Info("Testing database connection")
	if err := db.Ping(context.Background()); err != nil {
		log.Error("Failed to ping PostgreSQL database", zap.String("dbname", dbname), zap.Error(err))
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	log.Info("Successfully connected to database")

	log.Info("Starting database migrations")

	if err := runMigrations(connStr); err != nil {
		log.Error("Failed to run migrations", zap.Error(err))
		return nil, fmt.Errorf("failed to run migration: %w", err)
	}
	log.Info("Successfully migrated database")
	return &Storage{
		db:  db,
		log: log,
	}, nil
}

func runMigrations(connStr string) error {
	m, err := migrate.New("file://C:/Users/Равиль/GolandProjects/L0/migrations", connStr)
	if err != nil {
		return fmt.Errorf("start migrations error %v", err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("migration up error: %v", err)
	}
	return nil
}
