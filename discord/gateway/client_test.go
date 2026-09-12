package gateway

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/brazostech/discord/book"
	"github.com/brazostech/discord/discord"
	"github.com/brazostech/discord/discord/interactions"
)

func TestClientIdentifiesWithNoIntents(t *testing.T) {
	client, dialer, _ := newTestClient(t)
	runClient(t, client)

	conn := dialer.connection(t, 0)
	conn.send(t, helloEvent(45000))

	identify := conn.nextWrite(t)
	if identify["op"] != float64(opIdentify) {
		t.Fatalf("op = %v, want identify", identify["op"])
	}

	data, ok := identify["d"].(map[string]any)
	if !ok {
		t.Fatalf("identify data = %#v", identify["d"])
	}
	if data["token"] != "test-token" {
		t.Fatalf("identify token = %v", data["token"])
	}
	if data["intents"] != float64(0) {
		t.Fatalf("identify intents = %v, want 0", data["intents"])
	}
	if _, ok := data["properties"].(map[string]any); !ok {
		t.Fatalf("identify properties = %#v", data["properties"])
	}
}

func TestClientDispatchesGatewayInteractions(t *testing.T) {
	dispatcher := interactions.NewDispatcher()
	dispatcher.Subscribe("test", interactions.CommandTestHandler)

	commands := book.NewCommands(book.NewService(book.NewMemoryStore()))
	dispatcher.Subscribe("book register", commands.HandleRegister)
	dispatcher.Subscribe("book update-chapter", commands.HandleUpdateChapter)

	dialer := &fakeDialer{}
	responder := newFakeResponder()

	client := NewClient("test-token", dispatcher, responder)
	client.gatewayURL = "wss://gateway.test"
	client.dialer = dialer
	client.reconnectDelay = 0
	client.jitter = func() float64 { return 1 }

	runClient(t, client)

	conn := dialer.connection(t, 0)
	conn.send(t, helloEvent(45000))
	conn.nextWrite(t) // identify

	conn.send(t, dispatchEvent(1, "INTERACTION_CREATE", map[string]any{
		"id": "1", "token": "token-1", "type": interactions.ApplicationCommandInteractionType,
		"guild_id": "server-1",
		"data":     map[string]any{"id": "100", "name": "test", "type": 1},
	}))

	assertContent := func(response capturedResponse, id, content string) {
		t.Helper()

		if response.interaction.ID != id {
			t.Fatalf("interaction id = %s, want %s", response.interaction.ID, id)
		}
		if response.response.Type != interactions.ChannelMessageWithSourceInteractionResponseType {
			t.Fatalf("interaction %s response type = %d", id, response.response.Type)
		}
		if response.response.Data == nil || response.response.Data.Content != content {
			t.Fatalf("interaction %s content = %+v, want %q", id, response.response.Data, content)
		}
	}

	assertContent(responder.next(t), "1", "test succeeded")

	conn.send(t, dispatchEvent(2, "INTERACTION_CREATE", map[string]any{
		"id": "2", "token": "token-2", "type": interactions.ApplicationCommandInteractionType,
		"guild_id": "server-1",
		"data": map[string]any{
			"id": "101", "name": "book", "type": 1,
			"options": []any{
				map[string]any{
					"name": "register", "type": discord.SubCommandOptionType,
					"options": []any{
						map[string]any{"name": "name", "type": discord.StringOptionType, "value": "Dune"},
					},
				},
			},
		},
	}))

	assertContent(responder.next(t), "2", "Now reading **Dune**, starting at chapter 0.")

	conn.send(t, dispatchEvent(3, "INTERACTION_CREATE", map[string]any{
		"id": "3", "token": "token-3", "type": interactions.ApplicationCommandInteractionType,
		"guild_id": "server-1",
		"data": map[string]any{
			"id": "101", "name": "book", "type": 1,
			"options": []any{
				map[string]any{
					"name": "update-chapter", "type": discord.SubCommandOptionType,
					"options": []any{
						map[string]any{"name": "chapter", "type": discord.IntegerOptionType, "value": 7},
					},
				},
			},
		},
	}))

	assertContent(responder.next(t), "3", "**Dune** is now on chapter 7.")

	conn.send(t, dispatchEvent(4, "INTERACTION_CREATE", map[string]any{
		"id": "4", "token": "token-4", "type": interactions.PingInteractionType,
	}))

	if ping := responder.next(t).response; ping.Type != interactions.PongInteractionResponseType || ping.Data != nil {
		t.Fatalf("ping response = %+v, want pong", ping)
	}
}

func TestClientHeartbeatsAtIntervalAndTracksAck(t *testing.T) {
	client, dialer, _ := newTestClient(t)
	runClient(t, client)

	conn := dialer.connection(t, 0)
	conn.send(t, helloEvent(20))
	conn.nextWrite(t) // identify

	first := conn.nextWrite(t)
	if first["op"] != float64(opHeartbeat) {
		t.Fatalf("op = %v, want heartbeat", first["op"])
	}
	if first["d"] != nil {
		t.Fatalf("first heartbeat d = %v, want null", first["d"])
	}

	conn.send(t, dispatchEvent(7, "READY", map[string]any{}))
	conn.send(t, heartbeatAck())

	second := conn.nextWrite(t)
	if second["op"] != float64(opHeartbeat) {
		t.Fatalf("op = %v, want heartbeat", second["op"])
	}
	if second["d"] != float64(7) {
		t.Fatalf("second heartbeat d = %v, want the last sequence 7", second["d"])
	}

	conn.send(t, heartbeatAck())

	third := conn.nextWrite(t)
	if third["op"] != float64(opHeartbeat) {
		t.Fatalf("op = %v, want heartbeat", third["op"])
	}

	if count := dialer.count(); count != 1 {
		t.Fatalf("dialer opened %d connections, want 1", count)
	}
}

func TestClientAnswersServerHeartbeatRequest(t *testing.T) {
	client, dialer, _ := newTestClient(t)
	client.jitter = func() float64 { return 1 }
	runClient(t, client)

	conn := dialer.connection(t, 0)
	conn.send(t, helloEvent(45000))
	conn.nextWrite(t) // identify

	conn.send(t, map[string]any{"op": opHeartbeat})

	heartbeat := conn.nextWrite(t)
	if heartbeat["op"] != float64(opHeartbeat) {
		t.Fatalf("op = %v, want heartbeat", heartbeat["op"])
	}
}

func TestClientReconnectsWhenHeartbeatUnacked(t *testing.T) {
	client, dialer, _ := newTestClient(t)
	runClient(t, client)

	conn := dialer.connection(t, 0)
	conn.send(t, helloEvent(20))
	conn.nextWrite(t) // identify
	conn.nextWrite(t) // first heartbeat, deliberately left unacknowledged

	dialer.connection(t, 1) // waits for the reconnect

	select {
	case <-conn.closed:
	case <-time.After(testWaitTimeout):
		t.Fatal("stale connection was not closed")
	}

	if code := conn.closeCode.Load(); code == 1000 || code == 1001 {
		t.Fatalf("close code = %d, want a non-normal code", code)
	}
}

func TestClientReconnectsWhenConnectionCloses(t *testing.T) {
	client, dialer, _ := newTestClient(t)
	runClient(t, client)

	conn := dialer.connection(t, 0)
	conn.send(t, helloEvent(45000))
	conn.nextWrite(t) // identify

	conn.Close(1000, "closed by discord")

	assertReidentifies(t, dialer)
}

func TestClientReconnectsOnInvalidSession(t *testing.T) {
	client, dialer, _ := newTestClient(t)
	runClient(t, client)

	conn := dialer.connection(t, 0)
	conn.send(t, helloEvent(45000))
	conn.nextWrite(t) // identify

	conn.send(t, map[string]any{"op": opInvalidSession, "d": false})

	assertReidentifies(t, dialer)
}

func TestClientReconnectsOnReconnectOpcode(t *testing.T) {
	client, dialer, _ := newTestClient(t)
	runClient(t, client)

	conn := dialer.connection(t, 0)
	conn.send(t, helloEvent(45000))
	conn.nextWrite(t) // identify

	conn.send(t, map[string]any{"op": opReconnect, "d": nil})

	assertReidentifies(t, dialer)
}

func TestClientStopsReconnectingOnFatalCloseCode(t *testing.T) {
	client, dialer, _ := newTestClient(t)
	_, done := startClient(t, client)

	conn := dialer.connection(t, 0)
	conn.send(t, helloEvent(45000))
	conn.nextWrite(t) // identify

	conn.fail(&closeError{Code: 4004, Reason: "Authentication failed."})

	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "4004") {
			t.Fatalf("run error = %v, want the fatal close code", err)
		}
	case <-time.After(testWaitTimeout):
		t.Fatal("client kept reconnecting after a fatal close code")
	}

	if count := dialer.count(); count != 1 {
		t.Fatalf("dialer opened %d connections, want 1", count)
	}
}

func TestClientFinishesCallbacksAcrossReconnects(t *testing.T) {
	dispatcher := interactions.NewDispatcher()
	dispatcher.Subscribe("test", interactions.CommandTestHandler)

	dialer := &fakeDialer{}
	responder := &blockingResponder{
		entered:  make(chan struct{}),
		release:  make(chan struct{}),
		finished: make(chan error, 1),
	}

	client := NewClient("test-token", dispatcher, responder)
	client.gatewayURL = "wss://gateway.test"
	client.dialer = dialer
	client.reconnectDelay = 0
	client.jitter = func() float64 { return 1 }

	runClient(t, client)

	conn := dialer.connection(t, 0)
	conn.send(t, helloEvent(45000))
	conn.nextWrite(t) // identify

	conn.send(t, dispatchEvent(1, "INTERACTION_CREATE", map[string]any{
		"id": "1", "token": "token-1", "type": interactions.ApplicationCommandInteractionType,
		"guild_id": "server-1",
		"data":     map[string]any{"id": "100", "name": "test", "type": 1},
	}))

	<-responder.entered
	conn.Close(1000, "reconnecting")

	dialer.connection(t, 1) // the reconnect must not abort the callback

	close(responder.release)

	if err := <-responder.finished; err != nil {
		t.Fatalf("callback aborted across reconnect: %v", err)
	}
}

type blockingResponder struct {
	entered  chan struct{}
	release  chan struct{}
	finished chan error
}

func (b *blockingResponder) Respond(ctx context.Context, _ interactions.Interaction, _ interactions.InteractionResponse) error {
	b.entered <- struct{}{}
	<-b.release
	b.finished <- ctx.Err()

	return nil
}
