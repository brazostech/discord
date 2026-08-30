package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/brazostech/discord"
	"github.com/brazostech/discord/commands"
)

const commandRequestTimeout = 5 * time.Second

type config struct {
	discordToken string
	appID        string
}

func parseConfig() (*config, error) {
	c := &config{}

	flag.StringVar(&c.discordToken, "token", "", "discord bot token")
	flag.StringVar(&c.appID, "appid", "", "discord application id")
	flag.Parse()

	if c.discordToken == "" {
		return nil, errors.New("discord token is required")
	}
	if c.appID == "" {
		return nil, errors.New("discord application id is required")
	}

	return c, nil
}

func main() {
	config, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	allCommands := []commands.Command{
		commands.SimpleTestCommand,
		commands.BookCommand,
	}

	client := discord.NewAPIClient(config.discordToken)

	ctx, cancel := context.WithTimeout(context.Background(), commandRequestTimeout)
	defer cancel()

	if err := installCommands(ctx, client, config.appID, allCommands); err != nil {
		log.Fatalf("failed to install commands: %v", err)
	}
}

func installCommands(ctx context.Context, client *discord.APIClient, appID string, cmds []commands.Command) error {
	endpoint := fmt.Sprintf("applications/%s/commands", appID)

	resp, err := client.Request(ctx, endpoint, discord.RequestOptions{
		Method: discord.MethodPut,
		Body:   cmds,
	})
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord returned %s: %s", resp.Status, body)
	}

	log.Printf("installed %d commands", len(cmds))

	return nil
}
