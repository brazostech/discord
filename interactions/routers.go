package interactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/brazostech/discord/commands"
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
	Data        *commands.ApplicationCommandData
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
	return WriteJSON(w, http.StatusOK, InteractionResponse{Type: PongInteractionResponseType})
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

	var data commands.ApplicationCommandData
	if err := json.Unmarshal(i.Data, &data); err != nil {
		return fmt.Errorf("unmarshal application command data: %w", err)
	}

	fn, ok := router.subscriptions[data.Name]
	if !ok {
		return fmt.Errorf("%w: command %q", ErrHandlerNotFound, data.Name)
	}

	packet.Data = &data
	res, err := fn(ctx, packet)
	if err != nil {
		http.Error(w, "failed to handle interaction", http.StatusInternalServerError)
		return fmt.Errorf("command handler %q: %w", data.Name, err)
	}

	return WriteJSON(w, http.StatusOK, res)
}

func (r *CommandRouter) Subscribe(name string, fn InteractionsHandler) {
	if name == "" {
		panic("interaction command name must not be empty")
	}
	if fn == nil {
		panic("interaction handler must not be nil")
	}

	r.subscriptions[name] = fn
}
