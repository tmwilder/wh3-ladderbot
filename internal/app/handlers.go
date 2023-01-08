package app

import (
	"discordbot/internal/app/config"
	"discordbot/internal/app/discord/api"
	"discordbot/internal/app/discord/commands"
	"discordbot/internal/app/discord/interactions"
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
		"The goal of the WCL is to create a welcoming environment for both new players and hardened veterans to sharpen their skills. At the end of the day, this is about growing the WH3 multiplayer community and getting more people involved in the competitive scene. If you’re thinking about making the leap from quick battles/campaign into the competitive scene, this is a great place to start!\n",

		"**Ground Rules:**",
		"1. Treat ALL players with respect.",
		"2. Welcome and support new players.",
		"3. Play fair and have fun!\n",

		"**How to play:**",
		"1. Type the command ‘/queue’ in the #find-matches channel. You will have the option of choosing from gametypes Bo1, Bo3, or All. This will queue you up and attempt to match you against a player of a similar ELO rating. Wait in the queue until paired with an opponent (matchmaking duration may vary).",
		"a. If you’d like to restrict your opponents to a specific ELO range you can do so with the command ‘/queue elo’. For example, if your current ELO rating is 1000 and you enter the command ‘/queue 400`, you will be matched with players with ratings between 600-1400.",
		"b. If you want to dequeue from matchmaking, type the command ‘/dequeue`.",
		"2. Once you match with an opponent, contact them via Discord, and play your match using the format provided.",
		"a. If after matching with an opponent, you want to cancel the match without playing, use the command `/report cancel`. This is not intended for avoiding specific players.",
		"b. If you cancel the match after Picks & Bans have started, your opponent may choose to report the match as a win.",
		"3. When the match is completed, report the results with the commands `/report win` or `/report loss`. Only one player needs to report the results. After reporting, your ratings and records will be automatically updated and you can immediately queue again for further matches.",
		"4. If you or your opponent enters the wrong match results simply input the ‘/report win' or `/report loss` command again and it will overwrite the previous match results and ratings.\n",

		"**Leaderboard:**",
		"Each month, the WCL player with the most wins will be declared the winner! On the first of the month, the Leaderboard will be reset. Current standings are visible in the #leaderboard channel.",
		"Player ELO ratings are also being tracked for optimized matchmaking. These can be found in the #elo-ratings channel.\n",

		"**Bo3 Format:**",
		"All games are played in Domination format.",
		"Picks & Bans",
		"Global Bans: At the start of the match, each player chooses two global bans in the following order: P1, P2, P2, P1.",
		"Globally banned factions cannot be played by either player for the duration of the match.",
		"Local Bans: At the start of each game, players will take turns picking factions and local bans (see below). Local bans only apply to the game currently being played and are reset for the following game.",
		"A player may not local ban the same faction twice in a match. A player may not play the same faction twice in a match.",
		"After global bans, the Pick & Ban selections proceed as follows:\n",

		"Game 1",
		"P1 Picks 3 potential factions & Bans 1 of P2’s factions",
		"P2 Picks 1 faction & Bans 1 of P1’s 3 picks",
		"P1 Picks 1 of their remaining 2 factions\n",

		"Game 2",
		"P2 Picks 3 potential factions & Bans 1 of P1’s factions",
		"P1 Picks 1 faction & Bans 1 of P2’s 3 picks",
		"P2 Picks 1 of their remaining 2 factions\n",

		"Game 3 (Ace Match)",
		"Winner of game 2 picks 3 potential factions & Bans 1 opponent faction",
		"Loser of game 2 picks 1 faction & Bans 1 of opponent’s 3 picks",
		"Winner of game 2 Picks 1 of their remaining 2 factions\n",

		"**Bo1 Format:**",
		"Use the https://aoe2cm.net/preset/DNEXyv Pick & Ban tool.",
		"Each player chooses 3 global bans.",
		"Each player chooses 1 blind pick.",
		"A random map will be assigned and you may begin the game.\n",

		"**Ratings:**",
		"We use a standard Elo rating system to provide better matchmaking. You can see current ratings in #elo-ratings.",
		"Starting Elo is 1200. K for bo3 games is 32 which means the most your rating can change up or down is 32 points.",
		"Your rating will move more when you beat a much higher rated player or lose to a much lower rated player.",
		"New players have a provisional K value of 64 for their first 10 games to converge to an accurate rating faster.",
		"bo1 games have their K values halved to reflect the shorter time commitment and less competitive nature of bo1.",
		"Treat ratings as a useful matchmaking tool and that's it. We compete to win season score, not to be the highest rated player.\n",

		"**Crashes/Disconnects:**",
		"Suspected intentional abuse of crashes/disconnects will result in moderator review and potential suspension from the ladder. Play fair and try to work things out with your opponent.",
		"If at the time of the crash/disconnect there is a clear winner, that player is considered the winner. (see the Ground Rules).",
		"If at the time of the crash/disconnect there isn’t a clear winner, replay the match using the same armies.",
		"If your opponent crashes/disconnects and is unresponsive for 10-minutes, you may report the match as a win.",
		"If you cannot come to an agreeable resolution with your opponent, contact an admin. Remember the Ground Rules: treat opponents with respect and fairness. We're playing  here for fun and self-improvement - often it's best to just take a minor ratings hit and move on!\n",

		"**Map Pool:**",
		"Maps will be randomly chosen from the list below for each match.",
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

	PostMonthlyWinStandings(conn)
	PostEloStandings(conn)
}

func expireMatchRequestsHandler(c *gin.Context) {
	authorized := AuthorizeAdminAction(c)
	if !authorized {
		return
	}
	conn := db.GetDbConn()
	interactions.ExpireMatchRequests(conn)
}

func PostMonthlyWinStandings(conn *gorm.DB) {
	usersWithStats := db.GetMonthlyWinLeaderboard(conn)
	leaderBoardLines := []string{"Total wins this month: \n"}
	for i, v := range usersWithStats {
		line := fmt.Sprintf("%d - %s - %dW / %dL", i+1, v.User.DiscordUserName, v.Wins, v.Losses)
		leaderBoardLines = append(leaderBoardLines, line)
	}
	api.ReplaceChannelContents(config.GetAppConfig().HomeGuildId, "leaderboard", leaderBoardLines)
}

func PostEloStandings(conn *gorm.DB) {
	usersWithStats := db.GetEloLeaderboard(conn)
	leaderBoardLines := []string{"All time top Elo Ratings: \n"}
	for i, v := range usersWithStats {
		line := fmt.Sprintf("%d - %s - Elo %d - %dW / %dL", i+1, v.User.DiscordUserName, v.User.CurrentRating, v.Wins, v.Losses)
		leaderBoardLines = append(leaderBoardLines, line)
	}
	api.ReplaceChannelContents(config.GetAppConfig().HomeGuildId, "elo-ratings", leaderBoardLines)
}
