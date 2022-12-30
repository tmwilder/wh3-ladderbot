package config

import "os"

type AppConfig struct {
	DiscordBotToken     string
	DiscordAppId        string
	DiscordAppPublicKey string
	AdminKey            string
	HomeGuildId         string
}

func GetAppConfig() AppConfig {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		panic("Must provide DISCORD_BOT_TOKEN as env var.")
	}
	homeGuildId := os.Getenv("DISCORD_HOME_GUILD_ID")
	if homeGuildId == "" {
		panic("Must provide DISCORD_HOME_GUILD_ID as env var.")
	}
	appId := os.Getenv("DISCORD_APP_ID")
	if appId == "" {
		panic("Must provide DISCORD_APP_ID as env var.")
	}
	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")
	if publicKey == "" {
		panic("Must provide DISCORD_PUBLIC_KEY as env var.")
	}
	adminKey := os.Getenv("ADMIN_KEY")
	if publicKey == "" {
		panic("Must provide ADMIN_KEY as env var.")
	}

	return AppConfig{
		DiscordBotToken:     botToken,
		DiscordAppId:        appId,
		DiscordAppPublicKey: publicKey,
		AdminKey:            adminKey,
		HomeGuildId:         homeGuildId,
	}
}
