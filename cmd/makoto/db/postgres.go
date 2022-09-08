package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func ConnectPostgres(uri string) *sql.DB {
	con, err := sql.Open("postgres", uri)
	if err != nil {
		panic(err)
	}
	return con
}
