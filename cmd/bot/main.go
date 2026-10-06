package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"valersbot/internal/config"
	"valersbot/internal/handlers"
	"valersbot/internal/services"
	"valersbot/internal/storage"

	"github.com/bwmarrin/discordgo"
)

func main() {
	cfg := config.Load()
	if cfg.Token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}
	if cfg.GuildID == "" {
		log.Fatal("GUILD_ID is required")
	}

	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatalf("Error creando la sesión de Discord: %v", err)
	}

	store := storage.NewMemoryStore()
	lobbyService := services.NewLobbyService(store)
	lobby := handlers.NewLobbyHandler(lobbyService, cfg.LobbyRoleID)

	dg.AddHandler(lobby.HandleInteraction)

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("¡Bot conectado exitosamente como %s!\n", s.State.User.Username)
		if err := lobby.Register(s, cfg.GuildID); err != nil {
			log.Fatalf("Error registrando comandos: %v", err)
		}
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error abriendo conexión: %v", err)
	}
	defer dg.Close()

	fmt.Println("El bot está corriendo. Presiona CTRL-C para apagarlo.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
