package platform

import (
	"encoding/json"
	"net/http"
)

// Status is the body of GET /healthz.
type Status struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// HealthStatus returns the health body for a service.
func HealthStatus(service string) Status { return Status{Status: "ok", Service: service} }

// Health is an http.HandlerFunc for GET /healthz.
func Health(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthStatus(service))
	}
}
