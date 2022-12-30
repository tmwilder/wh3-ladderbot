package app

import (
	"discordbot/internal/app/config"
	"discordbot/internal/app/discord/api"
	"discordbot/internal/app/discord/commands"
	"discordbot/internal/db"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"net/http"
)

func AuthorizeAdminAction(c *gin.Context) (authorized bool) {
	appConfig := config.GetAppConfig()
	requestKey, foundKey := c.GetQuery("admin_key")
	if !foundKey {
		c.JSON(http.StatusUnauthorized, "Must supply query param admin key.")
		return false
	}
	if requestKey != appConfig.AdminKey {
		c.JSON(http.StatusUnauthorized, "Must supply correct query param admin key.")
		return false
	}
	return true
}

/*
	We use exclusively global commands because there are no use cases to scope to one guild.
	https://discord.com/developers/docs/interactions/application-commands#authorizing-your-application
*/
func installSlashCommandsHandler(c *gin.Context) {
	authorized := AuthorizeAdminAction(c)
	if !authorized {
		return
	}
	commands.InstallGlobalCommands(config.GetAppConfig())
	c.JSON(http.StatusOK, "Successfully installed all slash commands!")
}

func migrationHandler(c *gin.Context) {
	authorized := AuthorizeAdminAction(c)
	if !authorized {
		return
	}
	db.Migrate(db.GetMySQLConnStr())
	c.JSON(http.StatusOK, "Successfully migrated to latest version!")
}

func setMapsHandler(c *gin.Context) {
	authorized := AuthorizeAdminAction(c)
	if !authorized {
		return
	}

	requestBodyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(err)
	}

	var maps []string

	err = json.Unmarshal(requestBodyData, &maps)
	if err != nil {
		panic(err)
	}

	// We're not bifurcating yet.
	firstPersisted := db.InsertMapSet(db.GetDbConn(), maps, db.Bo3)

	if !firstPersisted {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	secondPersisted := db.InsertMapSet(db.GetDbConn(), maps, db.Bo1)

	if !secondPersisted {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, "Maps updated.")
}

func updateLeaderBoardHandler(c *gin.Context) {
	authorized := AuthorizeAdminAction(c)
	if !authorized {
		return
	}
	conn := db.GetDbConn()

	postMonthlyWinStandings(conn)
	postEloStandings(conn)
}

func postMonthlyWinStandings(conn *gorm.DB) {
	usersWithStats := db.GetMonthlyWinLeaderboard(conn)
	leaderBoardLines := []string{"Total wins this month: \n"}
	for i, v := range usersWithStats {
		line := fmt.Sprintf("%d - %s - %dW/%dL", i+1, v.User.DiscordUserName, v.Wins, v.Losses)
		leaderBoardLines = append(leaderBoardLines, line)
	}
	api.ReplaceChannelContents(config.GetAppConfig().HomeGuildId, "leaderboard", leaderBoardLines)
}

func postEloStandings(conn *gorm.DB) {
	usersWithStats := db.GetEloLeaderboard(conn)
	leaderBoardLines := []string{"All time top Elo Ratings: \n"}
	for i, v := range usersWithStats {
		line := fmt.Sprintf("%d - %s - Elo %d - %dW/%dL", i+1, v.User.DiscordUserName, v.User.CurrentRating, v.Wins, v.Losses)
		leaderBoardLines = append(leaderBoardLines, line)
	}
	api.ReplaceChannelContents(config.GetAppConfig().HomeGuildId, "elo-standings", leaderBoardLines)
}
