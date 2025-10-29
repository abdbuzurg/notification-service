package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
)

func NewPostgresConnection() (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("postgresql.host"),
		viper.GetString("postgresql.port"),
		viper.GetString("postgresql.user"),
		viper.GetString("postgresql.password"),
		viper.GetString("postgresql.dbname"),
		viper.GetString("postgresql.sslmode"),
	)

	pool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL.")
	return pool, nil
}
