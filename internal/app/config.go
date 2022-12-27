package app

import "os"

const DISCORD_V10_API_BASE = "https://discord.com/api/v10"

type DiscordAppConfig struct {
	DiscordBotToken     string
	DiscordAppId        string
	DiscordAppPublicKey string
}

func GetDiscordAppConfig() DiscordAppConfig {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		panic("Must provide DISCORD_BOT_TOKEN as env var.")
	}
	appId := os.Getenv("DISCORD_APP_ID")
	if appId == "" {
		panic("Must provide DISCORD_APP_ID as env var.")
	}
	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")
	if publicKey == "" {
		panic("Must provide DISCORD_PUBLIC_KEY as env var.")
	}
	return DiscordAppConfig{
		DiscordBotToken:     botToken,
		DiscordAppId:        appId,
		DiscordAppPublicKey: publicKey,
	}
}
