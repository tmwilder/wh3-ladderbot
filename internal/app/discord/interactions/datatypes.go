package interactions

import (
	"discordbot/internal/app/discord/commands"
)

type DiscordUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type DiscordMemberInfo struct {
	User DiscordUser `json:"user"`
}

type OptionData struct {
	Type  int    `json:"type"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type InteractionData struct {
	Options []OptionData         `json:"options"`
	Type    int                  `json:"type"`
	Name    commands.CommandName `json:"name"`
	Id      string               `json:"id"`
}

type Interaction struct {
	Type   int               `json:"type"`
	Token  string            `json:"token"`
	Member DiscordMemberInfo `json:"member"`
	Data   InteractionData   `json:"data"`
}
