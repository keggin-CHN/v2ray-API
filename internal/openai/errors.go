package openai

import (
	"encoding/json"
	"net/http"
)

func WriteError(w http.ResponseWriter, status int, typ, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{
		Error: APIError{
			Message: message,
			Type:    typ,
			Code:    code,
		},
	})
}
