package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"

	"release-bot/internal/handlers"
	"release-bot/internal/state"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is empty")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	store := &state.Store{}
	h := handlers.New(store)

	dg.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		fmt.Println("Bot is ready!")
	})
	dg.AddHandler(h.InteractionHandler)

	if err = dg.Open(); err != nil {
		log.Fatal("Error opening connection:", err)
	}
	defer dg.Close()

	fmt.Println("Bot is running...")
	registerCommands(dg)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	fmt.Println("Shutting down...")
}

func registerCommands(s *discordgo.Session) {
	guildID := os.Getenv("GUILD_ID")
	if guildID == "" {
		log.Fatal("GUILD_ID is empty")
	}

	commands := []*discordgo.ApplicationCommand{
		{Name: "ping", Description: "Ping test"},
		{Name: "release-init", Description: "Initialize a new release"},
		{Name: "release-add", Description: "Add your branch to the current release"},
		{Name: "release-update", Description: "Update one of your branches in the current release"},
		{Name: "release-summary", Description: "Show quick summary of current release"},
	}

	for _, cmd := range commands {
		if _, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd); err != nil {
			log.Fatalf("Cannot create %q command: %v", cmd.Name, err)
		}
	}
}