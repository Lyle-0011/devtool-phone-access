package phoneaccess

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPhoneAccessHandoff(t *testing.T) {
	tests := []struct {
		name      string
		flow      Flow
		wantLogin bool
	}{
		{name: "signup creates identity", flow: Signup, wantLogin: false},
		{name: "login resolves identity", flow: Login, wantLogin: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paths []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("method = %s, want POST", r.Method)
				}
				if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("authorization = %q", got)
				}
				paths = append(paths, r.URL.Path)
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				var fields map[string]any
				if err := json.Unmarshal(body, &fields); err != nil {
					t.Fatal(err)
				}
				if r.URL.Path == "/v1/sms/otp" && (len(fields) != 1 || fields["to"] != "+15551234567") {
					t.Fatalf("sms fields = %v, want only to", fields)
				}
				if r.URL.Path == "/v1/auth/phone/verify" && fields["login"] != tt.wantLogin {
					t.Fatalf("login = %v, want %v", fields["login"], tt.wantLogin)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"ok":true,"data":{"accepted":true},"error":null,"metadata":{}}`))
			}))
			defer server.Close()

			service := NewAccessService(NewClient(server.URL, "test-key", server.Client()))
			if _, err := service.RequestCode(context.Background(), tt.flow, "+15551234567", "en"); err != nil {
				t.Fatal(err)
			}
			if _, err := service.VerifyCode(context.Background(), tt.flow, "+15551234567", "123456"); err != nil {
				t.Fatal(err)
			}
			if len(paths) != 2 || paths[0] != "/v1/sms/otp" || paths[1] != "/v1/auth/phone/verify" {
				t.Fatalf("handoff paths = %v", paths)
			}
		})
	}
}

func TestEnvelopeErrorKeepsClientStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"data":null,"error":{"code":"BAD_CODE","message":"code rejected"},"metadata":{}}`))
	}))
	defer server.Close()

	service := NewAccessService(NewClient(server.URL, "test-key", server.Client()))
	_, err := service.VerifyCode(context.Background(), Login, "+15551234567", "000000")
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("error = %#v", err)
	}
}
