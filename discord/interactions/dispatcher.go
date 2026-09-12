package interactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/brazostech/discord/discord"
)

var ErrHandlerNotFound = errors.New("handler not found")

// CommandHandler responds to a parsed application command invocation.
type CommandHandler func(context.Context, Invocation) (InteractionResponse, error)

// Invocation is an application command invocation: the interaction that
// triggered it and the options the invoking user submitted. When the command
// uses subcommands, Options are the subcommand's options.
type Invocation struct {
	Interaction Interaction
	Options     []discord.InteractionOption
}

// Dispatcher resolves inbound interactions to responses. Register command
// handlers with Subscribe; unregistered commands and unsupported interaction
// types return ErrHandlerNotFound.
type Dispatcher struct {
	commands map[string]CommandHandler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		commands: make(map[string]CommandHandler),
	}
}

// Handle resolves one inbound interaction, returning the response to send.
func (d *Dispatcher) Handle(ctx context.Context, interaction Interaction) (InteractionResponse, error) {
	switch interaction.Type {
	case PingInteractionType:
		return InteractionResponse{Type: PongInteractionResponseType}, nil
	case ApplicationCommandInteractionType, ApplicationCommandAutocompleteInteractionType:
		return d.dispatchCommand(ctx, interaction)
	default:
		return InteractionResponse{}, fmt.Errorf("%w: interaction type %d", ErrHandlerNotFound, interaction.Type)
	}
}

func (d *Dispatcher) dispatchCommand(ctx context.Context, interaction Interaction) (InteractionResponse, error) {
	var data discord.ApplicationCommandData
	if err := json.Unmarshal(interaction.Data, &data); err != nil {
		return InteractionResponse{}, fmt.Errorf("unmarshal application command data: %w", err)
	}

	path := data.Name
	if subcommand, ok := data.Subcommand(); ok {
		path += " " + subcommand
	}

	fn, ok := d.commands[path]
	if !ok {
		return InteractionResponse{}, fmt.Errorf("%w: command %q", ErrHandlerNotFound, path)
	}

	return fn(ctx, Invocation{
		Interaction: interaction,
		Options:     data.SubcommandOptions(),
	})
}

// Subscribe registers a handler for a command path, e.g. "book register".
func (d *Dispatcher) Subscribe(path string, fn CommandHandler) {
	if path == "" {
		panic("interaction command path must not be empty")
	}
	if fn == nil {
		panic("interaction handler must not be nil")
	}

	d.commands[path] = fn
}
