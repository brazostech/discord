package interactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/brazostech/discord/discord"
	"github.com/brazostech/discord/utils"
)

var ErrHandlerNotFound = errors.New("handler not found")

type InteractionsHandler func(context.Context, InteractionPacket) (InteractionResponse, error)

type InteractionsTypeRouter interface {
	Route(context.Context, http.ResponseWriter, *http.Request, InteractionPacket) error
}

type InteractionsRouter struct {
	subscriptions map[InteractionType]InteractionsTypeRouter
}

type InteractionPacket struct {
	Interaction Interaction
	Data        *discord.ApplicationCommandData
}

func NewInteractionsRouter(subscriptions map[InteractionType]InteractionsTypeRouter) *InteractionsRouter {
	return &InteractionsRouter{
		subscriptions: subscriptions,
	}
}

func (ir *InteractionsRouter) Route(ctx context.Context, w http.ResponseWriter, r *http.Request, packet InteractionPacket) error {
	subrouter, ok := ir.subscriptions[packet.Interaction.Type]
	if !ok {
		return fmt.Errorf("%w: interaction type %d", ErrHandlerNotFound, packet.Interaction.Type)
	}

	return subrouter.Route(ctx, w, r, packet)
}

// ---------- Ping Router ----------

type PingRouter struct{}

func NewPingRouter() *PingRouter {
	return &PingRouter{}
}

func (router *PingRouter) Route(_ context.Context, w http.ResponseWriter, _ *http.Request, _ InteractionPacket) error {
	return utils.WriteJSON(w, http.StatusOK, InteractionResponse{Type: PongInteractionResponseType})
}

// ---------- Command Router ----------

type CommandRouter struct {
	subscriptions map[string]InteractionsHandler
}

func NewCommandRouter() *CommandRouter {
	return &CommandRouter{
		subscriptions: make(map[string]InteractionsHandler),
	}
}

func (router *CommandRouter) Route(ctx context.Context, w http.ResponseWriter, r *http.Request, packet InteractionPacket) error {
	i := packet.Interaction
	if i.Type != ApplicationCommandInteractionType && i.Type != ApplicationCommandAutocompleteInteractionType {
		return fmt.Errorf("interaction type %d does not contain application command data", i.Type)
	}

	var data discord.ApplicationCommandData
	if err := json.Unmarshal(i.Data, &data); err != nil {
		return fmt.Errorf("unmarshal application command data: %w", err)
	}

	path := data.Name
	if subcommand, ok := data.Subcommand(); ok {
		path += " " + subcommand
	}

	fn, ok := router.subscriptions[path]
	if !ok {
		return fmt.Errorf("%w: command %q", ErrHandlerNotFound, path)
	}

	packet.Data = &data
	res, err := fn(ctx, packet)
	if err != nil {
		return fmt.Errorf("command handler %q: %w", path, err)
	}

	return utils.WriteJSON(w, http.StatusOK, res)
}

// Subscribe registers a handler for a command path, e.g. "book register".
func (r *CommandRouter) Subscribe(path string, fn InteractionsHandler) {
	if path == "" {
		panic("interaction command path must not be empty")
	}
	if fn == nil {
		panic("interaction handler must not be nil")
	}

	r.subscriptions[path] = fn
}
