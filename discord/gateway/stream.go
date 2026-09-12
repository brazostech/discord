package gateway

import "context"

// stream carries the payloads read from one gateway connection.
type stream struct {
	messages <-chan []byte
	errors   <-chan error
}

func newStream(ctx context.Context, conn conn) *stream {
	messages := make(chan []byte)
	errors := make(chan error, 1)

	go func() {
		for {
			data, err := conn.Read(ctx)
			if err != nil {
				errors <- err
				return
			}

			select {
			case messages <- data:
			case <-ctx.Done():
				return
			}
		}
	}()

	return &stream{messages: messages, errors: errors}
}
