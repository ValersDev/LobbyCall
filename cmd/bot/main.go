package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)
 
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error cargando el archivo .env: %v", err)
	}

	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Error creando la sesión de Discord: %v", err)
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("¡Bot conectado exitosamente como %s!\n", s.State.User.Username)
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