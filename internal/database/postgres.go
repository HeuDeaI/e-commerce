package database

import (
	"context"
	"e-commerce/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool   *pgxpool.Pool
	config *config.PostgresConfig
}

func New(ctx context.Context, cfg *config.PostgresConfig) (*Database, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &Database{
		Pool:   pool,
		config: cfg,
	}, nil
}

func (db *Database) Close() error {
	db.Pool.Close()
	return nil
}

func (db *Database) Migrate(ctx context.Context, migrationsPath string) error {
	m, err := migrate.New(
		"file://"+migrationsPath,
		db.config.DSN(),
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
