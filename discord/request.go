package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const defaultBaseURL = "https://discord.com/api/v10/"

type APIClient struct {
	client  *http.Client
	baseURL string
	token   string
}

type Option func(c *APIClient)

func WithBaseURL(baseURL string) Option {
	return func(c *APIClient) {
		c.baseURL = baseURL
	}
}

func NewAPIClient(token string, opts ...Option) *APIClient {
	c := &APIClient{
		client:  &http.Client{},
		baseURL: defaultBaseURL,
		token:   token,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *APIClient) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", c.token))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User-Agent", "DiscordBot (https://github.com/brazostech/discord, 1.0.0)")

	return c.client.Do(req)
}

type Method string

const (
	MethodGet    Method = http.MethodGet
	MethodDelete Method = http.MethodDelete
	MethodPut    Method = http.MethodPut
	MethodPost   Method = http.MethodPost
)

type RequestOptions struct {
	Method Method
	Body   any
}

func (c *APIClient) Request(ctx context.Context, endpoint string, options RequestOptions) (*http.Response, error) {
	var body io.Reader
	if options.Body != nil {
		payload, err := json.Marshal(options.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		body = bytes.NewReader(payload)
	}

	path, err := url.JoinPath(c.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("join request path: %w", err)
	}

	if options.Method == "" {
		options.Method = MethodGet
	}

	req, err := http.NewRequestWithContext(ctx, string(options.Method), path, body)
	if err != nil {
		return nil, fmt.Errorf("construct request: %w", err)
	}

	log.Printf("sending %s request to %s", options.Method, path)

	return c.do(req)
}

const maxErrorBodySize = 1 << 12

// Execute sends the request and reports an error for any non-2xx response,
// including the response body.
func (c *APIClient) Execute(ctx context.Context, endpoint string, options RequestOptions) error {
	resp, err := c.Request(ctx, endpoint, options)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return fmt.Errorf("discord returned %s: %s", resp.Status, body)
	}

	return nil
}
