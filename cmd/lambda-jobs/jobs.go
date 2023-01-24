package main

import (
	"context"
	"discordbot/internal/app"
	"discordbot/internal/app/discord/api"
	"discordbot/internal/app/discord/interactions"
	"discordbot/internal/db"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
)

func HandleRequest(ctx context.Context) (string, error) {
	conn := db.GetDbConn()
	discordApi := api.ConcreteDiscordApi{}
	expirySuccess := interactions.ExpireMatchRequests(conn, discordApi)
	if !expirySuccess {
		log.Panic("Unable to expire match requests...")
	}
	app.PostMonthlyWinStandings(conn)
	app.PostEloStandings(conn)
	return "Success!", nil
}

/*
	This is uploaded to a 2nd lambda in prod that is used to run our scheduled jobs.
*/
func main() {
	lambda.Start(HandleRequest)
}
