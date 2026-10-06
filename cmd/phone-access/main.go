package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/infrai-examples/devtool-phone-access/internal/phoneaccess"
)

type phoneInput struct {
	Phone  string `json:"phone"`
	Code   string `json:"code,omitempty"`
	Locale string `json:"locale,omitempty"`
}

type server struct {
	access *phoneaccess.AccessService
}

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	baseURL := os.Getenv("INFRAI_BASE_URL")
	if baseURL == "" {
		baseURL = phoneaccess.DefaultBaseURL
	}

	s := &server{access: phoneaccess.NewAccessService(phoneaccess.NewClient(baseURL, apiKey, nil))}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /{flow}/code", s.requestCode)
	mux.HandleFunc("POST /{flow}/verify", s.verifyCode)

	addr := ":8080"
	log.Printf("phone access service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *server) requestCode(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeInput(w, r)
	if !ok {
		return
	}
	data, err := s.access.RequestCode(r.Context(), phoneaccess.Flow(r.PathValue("flow")), input.Phone, input.Locale)
	writeResult(w, data, err)
}

func (s *server) verifyCode(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeInput(w, r)
	if !ok {
		return
	}
	data, err := s.access.VerifyCode(r.Context(), phoneaccess.Flow(r.PathValue("flow")), input.Phone, input.Code)
	writeResult(w, data, err)
}

func decodeInput(w http.ResponseWriter, r *http.Request) (phoneInput, bool) {
	defer r.Body.Close()
	var input phoneInput
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || strings.TrimSpace(input.Phone) == "" {
		http.Error(w, "invalid JSON or missing phone", http.StatusBadRequest)
		return phoneInput{}, false
	}
	return input, true
}

func writeResult(w http.ResponseWriter, data json.RawMessage, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		status := http.StatusBadGateway
		var apiErr *phoneaccess.APIError
		if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
			status = apiErr.HTTPStatus
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
}
