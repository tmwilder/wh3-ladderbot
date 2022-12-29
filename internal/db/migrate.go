package db

import (
	"discordbot/internal/app/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"net/http"
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

func MigrationHandler(c *gin.Context) {
	appConfig := config.GetAppConfig()
	requestKey, foundKey := c.GetQuery("admin_key")
	if !foundKey {
		c.JSON(http.StatusUnauthorized, "Must supply query param admin key.")
		return
	}
	if requestKey != appConfig.AdminKey {
		c.JSON(http.StatusUnauthorized, "Must supply correct query param admin key.")
		return
	}
	Migrate(GetMySQLConnStr())
	c.JSON(http.StatusOK, nil)
}
