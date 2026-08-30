package bot

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	maxRequestBodySize = 1 << 20
	maxTimestampSkew   = 5 * time.Minute
)

func (b *DiscordBot) VerifyKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sig, err := hex.DecodeString(r.Header.Get("X-Signature-Ed25519"))
		if err != nil || len(sig) != ed25519.SignatureSize {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ts := r.Header.Get("X-Signature-Timestamp")
		tsInt, err := strconv.ParseInt(ts, 10, 64)
		if err != nil || time.Since(time.Unix(tsInt, 0)).Abs() > maxTimestampSkew {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		msg := append([]byte(ts), body...)
		if !ed25519.Verify(b.publicKey, msg, sig) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}
