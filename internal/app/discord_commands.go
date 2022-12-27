package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

type GlobalCommandPost struct {
	Name        string          `json:"name"`
	Type        int             `json:"type"`
	Description string          `json:"description"`
	Options     []CommandOption `json:"options"`
}

type CommandOption struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        int    `json:"type"`
	Required    bool   `json:"required"`
}

func installGlobalCommands(config DiscordAppConfig) {
	commands := []GlobalCommandPost{
		{
			Name:        "queue",
			Type:        1,
			Description: "Enter the matchmaking queue",
			Options: []CommandOption{{
				Name:        "range",
				Description: "How many divisions up and down you want to match into. Defaults to 2.",
				Type:        4,
				Required:    false,
			}},
		},
		{
			Name:        "dequeue",
			Type:        1,
			Description: "Leave the matchmaking queue",
		},
	}

	for _, v := range commands {
		upsertCommand(config, v)
	}
}

func upsertCommand(config DiscordAppConfig, commandBody GlobalCommandPost) {
	url := fmt.Sprintf("%s/applications/%s/commands", DISCORD_V10_API_BASE, config.DiscordAppId)

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
	// req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bot %s", config.DiscordBotToken))
	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, readErr := httputil.DumpResponse(resp, true)

	switch resp.StatusCode {
	case http.StatusCreated:
	case http.StatusOK:
		fmt.Println("Command: " + commandBody.Name + " created...")
		break
	default:
		log.Fatal("Unrecognized status code for new command: " + commandBody.Name + " " + fmt.Sprintf("%s", body))
	}

	if err != nil || readErr != nil {
		log.Fatal("Failure to post command: " + fmt.Sprintf("%s", body))
	}
}
