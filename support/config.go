package support

import (
	"os"
)

// Config is a config interface.
var Config config

type config struct {
	DiscordToken string
	ChannelID    string
	GuildID      string
	UserRole     string
	MessageID    string
	SleepSec     string
}

func (conf *config) LoadEnv() {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		print("No .env file.\n")
		os.Exit(1)
	}

	Config = config{
		DiscordToken: os.Getenv("DiscordToken"),
		ChannelID:    os.Getenv("ChannelID"),
		GuildID:      os.Getenv("GuildID"),
		UserRole:     os.Getenv("UserRole"),
		MessageID:    os.Getenv("MessageID"),
		SleepSec:     os.Getenv("SleepSec"),
	}

}
