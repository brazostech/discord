package gateway

import (
	"context"
	"fmt"

	"github.com/coder/websocket"
)

type conn interface {
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, data []byte) error
	Close(code int, reason string) error
}

type dialer interface {
	dial(ctx context.Context, url string) (conn, error)
}

// closeError reports the Discord gateway close code that ended a connection.
type closeError struct {
	Code   int
	Reason string
}

func (e *closeError) Error() string {
	return fmt.Sprintf("gateway closed with code %d: %s", e.Code, e.Reason)
}

type websocketDialer struct{}

func (websocketDialer) dial(ctx context.Context, url string) (conn, error) {
	ws, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	return websocketConn{conn: ws}, nil
}

type websocketConn struct {
	conn *websocket.Conn
}

func (c websocketConn) Read(ctx context.Context) ([]byte, error) {
	_, data, err := c.conn.Read(ctx)
	if err != nil {
		if code := websocket.CloseStatus(err); code != -1 {
			return nil, &closeError{Code: int(code), Reason: err.Error()}
		}

		return nil, err
	}

	return data, nil
}

func (c websocketConn) Write(ctx context.Context, data []byte) error {
	return c.conn.Write(ctx, websocket.MessageText, data)
}

func (c websocketConn) Close(code int, reason string) error {
	return c.conn.Close(websocket.StatusCode(code), reason)
}
