package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/brazostech/discord/discord/interactions"
)

const (
	defaultGatewayURL     = "wss://gateway.discord.gg/?v=10&encoding=json"
	defaultReconnectDelay = time.Second
	reconnectCloseCode    = 4000
	identifyIntents       = 0
)

const (
	opDispatch       = 0
	opHeartbeat      = 1
	opIdentify       = 2
	opReconnect      = 7
	opInvalidSession = 9
	opHello          = 10
	opHeartbeatAck   = 11
)

var (
	errHeartbeatUnacked = errors.New("gateway heartbeat was not acknowledged")
	errReconnect        = errors.New("gateway requested a reconnect")
	errInvalidSession   = errors.New("gateway invalidated the session")
)

// Dispatcher resolves an inbound interaction to the response to send.
type Dispatcher interface {
	Handle(ctx context.Context, interaction interactions.Interaction) (interactions.InteractionResponse, error)
}

// Responder delivers an interaction response back to Discord.
type Responder interface {
	Respond(ctx context.Context, interaction interactions.Interaction, response interactions.InteractionResponse) error
}

type Client struct {
	token          string
	dispatcher     Dispatcher
	responder      Responder
	gatewayURL     string
	dialer         dialer
	jitter         func() float64
	reconnectDelay time.Duration
}

func NewClient(token string, dispatcher Dispatcher, responder Responder) *Client {
	return &Client{
		token:          token,
		dispatcher:     dispatcher,
		responder:      responder,
		gatewayURL:     defaultGatewayURL,
		dialer:         websocketDialer{},
		jitter:         rand.Float64,
		reconnectDelay: defaultReconnectDelay,
	}
}

// Run keeps a gateway session alive until ctx is done, reconnecting after
// closes, reconnect requests, and invalid sessions.
func (c *Client) Run(ctx context.Context) error {
	for {
		if err := c.runSession(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("discord gateway: %v", err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.reconnectDelay):
		}
	}
}

func (c *Client) runSession(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := c.dialer.dial(ctx, c.gatewayURL)
	if err != nil {
		return fmt.Errorf("dial gateway: %w", err)
	}
	defer conn.Close(reconnectCloseCode, "reconnecting")

	incoming, readErrors := readMessages(ctx, conn)

	interval, err := c.awaitHello(ctx, incoming, readErrors)
	if err != nil {
		return err
	}

	if err := c.send(ctx, conn, outbound{
		Op: opIdentify,
		Data: identifyData{
			Token:   c.token,
			Intents: identifyIntents,
			Properties: identifyProperties{
				OS:      runtime.GOOS,
				Browser: "github.com/brazostech/discord",
				Device:  "github.com/brazostech/discord",
			},
		},
	}); err != nil {
		return fmt.Errorf("identify: %w", err)
	}

	return c.serve(ctx, conn, incoming, readErrors, interval)
}

func (c *Client) serve(
	ctx context.Context,
	conn conn,
	incoming <-chan []byte,
	readErrors <-chan error,
	interval time.Duration,
) error {
	var (
		sequence      *int64
		heartbeatSent bool
	)

	timer := time.NewTimer(c.heartbeatDelay(interval))
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-readErrors:
			return fmt.Errorf("read gateway: %w", err)
		case <-timer.C:
			if heartbeatSent {
				return errHeartbeatUnacked
			}
			if err := c.sendHeartbeat(ctx, conn, sequence); err != nil {
				return err
			}
			heartbeatSent = true
			timer.Reset(interval)
		case data := <-incoming:
			p, err := decodePayload(data)
			if err != nil {
				log.Printf("discord gateway: %v", err)
				continue
			}

			switch p.Op {
			case opDispatch:
				if p.Sequence != nil {
					sequence = p.Sequence
				}
				if p.Type == "INTERACTION_CREATE" {
					go c.handleInteraction(ctx, p.Data)
				}
			case opHeartbeat:
				if err := c.sendHeartbeat(ctx, conn, sequence); err != nil {
					return err
				}
			case opHeartbeatAck:
				heartbeatSent = false
			case opReconnect:
				return errReconnect
			case opInvalidSession:
				return errInvalidSession
			}
		}
	}
}

func (c *Client) awaitHello(ctx context.Context, incoming <-chan []byte, readErrors <-chan error) (time.Duration, error) {
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case err := <-readErrors:
			return 0, fmt.Errorf("read gateway: %w", err)
		case data := <-incoming:
			p, err := decodePayload(data)
			if err != nil {
				return 0, err
			}

			switch p.Op {
			case opHello:
				var hello helloData
				if err := json.Unmarshal(p.Data, &hello); err != nil {
					return 0, fmt.Errorf("decode hello: %w", err)
				}
				if hello.HeartbeatInterval <= 0 {
					return 0, fmt.Errorf("hello: invalid heartbeat interval %d", hello.HeartbeatInterval)
				}

				return time.Duration(hello.HeartbeatInterval) * time.Millisecond, nil
			case opReconnect:
				return 0, errReconnect
			case opInvalidSession:
				return 0, errInvalidSession
			}
		}
	}
}

func (c *Client) sendHeartbeat(ctx context.Context, conn conn, sequence *int64) error {
	if err := c.send(ctx, conn, outbound{Op: opHeartbeat, Data: sequence}); err != nil {
		return fmt.Errorf("heartbeat: %w", err)
	}

	return nil
}

func (c *Client) send(ctx context.Context, conn conn, payload outbound) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode gateway payload: %w", err)
	}

	if err := conn.Write(ctx, data); err != nil {
		return fmt.Errorf("write gateway payload: %w", err)
	}

	return nil
}

func (c *Client) handleInteraction(ctx context.Context, data json.RawMessage) {
	var interaction interactions.Interaction
	if err := json.Unmarshal(data, &interaction); err != nil {
		log.Printf("discord gateway: decode interaction: %v", err)
		return
	}

	response, err := c.dispatcher.Handle(ctx, interaction)
	if err != nil {
		log.Printf("discord gateway: handle interaction %s: %v", interaction.ID, err)
		return
	}

	if err := c.responder.Respond(ctx, interaction, response); err != nil {
		log.Printf("discord gateway: respond to interaction %s: %v", interaction.ID, err)
	}
}

func (c *Client) heartbeatDelay(interval time.Duration) time.Duration {
	return time.Duration(float64(interval) * c.jitter())
}

func readMessages(ctx context.Context, conn conn) (<-chan []byte, <-chan error) {
	messages := make(chan []byte)
	errs := make(chan error, 1)

	go func() {
		for {
			data, err := conn.Read(ctx)
			if err != nil {
				errs <- err
				return
			}

			select {
			case messages <- data:
			case <-ctx.Done():
				return
			}
		}
	}()

	return messages, errs
}

type payload struct {
	Op       int             `json:"op"`
	Data     json.RawMessage `json:"d"`
	Sequence *int64          `json:"s"`
	Type     string          `json:"t"`
}

func decodePayload(data []byte) (payload, error) {
	var p payload
	if err := json.Unmarshal(data, &p); err != nil {
		return payload{}, fmt.Errorf("decode gateway payload: %w", err)
	}

	return p, nil
}

type outbound struct {
	Op   int `json:"op"`
	Data any `json:"d"`
}

type helloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type identifyData struct {
	Token      string             `json:"token"`
	Intents    int                `json:"intents"`
	Properties identifyProperties `json:"properties"`
}

type identifyProperties struct {
	OS      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}
