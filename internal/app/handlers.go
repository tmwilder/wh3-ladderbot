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

	rulesAndMapsCopy := []string{
		"**Welcome to the Warhammer Community Ladder!**",
		"The WCL is a ranked matchmaking Ladder whose goal is to grow the WH3 competitive community by providing an accessible starting point for new players and easy access to high quality matches for everyone.\n",

		"**Ground Rules:**",
		"1. Don't be an asshole.",
		"2. Be supportive and welcoming of new players.",
		"3. That's it.\n",

		"**How to play:**",
		"1. Use the command `/queue` in the channel #find-matches. This will queue you up and try to pair you against a player of matching skill when possible. If you want you can define an elo range above and below your rating to restrict matches to with `/queue <elo>`.",
		"2. Once you are paired contact your opponent and play your match.",
		"3. When done you or your opponent reports results with `/report win` or `/report loss`. Your ratings and records will be updated and you can immediately queue again after reporting.",
		"4. If you want to leave the queue before getting matched you can use `/dequeue`. If you want to cancel a match without playing it you can use `/report cancel`. If you or your opponent puts in the wrong match results you can fix this by just reporting again with the right result, which will overwrite your previous entry.\n",

		"**Leaderboard:**",
		"Players compete to have the most wins each month. On the first of the month we reset. Current standings are visible in #leaderboard.",
		"We also track elo ratings for players so that we can do better matchmaking. These can be found in #elo-ratings.\n",

		"**Bo3 Format:**",
		"All games are Domination format.",
		"-Pick & Ban; (P1= first player posted by the bot; P2= second player posted by the bot.)",
		"Players will have two global bans each. At the start of the match players take turns banning factions in the order: P1, P2, P2, P1.",
		"Globally banned factions cannot be played by either player for the duration of the match.",
		"Factions banned for individual games are local bans meaning they only impact your opponent and only last one game. (i.e. if you ban empire in game 1 your opponent cannot play empire in game 1 but will be able to play it in game 2 or 3)",
		"A player may not local ban the same faction twice in a match. A player may not play the same faction twice in a match.",
		"After globals match selections proceed as follows:\n",

		"Game 1",
		"Player 1 (Top of Bracket) Picks 3 & Bans 1",
		"Player 2 then picks 1, and bans 1 of player twos 3 choices",
		"Player 1 Picks one of remaining 2\n",

		"Game 2",
		"Player 2 Picks 3 & Bans 1",
		"Player 1 then picks 1, and bans 1 of player twos 3 choices",
		"Player 2 Picks one of remaining 2\n",

		"Game 3 (Ace Match)",
		"Winner of game 2 picks 3 and bans 1 (the player can not ban the same faction that he/she already banned in game 1 or 2)",
		"Loser of game 2 picks 1 and bans of the winners 3 choices",
		"Winner of game 2 then picks 1 of remaining two choices\n",

		"**Crashes/Disconnects:**",
		"If it's clear who would have won, give the game to that player",
		"Else, if it's before the first engagement, reset the game. Both players must deploy in the same fashion and summon the same units. The player that didn't DC gets to choose which side of the map they start on",
		"Else, win goes to the person that didn't DC",
		"Contact an admin if there's an unresolved dispute, but do try hard to resolve it. See Ground Rules #1 - players are expected to resolve most disputes amicably. Remember that we're playing for fun and self improvement - often it's best to just take a minor ratings hit and move on!\n",

		"**Map Pool:**",
	}

	for _, v := range maps {
		rulesAndMapsCopy = append(rulesAndMapsCopy, v)
	}

	api.ReplaceChannelContents(config.GetAppConfig().HomeGuildId, "rules-and-maps", rulesAndMapsCopy)

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
		line := fmt.Sprintf("%d - %s - %dW / %dL", i+1, v.User.DiscordUserName, v.Wins, v.Losses)
		leaderBoardLines = append(leaderBoardLines, line)
	}
	api.ReplaceChannelContents(config.GetAppConfig().HomeGuildId, "leaderboard", leaderBoardLines)
}

func postEloStandings(conn *gorm.DB) {
	usersWithStats := db.GetEloLeaderboard(conn)
	leaderBoardLines := []string{"All time top Elo Ratings: \n"}
	for i, v := range usersWithStats {
		line := fmt.Sprintf("%d - %s - Elo %d - %dW / %dL", i+1, v.User.DiscordUserName, v.User.CurrentRating, v.Wins, v.Losses)
		leaderBoardLines = append(leaderBoardLines, line)
	}
	api.ReplaceChannelContents(config.GetAppConfig().HomeGuildId, "elo-ratings", leaderBoardLines)
}
