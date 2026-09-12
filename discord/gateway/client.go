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
// closes, reconnect requests, and invalid sessions. Close codes Discord marks
// as non-reconnectable end the run.
func (c *Client) Run(ctx context.Context) error {
	for {
		err := c.runSession(ctx)
		if err != nil {
			var closed *closeError
			if errors.As(err, &closed) && !reconnectable(closed.Code) {
				return fmt.Errorf("gateway session: %w", err)
			}
			if !errors.Is(err, context.Canceled) {
				log.Printf("discord gateway: %v", err)
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.reconnectDelay):
		}
	}
}

// session is one gateway connection: the socket, its read stream, the context
// that ends with the connection, and the long-lived context callbacks use.
type session struct {
	conn        conn
	stream      *stream
	ctx         context.Context
	dispatchCtx context.Context
	interval    time.Duration
}

func (c *Client) runSession(ctx context.Context) error {
	sessionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := c.dialer.dial(sessionCtx, c.gatewayURL)
	if err != nil {
		return fmt.Errorf("dial gateway: %w", err)
	}
	defer conn.Close(reconnectCloseCode, "reconnecting")

	s := &session{
		conn:        conn,
		stream:      newStream(sessionCtx, conn),
		ctx:         sessionCtx,
		dispatchCtx: ctx,
	}

	interval, err := c.awaitHello(s)
	if err != nil {
		return err
	}
	s.interval = interval

	if err := c.send(sessionCtx, conn, outbound{
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

	return c.serve(s)
}

func (c *Client) serve(s *session) error {
	var (
		sequence      *int64
		heartbeatSent bool
	)

	timer := time.NewTimer(c.heartbeatDelay(s.interval))
	defer timer.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case err := <-s.stream.errors:
			return fmt.Errorf("read gateway: %w", err)
		case <-timer.C:
			if heartbeatSent {
				return errHeartbeatUnacked
			}
			if err := c.sendHeartbeat(s.ctx, s.conn, sequence); err != nil {
				return err
			}
			heartbeatSent = true
			timer.Reset(s.interval)
		case data := <-s.stream.messages:
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
					go c.handleInteraction(s.dispatchCtx, p.Data)
				}
			case opHeartbeat:
				if err := c.sendHeartbeat(s.ctx, s.conn, sequence); err != nil {
					return err
				}
			case opHeartbeatAck:
				heartbeatSent = false
			default:
				if err := sessionEnd(p.Op); err != nil {
					return err
				}
			}
		}
	}
}

func (c *Client) awaitHello(s *session) (time.Duration, error) {
	for {
		select {
		case <-s.ctx.Done():
			return 0, s.ctx.Err()
		case err := <-s.stream.errors:
			return 0, fmt.Errorf("read gateway: %w", err)
		case data := <-s.stream.messages:
			p, err := decodePayload(data)
			if err != nil {
				return 0, err
			}
			if err := sessionEnd(p.Op); err != nil {
				return 0, err
			}
			if p.Op != opHello {
				continue
			}

			var hello helloData
			if err := json.Unmarshal(p.Data, &hello); err != nil {
				return 0, fmt.Errorf("decode hello: %w", err)
			}
			if hello.HeartbeatInterval <= 0 {
				return 0, fmt.Errorf("hello: invalid heartbeat interval %d", hello.HeartbeatInterval)
			}

			return time.Duration(hello.HeartbeatInterval) * time.Millisecond, nil
		}
	}
}

// sessionEnd maps the ops that end a session to their sentinel errors, whether
// they arrive while waiting for Hello or mid-session.
func sessionEnd(op int) error {
	switch op {
	case opReconnect:
		return errReconnect
	case opInvalidSession:
		return errInvalidSession
	default:
		return nil
	}
}

// reconnectable reports whether Discord expects clients to reconnect after a
// close code. A bad token or invalid identify must not loop forever.
func reconnectable(code int) bool {
	switch code {
	case 4004, 4010, 4011, 4012, 4013, 4014:
		return false
	default:
		return true
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

func (c *Client) heartbeatDelay(interval time.Duration) time.Duration {
	return time.Duration(float64(interval) * c.jitter())
}
