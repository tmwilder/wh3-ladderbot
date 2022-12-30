package db

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(connectionString string) {
	m, err := migrate.New(
		getMigrationDirURL(),
		connectionString,
	)

	if err != nil {
		panic(err)
	}

	err2 := m.Up()

	if err2 != nil {
		panic(err2)
	}
}
