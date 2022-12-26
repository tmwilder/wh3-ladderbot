package main

import (
	"discordbot/internal/db"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	db.Migrate(db.GetMySQLConnStr())
}
