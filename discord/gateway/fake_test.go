package gateway

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/brazostech/discord/discord/interactions"
)

const testWaitTimeout = 2 * time.Second

type fakeConn struct {
	inbound   chan []byte
	writes    chan []byte
	closed    chan struct{}
	closeOnce sync.Once
	closeCode atomic.Int64
}

func newFakeConn() *fakeConn {
	return &fakeConn{
		inbound: make(chan []byte, 16),
		writes:  make(chan []byte, 16),
		closed:  make(chan struct{}),
	}
}

func (f *fakeConn) Read(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.closed:
		return nil, net.ErrClosed
	case message := <-f.inbound:
		return message, nil
	}
}

func (f *fakeConn) Write(ctx context.Context, data []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-f.closed:
		return net.ErrClosed
	case f.writes <- data:
		return nil
	}
}

func (f *fakeConn) Close(code int, _ string) error {
	f.closeCode.Store(int64(code))
	f.closeOnce.Do(func() { close(f.closed) })
	return nil
}

func (f *fakeConn) send(t *testing.T, payload any) {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal gateway payload: %v", err)
	}

	select {
	case f.inbound <- data:
	case <-time.After(testWaitTimeout):
		t.Fatal("timed out feeding the gateway stream")
	}
}

func (f *fakeConn) nextWrite(t *testing.T) map[string]any {
	t.Helper()

	select {
	case data := <-f.writes:
		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("decode gateway write %q: %v", data, err)
		}
		return payload
	case <-time.After(testWaitTimeout):
		t.Fatal("timed out waiting for a gateway write")
		return nil
	}
}

type fakeDialer struct {
	mu    sync.Mutex
	conns []*fakeConn
}

func (d *fakeDialer) dial(context.Context, string) (conn, error) {
	c := newFakeConn()

	d.mu.Lock()
	defer d.mu.Unlock()
	d.conns = append(d.conns, c)

	return c, nil
}

func (d *fakeDialer) connection(t *testing.T, index int) *fakeConn {
	t.Helper()

	deadline := time.Now().Add(testWaitTimeout)
	for time.Now().Before(deadline) {
		d.mu.Lock()
		if len(d.conns) > index {
			c := d.conns[index]
			d.mu.Unlock()
			return c
		}
		d.mu.Unlock()

		time.Sleep(time.Millisecond)
	}

	t.Fatalf("gateway connection %d was never dialed", index+1)
	return nil
}

func (d *fakeDialer) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return len(d.conns)
}

type interactionResponse struct {
	interaction interactions.Interaction
	response    interactions.InteractionResponse
}

type fakeResponder struct {
	responses chan interactionResponse
}

func newFakeResponder() *fakeResponder {
	return &fakeResponder{responses: make(chan interactionResponse, 16)}
}

func (f *fakeResponder) Respond(_ context.Context, interaction interactions.Interaction, response interactions.InteractionResponse) error {
	f.responses <- interactionResponse{interaction: interaction, response: response}

	return nil
}

func (f *fakeResponder) next(t *testing.T) interactionResponse {
	t.Helper()

	select {
	case response := <-f.responses:
		return response
	case <-time.After(testWaitTimeout):
		t.Fatal("timed out waiting for an interaction response")
		return interactionResponse{}
	}
}

func newTestClient(t *testing.T) (*Client, *fakeDialer, *fakeResponder) {
	t.Helper()

	dispatcher := interactions.NewDispatcher()
	dispatcher.Subscribe("test", interactions.CommandTestHandler)

	dialer := &fakeDialer{}
	responder := newFakeResponder()

	client := NewClient("test-token", dispatcher, responder)
	client.gatewayURL = "wss://gateway.test"
	client.dialer = dialer
	client.reconnectDelay = 0
	client.jitter = func() float64 { return 0 }

	return client, dialer, responder
}

func runClient(t *testing.T, client *Client) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		if err := client.Run(ctx); err != nil {
			t.Errorf("run client: %v", err)
		}
	}()

	t.Cleanup(func() {
		cancel()

		select {
		case <-done:
		case <-time.After(testWaitTimeout):
			t.Error("client did not stop after cancellation")
		}
	})
}

func helloEvent(intervalMS int) map[string]any {
	return map[string]any{
		"op": opHello,
		"d":  map[string]any{"heartbeat_interval": intervalMS},
	}
}

func dispatchEvent(sequence int, eventType string, data any) map[string]any {
	return map[string]any{
		"op": opDispatch,
		"s":  sequence,
		"t":  eventType,
		"d":  data,
	}
}

func heartbeatAck() map[string]any {
	return map[string]any{"op": opHeartbeatAck}
}
