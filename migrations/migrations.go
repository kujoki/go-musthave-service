package migrations

import (
    "log"
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    "database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigrations(dsn string) {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        log.Fatalf("failed to create driver: %v", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations",
        "postgres", driver)
    if err != nil {
        log.Fatalf("failed to init migrate: %v", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatalf("migration failed: %v", err)
    }

    log.Println("migrations applied successfully")
}
