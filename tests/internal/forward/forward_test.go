package forward_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elysiandb/elysian-gate/internal/forward"
)

func TestForwardRequest_Success(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST method")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer s.Close()

	status, body, err := forward.ForwardRequest("POST", s.URL, `{"id":1}`)
	if err != nil || status != 200 || body == "" {
		t.Fatalf("unexpected result: %v %d %s", err, status, body)
	}
}

func TestForwardRequest_InvalidURL(t *testing.T) {
	_, _, err := forward.ForwardRequest("GET", ":", "")
	if err == nil {
		t.Fatalf("expected error for invalid url")
	}
}

func TestForwardRequest_ServerError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", 500)
	}))
	defer s.Close()

	status, body, err := forward.ForwardRequest("GET", s.URL, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != 500 || body == "" {
		t.Fatalf("expected 500 status, got %d", status)
	}
}
