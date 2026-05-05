package health

import "net/http"

// Liveness — tanpa cek DB/Redis (untuk probe Kubernetes / smoke load).
func Liveness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
