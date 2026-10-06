package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found. Reading environment variables from the system.")
	}

	return &Config{
		Token:       strings.Trim(os.Getenv("DISCORD_TOKEN"), `"'`),
		GuildID:     strings.Trim(os.Getenv("GUILD_ID"), `"'`),
		LobbyRoleID: strings.Trim(os.Getenv("LOBBY_ROLE_ID"), `"'`),
	}
}
