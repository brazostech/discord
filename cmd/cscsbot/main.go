package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/brazostech/discord/bot"
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

	b, err := bot.NewDiscordBot(config.discordPublicKey)
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
