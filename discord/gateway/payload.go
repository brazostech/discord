package gateway

import (
	"encoding/json"
	"fmt"
)

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
