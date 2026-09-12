package discord

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequest(t *testing.T) {
	expectedRequestURL := "/foobar"
	expectedMethod := http.MethodPut
	expectedAuth := "Bot test-token"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != expectedRequestURL {
			t.Fatalf("unexpected path: want %s, got %s", expectedRequestURL, r.URL.String())
		}
		if r.Method != expectedMethod {
			t.Fatalf("unexpected method: want %s, got %s", expectedMethod, r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != expectedAuth {
			t.Fatalf("unexpected authorization header: want %q, got %q", expectedAuth, auth)
		}
	}))

	defer ts.Close()

	client := NewAPIClient("test-token", WithBaseURL(ts.URL))
	body := map[string]any{
		"foo": "bar",
	}

	res, err := client.Request(t.Context(), "/foobar", RequestOptions{
		Method: MethodPut,
		Body:   body,
	})
	if err != nil {
		t.Fatalf("unexpected error on request: %v", err)
	}
	defer res.Body.Close()
}
