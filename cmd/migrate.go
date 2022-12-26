package main

import (
	"discordbot/internal/db"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	connectionParams := db.GetConnectionParams()

	m, err := migrate.New(
		"file://./migrations",
		fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/ladderbot", connectionParams.Username, connectionParams.Password, connectionParams.Host, connectionParams.Port),
	)

	if err != nil {
		panic(err)
	}

	err2 := m.Up()

	if err2 != nil {
		panic(err2)
	}
}
