package repository

import (
	"log"
	"github.com/jackc/pgx"
)

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

func CreatePool(connectionDSN string) (*pgx.ConnPool, error) {
	config := parseConfig(connectionDSN)
	p, err := pgx.NewConnPool(*config)
	if err != nil {
		log.Printf("failed to create pool: %v", err)
		return nil, err
	}
	log.Println("create connection pool")
	return p, nil
}