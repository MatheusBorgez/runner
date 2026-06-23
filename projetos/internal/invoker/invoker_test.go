package invoker_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyriosdata/runner/internal/invoker"
)

func TestHTTPInvokerSign(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"signature": "FAKE_SIG_==",
			"valid":     true,
			"message":   "ok",
		})
	}))
	defer srv.Close()

	inv := invoker.NewHTTPInvoker(srv.URL)
	resp, err := inv.Sign(invoker.SignRequest{Content: "dGVzdGU="})
	if err != nil {
		t.Fatalf("Sign HTTP falhou: %v", err)
	}
	if !resp.Valid {
		t.Errorf("esperado valid=true, obteve: %v", resp)
	}
}

func TestHTTPInvokerValidate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"signature": "FAKE_SIG_==",
			"valid":     true,
			"message":   "válida",
		})
	}))
	defer srv.Close()

	inv := invoker.NewHTTPInvoker(srv.URL)
	resp, err := inv.Validate(invoker.ValidateRequest{Content: "dGVzdGU=", Signature: "FAKE_SIG_=="})
	if err != nil {
		t.Fatalf("Validate HTTP falhou: %v", err)
	}
	if !resp.Valid {
		t.Errorf("esperado valid=true, obteve: %v", resp)
	}
}

func TestHTTPInvokerHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	inv := invoker.NewHTTPInvoker(srv.URL)
	if err := inv.Health(); err != nil {
		t.Errorf("Health falhou: %v", err)
	}
}

func TestHTTPInvokerHealthFails(t *testing.T) {
	inv := invoker.NewHTTPInvoker("http://localhost:19999")
	if err := inv.Health(); err == nil {
		t.Error("esperado erro para servidor inexistente")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	inv := invoker.NewHTTPInvoker(srv.URL)
	_, err := inv.Sign(invoker.SignRequest{Content: "dGVzdGU="})
	if err == nil {
		t.Error("esperado erro para resposta JSON inválida")
	}
}
