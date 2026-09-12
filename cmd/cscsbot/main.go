package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/brazostech/discord/book"
	"github.com/brazostech/discord/discord"
	"github.com/brazostech/discord/discord/gateway"
	"github.com/brazostech/discord/discord/interactions"
)

type config struct {
	discordToken string
}

func parseConfig() (*config, error) {
	c := &config{}

	flag.StringVar(&c.discordToken, "token", "", "discord bot token")
	flag.Parse()

	if c.discordToken == "" {
		return nil, errors.New("discord token is required")
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

	// ----- build inbound dispatch -----
	dispatcher := interactions.NewDispatcher()
	dispatcher.Subscribe("test", interactions.CommandTestHandler)
	dispatcher.Subscribe("book register", commands.HandleRegister)
	dispatcher.Subscribe("book update-chapter", commands.HandleUpdateChapter)

	// ----- build gateway transport -----
	api := discord.NewAPIClient(config.discordToken)
	client := gateway.NewClient(config.discordToken, dispatcher, gateway.NewCallback(api))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := client.Run(ctx); err != nil {
		return err
	}

	log.Println("shutdown complete")
	return nil
}
