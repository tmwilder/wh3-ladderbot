package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

type ConnectionParams struct {
	Username string
	Password string
	Host     string
	Port     string
}

func GetConnectionParams() ConnectionParams {
	username := os.Getenv("DB_USERNAME")
	if username == "" {
		panic("Must provide DB_USERNAME as env var.")
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		panic("Must provide DB_PASSWORD as env var.")
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		panic("Must provide DB_HOST as env var.")
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		panic("Must provide DB_PORT as env var.")
	}
	return ConnectionParams{Username: username, Password: password, Host: host, Port: port}
}

func getMySQLDSN(params ConnectionParams) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/ladderbot?parseTime=true", params.Username, params.Password, params.Host, params.Port)
}

func GetMySQLConnStr() string {
	return fmt.Sprintf("mysql://" + getMySQLDSN(GetConnectionParams()))
}

func GetTestMysSQLConnStr() string {
	params := ConnectionParams{
		Username: "root",
		Password: "password",
		Port:     "3306",
		Host:     "127.0.0.1",
	}
	return getMySQLDSN(params)
}

func getMigrationDirURL() (path string) {
	return "file://./internal/db/migrations"
}

func GetGorm(connectionStr string) (db *gorm.DB) {
	db, err := gorm.Open(mysql.Open(connectionStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func GetDbConn() (db *gorm.DB) {
	return GetGorm(getMySQLDSN(GetConnectionParams()))
}
