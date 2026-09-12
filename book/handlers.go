package book

import (
	"context"
	"errors"
	"fmt"

	"github.com/brazostech/discord/discord"
	"github.com/brazostech/discord/discord/interactions"
)

// Commands adapts Discord command interactions to the Book domain.
type Commands struct {
	service *Service
}

func NewCommands(service *Service) *Commands {
	return &Commands{service: service}
}

// HandleRegister handles /book register.
func (c *Commands) HandleRegister(ctx context.Context, packet interactions.InteractionPacket) (interactions.InteractionResponse, error) {
	opts, err := subcommandOptions(packet)
	if err != nil {
		return interactions.InteractionResponse{}, err
	}

	name, _ := stringOption(opts, "name")
	url, _ := stringOption(opts, "url")

	registered, err := c.service.Register(ctx, packet.Interaction.GuildID, name, url)
	switch {
	case errors.Is(err, ErrNameRequired):
		return message("A book name is required."), nil
	case err != nil:
		return interactions.InteractionResponse{}, fmt.Errorf("register book: %w", err)
	}

	return message(fmt.Sprintf("Now reading **%s**, starting at chapter 0.", registered.Name)), nil
}

// HandleUpdateChapter handles /book update-chapter.
func (c *Commands) HandleUpdateChapter(ctx context.Context, packet interactions.InteractionPacket) (interactions.InteractionResponse, error) {
	opts, err := subcommandOptions(packet)
	if err != nil {
		return interactions.InteractionResponse{}, err
	}

	chapter, ok := intOption(opts, "chapter")
	if !ok {
		return interactions.InteractionResponse{}, errors.New("chapter option is missing or not an integer")
	}

	updated, err := c.service.UpdateChapter(ctx, packet.Interaction.GuildID, chapter)
	switch {
	case errors.Is(err, ErrNoCurrentBook):
		return message("No book registered yet — use `/book register`."), nil
	case errors.Is(err, ErrNegativeChapter):
		return message("The chapter must be 0 or greater."), nil
	case err != nil:
		return interactions.InteractionResponse{}, fmt.Errorf("update chapter: %w", err)
	}

	return message(fmt.Sprintf("**%s** is now on chapter %d.", updated.Name, updated.Chapter)), nil
}

func subcommandOptions(packet interactions.InteractionPacket) ([]discord.InteractionOption, error) {
	if packet.Data == nil {
		return nil, errors.New("missing application command data")
	}

	return packet.Data.SubcommandOptions(), nil
}

func message(content string) interactions.InteractionResponse {
	return interactions.InteractionResponse{
		Type: interactions.ChannelMessageWithSourceInteractionResponseType,
		Data: &interactions.InteractionResponseData{Content: content},
	}
}

func stringOption(opts []discord.InteractionOption, name string) (string, bool) {
	opt, ok := findOption(opts, name)
	if !ok {
		return "", false
	}

	return opt.StringValue()
}

func intOption(opts []discord.InteractionOption, name string) (int, bool) {
	opt, ok := findOption(opts, name)
	if !ok {
		return 0, false
	}

	return opt.IntValue()
}

func findOption(opts []discord.InteractionOption, name string) (discord.InteractionOption, bool) {
	for _, opt := range opts {
		if opt.Name == name {
			return opt, true
		}
	}

	return discord.InteractionOption{}, false
}
