package http

import (
	"encoding/json"
	"net/http"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/settings/service"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for settings
type Handler struct {
	service *service.SettingsService
	logger  *logger.Logger
}

// NewHandler creates a new settings handler
func NewHandler(service *service.SettingsService, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers the settings routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/settings", func(r chi.Router) {
		r.Post("/", h.CreateOrUpdateSetting) // Create or update the default setting
		r.Get("/", h.GetSetting)             // Get the default setting
		r.Put("/", h.CreateOrUpdateSetting)  // Update the default setting (same as POST)
	})
}

// CreateOrUpdateSetting handles setting creation or update
// @Summary Create or update the default setting
// @Description Create or update the default setting with the provided configuration
// @Tags settings
// @Accept json
// @Produce json
// @Param setting body service.CreateSettingParams true "Setting configuration"
// @Success 200 {object} service.Setting
// @Router /settings [post]
// @BasePath /api/v1
func (h *Handler) CreateOrUpdateSetting(w http.ResponseWriter, r *http.Request) {
	var params service.CreateSettingParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	setting, err := h.service.CreateSetting(r.Context(), params)
	if err != nil {
		h.logger.Error("Failed to create/update setting", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setting)
}

// GetSetting handles default setting retrieval
// @Summary Get the default setting
// @Description Get the default setting's details
// @Tags settings
// @Produce json
// @Success 200 {object} service.Setting
// @Router /settings [get]
// @BasePath /api/v1
func (h *Handler) GetSetting(w http.ResponseWriter, r *http.Request) {
	setting, err := h.service.GetSetting(r.Context())
	if err != nil {
		h.logger.Error("Failed to get setting", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setting)
}
