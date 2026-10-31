package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Database struct {
	Host     string
	User     string
	Password string
	Name     string
}

type PostgresClient struct {
	DB *sqlx.DB
}

func NewPostgresClient(dbCfg *Database) (pgClient *PostgresClient, err error) {
	pgClient = &PostgresClient{}
	pgClient.DB, err = sqlx.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
			dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Name))
	if err != nil {
		return
	}
	err = pgClient.DB.Ping()
	return
}
