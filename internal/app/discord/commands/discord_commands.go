package commands

import (
	"bytes"
	"discordbot/internal/app/config"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

const DiscordV10AppBase = "https://discord.com/api/v10"

type GlobalCommandPost struct {
	Name        CommandName     `json:"name"`
	Type        int             `json:"type"`
	Description string          `json:"description"`
	Options     []CommandOption `json:"options"`
}

type CommandOptionChoice struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type CommandOption struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Type        int                   `json:"type"`
	Required    bool                  `json:"required"`
	Choices     []CommandOptionChoice `json:"choices"`
}

type CommandName string

const (
	Queue   CommandName = "queue"
	Dequeue CommandName = "dequeue"
	Report  CommandName = "report"
)

type ReportOutcome int

const (
	Win    ReportOutcome = 0
	Loss   ReportOutcome = 1
	Cancel ReportOutcome = 2
)

/*
	InstallGlobalCommands
		Uploads our commands that are common to all installs of our app to the Discord bot defined
		by env variables. Defines this with the static config in this file.
*/
func InstallGlobalCommands(config config.DiscordAppConfig) {
	commands := []GlobalCommandPost{
		{
			Name:        Queue,
			Type:        1,
			Description: "Enter the matchmaking queue.",
			Options: []CommandOption{{
				Name:        "range",
				Description: "How many elo points up and down you want to match into. Defaults to 300.",
				Type:        4,
				Required:    false,
			}},
		},
		{
			Name:        Dequeue,
			Type:        1,
			Description: "Leave the matchmaking queue.",
		},
		{
			Name:        Report,
			Type:        1,
			Description: "Report the result of your most recent match.",
			Options: []CommandOption{{
				Name:        "outcome",
				Description: "Win or Loss reports a result, Cancel cancels the match without playing it or changing ratings.",
				Type:        4,
				Required:    true,
				Choices: []CommandOptionChoice{
					{
						Name:  "Win",
						Value: int(Win),
					},
					{
						Name:  "Loss",
						Value: int(Loss),
					},
					{
						Name:  "Cancel",
						Value: int(Cancel),
					},
				},
			}},
		},
	}

	for _, v := range commands {
		upsertCommand(config, v)
	}
}

func upsertCommand(config config.DiscordAppConfig, commandBody GlobalCommandPost) {
	url := fmt.Sprintf("%s/applications/%s/commands", DiscordV10AppBase, config.DiscordAppId)

	client := &http.Client{}
	data, err := json.Marshal(commandBody)
	if err != nil {
		panic(err)
	}
	reader := bytes.NewReader(data)
	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bot %s", config.DiscordBotToken))
	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, readErr := httputil.DumpResponse(resp, true)

	switch resp.StatusCode {
	case http.StatusCreated:
		fmt.Println("CommandName: " + commandBody.Name + " created...")
		break
	case http.StatusOK:
		fmt.Println("CommandName: " + commandBody.Name + " created...")
		break
	default:
		log.Fatal("Unrecognized status code for new command: " + string(commandBody.Name) + " " + fmt.Sprintf("%s", body))
	}

	if err != nil || readErr != nil {
		log.Fatal("Failure to post command: " + fmt.Sprintf("%s", body))
	}
}
