package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cansel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cansel()

	if err := h.db.Ping(ctx); err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"status":   "error",
			"database": "Database connection failed",
		})
		return
	}

	writeJson(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"database": "Database connection successful",
	})
}

func writeJson(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(data)
}
