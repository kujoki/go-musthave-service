package storage

import (
	"context"
	"errors"
	"log"
	"time"
	"github.com/jackc/pgx"
	"github.com/kujoki/go-musthave-service/internal/model"
	"strings"
	"fmt"
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


func (r *PostgresRepository) GetLongURL(shortURL string) (model.LongURLResult, bool, error) {
    var longURLRes model.LongURLResult
    err := r.Pool.QueryRow("SELECT long_url, is_deleted FROM url_data WHERE short_url=$1",
	 shortURL).Scan(&longURLRes.OriginURL, &longURLRes.IsDeleted)
    if errors.Is(err, pgx.ErrNoRows) {
        return longURLRes, false, nil
    } else if err != nil {
        return longURLRes, false, err
    }
    return longURLRes, true, nil
}

func (r *PostgresRepository) GetShortURL(longURL string) (model.ShortURLResult, bool, error) {
	var shortURLRes model.ShortURLResult
    err := r.Pool.QueryRow("SELECT short_url, is_deleted FROM url_data WHERE long_url=$1", longURL).Scan(&shortURLRes.ShortURL, &shortURLRes.IsDeleted)
    if errors.Is(err, pgx.ErrNoRows) {
		log.Printf("error during get short url query - there is no rows: %v \n", err)
        return shortURLRes, false, nil
    } else if err != nil {
		log.Printf("error during get short url query: %v \n", err)
        return shortURLRes, false, err
    }
    return shortURLRes, true, nil
}

func (r *PostgresRepository) SaveURL(shortURL string, longURL string, userUUID string) error {
    now := time.Now().UTC()
    _, err := r.Pool.Exec(`
        INSERT INTO url_data (short_url, long_url, username, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (short_url) DO UPDATE
        SET long_url = EXCLUDED.long_url, updated_at = EXCLUDED.updated_at
    `, shortURL, longURL, userUUID, now, now)
    return err
}

func (r *PostgresRepository) GetAll() map[string]model.URLRecord {
    rows, err := r.Pool.Query(`
        SELECT short_url, long_url, username, is_deleted 
        FROM url_data
    `)
    if err != nil {
        log.Printf("error during get all data query: %v", err)
        return nil
    }
    defer rows.Close()

    copyData := make(map[string]model.URLRecord)

    for rows.Next() {
        var shortURL, longURL, username string
        var isDeleted bool

        if err := rows.Scan(&shortURL, &longURL, &username, &isDeleted); err != nil {
            log.Printf("error during scanning: %v", err)
            continue
        }

        if !isDeleted {
            copyData[shortURL] = model.URLRecord{
                LongURL:   longURL,
                UserUUID: username,
                IsDeleted: isDeleted,
            }
        }
    }

    if rows.Err() != nil {
        log.Printf("rows error: %v", rows.Err())
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

func (r *PostgresRepository) DeleteURL(ctx context.Context, tasks []model.Task) error {
    if len(tasks) == 0 {
        return nil
    }
	log.Println("start to delete short URLs")

    var pairs []string
    var args []interface{}

    for _, task := range tasks {
        args = append(args, task.UserUUID, task.Item)
        pairs = append(pairs, fmt.Sprintf("($%d, $%d)", len(args)-1, len(args)))
    }

    query := fmt.Sprintf(`
        UPDATE url_data
        SET is_deleted = true
        WHERE (username, short_url) IN (%s)`,
        strings.Join(pairs, ", "))

    _, err := r.Pool.Exec(query, args...)
	log.Println("delete process was done")
	if err != nil {
		log.Println("there is an error during delete process ", err)
	}
	log.Println("result after deleting", r.GetAll())
	return err
	} 

func (r *PostgresRepository) GetUserURLs(userUUID string) ([]model.UserURL, error) {
	var userURLs []model.UserURL
    rows, err := r.Pool.Query("SELECT short_url, long_url FROM url_data WHERE username=$1 AND is_deleted=false", userUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u model.UserURL
		if err := rows.Scan(&u.ShortURL, &u.OriginalURL); err != nil {
			return nil, err
		}
		userURLs = append(userURLs, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

    return userURLs, nil
}