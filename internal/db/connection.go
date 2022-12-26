package db

import "os"

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
