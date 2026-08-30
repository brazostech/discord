package bot

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/brazostech/discord/interactions"
)

const (
	addr            = ":8080"
	shutdownTimeout = 5 * time.Second
)

type DiscordBot struct {
	server    *http.Server
	publicKey ed25519.PublicKey
	router    interactions.InteractionsTypeRouter
}

func NewDiscordBot(pubKeyHex string, router interactions.InteractionsTypeRouter) (*DiscordBot, error) {
	pubKey, err := hex.DecodeString(pubKeyHex)
	if err != nil || len(pubKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid discord public key: %w", err)
	}

	return &DiscordBot{
		publicKey: pubKey,
		router:    router,
	}, nil
}

func (d *DiscordBot) InteractionsHandler(w http.ResponseWriter, r *http.Request) {
	var interaction interactions.Interaction
	if err := json.NewDecoder(r.Body).Decode(&interaction); err != nil {
		log.Printf("decode interaction: %v", err)
		http.Error(w, "invalid interaction payload", http.StatusBadRequest)
		return
	}

	packet := interactions.InteractionPacket{
		Interaction: interaction,
	}

	if err := d.router.Route(r.Context(), w, r, packet); err != nil {
		log.Printf("route interaction: %v", err)
		http.Error(w, "failed to process interaction", http.StatusBadRequest)
	}
}

func (d *DiscordBot) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/interactions", d.InteractionsHandler)

	d.server = &http.Server{
		Addr:    addr,
		Handler: d.VerifyKeyMiddleware(mux),
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", d.server.Addr)
		if err := d.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()
		return d.server.Shutdown(shutdownCtx)
	}
}
