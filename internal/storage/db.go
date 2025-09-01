package storage

import (
	"context"
	"errors"
	"log"
	"time"
	"github.com/jackc/pgx"
	//"github.com/kujoki/go-musthave-service/internal/model"
)

type PostgresRepository struct {
    Pool *pgx.ConnPool
}

func parseConfig(connectionDSN string) (*pgx.ConnPoolConfig) {
	config, err := pgx.ParseConnectionString(connectionDSN)
	if err != nil {
		log.Printf("failed to parse connection string: %v", err)
		return nil
	}
	log.Println("successful parse connection string")
	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     config,
		MaxConnections: 10,
	}
	return &poolConfig
}

func NewPostgresRepository(connectionDSN string) (*PostgresRepository, error) {
	config := parseConfig(connectionDSN)
	pool, err := pgx.NewConnPool(*config)
	if err != nil {
		log.Printf("failed to create pool: %v", err)
		return nil, err
	}
	log.Println("create connection pool")
	return &PostgresRepository{Pool: pool}, nil
}


func (r *PostgresRepository) GetLongURL(shortURL string) (string, bool, error) {
    var longURL string
    err := r.Pool.QueryRow("SELECT long_url FROM url_data WHERE short_url=$1", shortURL).Scan(&longURL)
    if errors.Is(err, pgx.ErrNoRows) {
        return "", false, nil
    } else if err != nil {
        return "", false, err
    }
    return longURL, true, nil
}

func (r *PostgresRepository) GetShortURL(longURL string) (string, bool, error) {
	var shortURL string
    err := r.Pool.QueryRow("SELECT short_url FROM url_data WHERE long_url=$1", longURL).Scan(&shortURL)
    if errors.Is(err, pgx.ErrNoRows) {
		log.Printf("error during get short url query - there is no rows: %v \n", err)
        return "", false, nil
    } else if err != nil {
		log.Printf("error during get short url query: %v \n", err)
        return "", false, err
    }
    return shortURL, true, nil
}

func (r *PostgresRepository) SaveURL(shortURL string, longURL string) error {
    now := time.Now().UTC()
    _, err := r.Pool.Exec(`
        INSERT INTO url_data (short_url, long_url, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (short_url) DO UPDATE
        SET long_url = EXCLUDED.long_url, updated_at = EXCLUDED.updated_at
    `, shortURL, longURL, now, now)
    return err
}

func (r *PostgresRepository) GetAll() map[string]string {
	rows, err := r.Pool.Query(`SELECT short_url, long_url FROM url_data`)
	if err != nil {
		log.Printf("error during get all data query: %v", err)
		return nil
	}
	defer rows.Close()

	copyData := make(map[string]string)

	for rows.Next() {
		var shortURL, longURL string
		if err := rows.Scan(&shortURL, &longURL); err != nil {
			log.Printf("error during scaning: %v", err)
			continue
		}
		copyData[shortURL] = longURL
	}

	if rows.Err() != nil {
		log.Printf("there are no rows: %v", rows.Err())
	}

	log.Printf("result: %+v", copyData)
	return copyData
}


func (r *PostgresRepository) Close() {
    r.Pool.Close()
}

func (r *PostgresRepository) Ping(ctx context.Context) bool {
	return true
}