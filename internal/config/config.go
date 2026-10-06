package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found. Reading environment variables from the system.")
	}

	return &Config{
		Token:   os.Getenv("DISCORD_TOKEN"),
		GuildID: os.Getenv("GUILD_ID"),
	}
}
