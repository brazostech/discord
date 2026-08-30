package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSON(
	w http.ResponseWriter,
	status int,
	value any,
) error {
	w.Header().Set("Content-Type", "application/json")

	if status == 0 {
		status = http.StatusOK
	}

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		return fmt.Errorf("encode JSON response: %w", err)
	}

	return nil
}
