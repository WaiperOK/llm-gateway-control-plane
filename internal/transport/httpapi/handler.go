package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/WaiperOK/llm-gateway-control-plane/internal/app"
	"github.com/WaiperOK/llm-gateway-control-plane/pkg/contracts"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler wires gateway application logic into HTTP endpoints.
type Handler struct {
	logger *slog.Logger
	app    *app.Service
	mux    *http.ServeMux
}

func NewHandler(logger *slog.Logger, svc *app.Service) *Handler {
	h := &Handler{logger: logger, app: svc, mux: http.NewServeMux()}
	h.routes()
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) routes() {
	h.mux.HandleFunc("/healthz", h.handleHealth)
	h.mux.Handle("/metrics", promhttp.Handler())
	h.mux.HandleFunc("/v1/gateway/completions", h.handleCompletion)
	h.mux.HandleFunc("/v1/teams/me/usage", h.handleUsage)
	h.mux.HandleFunc("/v1/audit", h.handleAudit)
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleCompletion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, contracts.ErrorResponse{Error: "method not allowed", Code: "method_not_allowed"})
		return
	}
	requestID := reqID()
	principal, authErr := h.app.Authenticate(r)
	if authErr != nil {
		writeJSON(w, authErr.HTTPStatus, authErr.WithRequestID(requestID))
		return
	}

	var req contracts.CompletionRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, contracts.ErrorResponse{Error: "invalid JSON body", Code: "invalid_json", RequestID: requestID})
		return
	}

	resp, appErr := h.app.HandleCompletion(r.Context(), requestID, principal, req)
	if appErr != nil {
		writeJSON(w, appErr.HTTPStatus, appErr.WithRequestID(requestID))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, contracts.ErrorResponse{Error: "method not allowed", Code: "method_not_allowed"})
		return
	}
	requestID := reqID()
	principal, authErr := h.app.Authenticate(r)
	if authErr != nil {
		writeJSON(w, authErr.HTTPStatus, authErr.WithRequestID(requestID))
		return
	}
	writeJSON(w, http.StatusOK, h.app.Usage(principal))
}

func (h *Handler) handleAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, contracts.ErrorResponse{Error: "method not allowed", Code: "method_not_allowed"})
		return
	}
	requestID := reqID()
	principal, authErr := h.app.Authenticate(r)
	if authErr != nil {
		writeJSON(w, authErr.HTTPStatus, authErr.WithRequestID(requestID))
		return
	}

	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": h.app.AuditEvents(principal, limit)})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func reqID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "req-fallback"
	}
	return "req-" + hex.EncodeToString(buf)
}
