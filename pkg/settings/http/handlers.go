package http

import (
	"encoding/json"
	"net/http"
	"strconv"

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
		r.Post("/", h.CreateSetting)       // Create a new setting
		r.Get("/{id}", h.GetSetting)       // Get a setting by ID
		r.Get("/", h.ListSettings)         // List all settings
		r.Put("/{id}", h.UpdateSetting)    // Update a setting
		r.Delete("/{id}", h.DeleteSetting) // Delete a setting
	})
}

// CreateSetting handles setting creation
// @Summary Create a new setting
// @Description Create a new setting with the provided configuration
// @Tags settings
// @Accept json
// @Produce json
// @Param setting body service.CreateSettingParams true "Setting configuration"
// @Success 200 {object} service.Setting
// @Router /settings [post]
func (h *Handler) CreateSetting(w http.ResponseWriter, r *http.Request) {
	var params service.CreateSettingParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	setting, err := h.service.CreateSetting(r.Context(), params)
	if err != nil {
		h.logger.Error("Failed to create setting", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setting)
}

// GetSetting handles setting retrieval
// @Summary Get a setting by ID
// @Description Get a setting's details by its ID
// @Tags settings
// @Produce json
// @Param id path int true "Setting ID"
// @Success 200 {object} service.Setting
// @Router /settings/{id} [get]
func (h *Handler) GetSetting(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid setting ID", http.StatusBadRequest)
		return
	}

	setting, err := h.service.GetSetting(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get setting", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setting)
}

// ListSettings handles settings listing
// @Summary List all settings
// @Description Get a list of all settings
// @Tags settings
// @Produce json
// @Success 200 {array} service.Setting
// @Router /settings [get]
func (h *Handler) ListSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.ListSettings(r.Context())
	if err != nil {
		h.logger.Error("Failed to list settings", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateSetting handles setting updates
// @Summary Update a setting
// @Description Update an existing setting by its ID
// @Tags settings
// @Accept json
// @Produce json
// @Param id path int true "Setting ID"
// @Param setting body service.UpdateSettingParams true "Updated setting configuration"
// @Success 200 {object} service.Setting
// @Router /settings/{id} [put]
func (h *Handler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid setting ID", http.StatusBadRequest)
		return
	}

	var params service.UpdateSettingParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	setting, err := h.service.UpdateSetting(r.Context(), id, params)
	if err != nil {
		h.logger.Error("Failed to update setting", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setting)
}

// DeleteSetting handles setting deletion
// @Summary Delete a setting
// @Description Delete a setting by its ID
// @Tags settings
// @Param id path int true "Setting ID"
// @Success 204 "No Content"
// @Router /settings/{id} [delete]
func (h *Handler) DeleteSetting(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid setting ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteSetting(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete setting", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
