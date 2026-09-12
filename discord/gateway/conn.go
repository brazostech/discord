package gateway

import (
	"context"

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
	return data, err
}

func (c websocketConn) Write(ctx context.Context, data []byte) error {
	return c.conn.Write(ctx, websocket.MessageText, data)
}

func (c websocketConn) Close(code int, reason string) error {
	return c.conn.Close(websocket.StatusCode(code), reason)
}
