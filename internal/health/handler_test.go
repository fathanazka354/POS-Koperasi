package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLiveness(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	Liveness(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Body.String(); got != `{"status":"ok"}` {
		t.Fatalf("body = %q", got)
	}
}
