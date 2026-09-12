package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/brazostech/discord/book"
	"github.com/brazostech/discord/bot"
	"github.com/brazostech/discord/discord/interactions"
)

type config struct {
	discordPublicKey string
}

func parseConfig() (*config, error) {
	c := &config{}

	flag.StringVar(&c.discordPublicKey, "pubkey", "", "discord application public key")
	flag.Parse()

	if c.discordPublicKey == "" {
		return nil, errors.New("discord public key is required")
	}

	return c, nil
}

func main() {
	if err := start(); err != nil {
		log.Fatal(err)
	}
}

func start() error {
	config, err := parseConfig()
	if err != nil {
		return err
	}

	// ----- build services -----
	store := book.NewMemoryStore()
	service := book.NewService(store)
	commands := book.NewCommands(service)

	// ----- build subrouters -----
	commandRouter := interactions.NewCommandRouter() // a command is type of interaction
	commandRouter.Subscribe("test", interactions.CommandTestHandler)
	commandRouter.Subscribe("book register", commands.HandleRegister)
	commandRouter.Subscribe("book update-chapter", commands.HandleUpdateChapter)

	// ----- builder top-level router -----
	router := interactions.NewInteractionsRouter(map[interactions.InteractionType]interactions.InteractionsTypeRouter{
		interactions.PingInteractionType:               interactions.NewPingRouter(),
		interactions.ApplicationCommandInteractionType: commandRouter,
	})

	b, err := bot.NewDiscordBot(config.discordPublicKey, router)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := b.Start(ctx); err != nil {
		return err
	}

	log.Println("shutdown complete")
	return nil
}
