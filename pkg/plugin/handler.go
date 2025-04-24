package plugin

import (
	"encoding/json"
	"net/http"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/plugin/types"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for plugins
type Handler struct {
	store  Store
	pm     *PluginManager
	logger *logger.Logger
}

// NewHandler creates a new plugin handler
func NewHandler(store Store, pm *PluginManager, logger *logger.Logger) *Handler {
	return &Handler{
		store:  store,
		pm:     pm,
		logger: logger,
	}
}

// RegisterRoutes registers the plugin routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/plugins", func(r chi.Router) {
		r.Get("/", h.listPlugins)
		r.Post("/", h.createPlugin)
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", h.getPlugin)
			r.Put("/", h.updatePlugin)
			r.Delete("/", h.deletePlugin)
			r.Post("/deploy", h.deployPlugin)
			r.Post("/stop", h.stopPlugin)
			r.Get("/status", h.getPluginStatus)
		})
	})
}

// @Summary List all plugins
// @Description Get a list of all available plugins
// @Tags plugins
// @Accept json
// @Produce json
// @Success 200 {array} types.Plugin
// @Failure 500 {object} string
// @Router /plugins [get]
func (h *Handler) listPlugins(w http.ResponseWriter, r *http.Request) {
	plugins, err := h.store.ListPlugins(r.Context())
	if err != nil {
		h.logger.Errorf("Failed to list plugins: %v", err)
		http.Error(w, "Failed to list plugins", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(plugins); err != nil {
		h.logger.Errorf("Failed to encode plugins: %v", err)
		http.Error(w, "Failed to encode plugins", http.StatusInternalServerError)
		return
	}
}

// @Summary Get a plugin
// @Description Get a specific plugin by name
// @Tags plugins
// @Accept json
// @Produce json
// @Param name path string true "Plugin name"
// @Success 200 {object} types.Plugin
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /plugins/{name} [get]
func (h *Handler) getPlugin(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	plugin, err := h.store.GetPlugin(r.Context(), name)
	if err != nil {
		h.logger.Errorf("Failed to get plugin: %v", err)
		http.Error(w, "Failed to get plugin", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(plugin); err != nil {
		h.logger.Errorf("Failed to encode plugin: %v", err)
		http.Error(w, "Failed to encode plugin", http.StatusInternalServerError)
		return
	}
}

// @Summary Create a plugin
// @Description Create a new plugin
// @Tags plugins
// @Accept json
// @Produce json
// @Param plugin body types.Plugin true "Plugin to create"
// @Success 201 {object} types.Plugin
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /plugins [post]
func (h *Handler) createPlugin(w http.ResponseWriter, r *http.Request) {
	var plugin types.Plugin
	if err := json.NewDecoder(r.Body).Decode(&plugin); err != nil {
		h.logger.Errorf("Failed to decode plugin: %v", err)
		http.Error(w, "Failed to decode plugin", http.StatusBadRequest)
		return
	}

	if err := h.pm.ValidatePlugin(&plugin); err != nil {
		h.logger.Errorf("Invalid plugin: %v", err)
		http.Error(w, "Invalid plugin", http.StatusBadRequest)
		return
	}

	if err := h.store.CreatePlugin(r.Context(), &plugin); err != nil {
		h.logger.Errorf("Failed to create plugin: %v", err)
		http.Error(w, "Failed to create plugin", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(plugin); err != nil {
		h.logger.Errorf("Failed to encode plugin: %v", err)
		http.Error(w, "Failed to encode plugin", http.StatusInternalServerError)
		return
	}
}

// @Summary Update a plugin
// @Description Update an existing plugin
// @Tags plugins
// @Accept json
// @Produce json
// @Param name path string true "Plugin name"
// @Param plugin body types.Plugin true "Plugin to update"
// @Success 200 {object} types.Plugin
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /plugins/{name} [put]
func (h *Handler) updatePlugin(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var plugin types.Plugin
	if err := json.NewDecoder(r.Body).Decode(&plugin); err != nil {
		h.logger.Errorf("Failed to decode plugin: %v", err)
		http.Error(w, "Failed to decode plugin", http.StatusBadRequest)
		return
	}

	if plugin.Metadata.Name != name {
		http.Error(w, "Plugin name in URL does not match body", http.StatusBadRequest)
		return
	}

	if err := h.pm.ValidatePlugin(&plugin); err != nil {
		h.logger.Errorf("Invalid plugin: %v", err)
		http.Error(w, "Invalid plugin", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdatePlugin(r.Context(), &plugin); err != nil {
		h.logger.Errorf("Failed to update plugin: %v", err)
		http.Error(w, "Failed to update plugin", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(plugin); err != nil {
		h.logger.Errorf("Failed to encode plugin: %v", err)
		http.Error(w, "Failed to encode plugin", http.StatusInternalServerError)
		return
	}
}

// @Summary Delete a plugin
// @Description Delete an existing plugin
// @Tags plugins
// @Accept json
// @Produce json
// @Param name path string true "Plugin name"
// @Success 204
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /plugins/{name} [delete]
func (h *Handler) deletePlugin(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := h.store.DeletePlugin(r.Context(), name); err != nil {
		h.logger.Errorf("Failed to delete plugin: %v", err)
		http.Error(w, "Failed to delete plugin", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Deploy a plugin
// @Description Deploy a plugin with the given parameters
// @Tags plugins
// @Accept json
// @Produce json
// @Param name path string true "Plugin name"
// @Param parameters body map[string]interface{} true "Deployment parameters"
// @Success 200
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /plugins/{name}/deploy [post]
func (h *Handler) deployPlugin(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	plugin, err := h.store.GetPlugin(r.Context(), name)
	if err != nil {
		h.logger.Errorf("Failed to get plugin: %v", err)
		http.Error(w, "Failed to get plugin", http.StatusInternalServerError)
		return
	}

	var parameters map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&parameters); err != nil {
		h.logger.Errorf("Failed to decode parameters: %v", err)
		http.Error(w, "Failed to decode parameters", http.StatusBadRequest)
		return
	}

	// Validate required parameters
	for _, required := range plugin.Spec.Parameters.Required {
		if _, ok := parameters[required]; !ok {
			http.Error(w, "Missing required parameter: "+required, http.StatusBadRequest)
			return
		}
	}

	// Validate parameter types and values
	for name, value := range parameters {
		if spec, ok := plugin.Spec.Parameters.Properties[name]; ok {
			// Check if value is of correct type
			if spec.Type == "string" {
				if _, ok := value.(string); !ok {
					http.Error(w, "Invalid type for parameter "+name+": expected string", http.StatusBadRequest)
					return
				}
			}

			// Check if value is in enum if specified
			if len(spec.Enum) > 0 {
				valid := false
				for _, allowed := range spec.Enum {
					if value == allowed {
						valid = true
						break
					}
				}
				if !valid {
					http.Error(w, "Invalid value for parameter "+name+": not in allowed values", http.StatusBadRequest)
					return
				}
			}
		}
	}

	if err := h.pm.DeployPlugin(r.Context(), plugin, parameters); err != nil {
		h.logger.Errorf("Failed to deploy plugin: %v", err)
		http.Error(w, "Failed to deploy plugin", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Stop a plugin deployment
// @Description Stop a running plugin deployment
// @Tags plugins
// @Accept json
// @Produce json
// @Param name path string true "Plugin name"
// @Success 200
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /plugins/{name}/stop [post]
func (h *Handler) stopPlugin(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	plugin, err := h.store.GetPlugin(r.Context(), name)
	if err != nil {
		h.logger.Errorf("Failed to get plugin: %v", err)
		http.Error(w, "Failed to get plugin", http.StatusInternalServerError)
		return
	}

	if err := h.pm.StopPlugin(r.Context(), plugin); err != nil {
		h.logger.Errorf("Failed to stop plugin: %v", err)
		http.Error(w, "Failed to stop plugin", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Get plugin deployment status
// @Description Get the current status of a plugin deployment
// @Tags plugins
// @Accept json
// @Produce json
// @Param name path string true "Plugin name"
// @Success 200 {object} types.DeploymentStatus
// @Failure 404 {object} string
// @Failure 500 {object} string
// @Router /plugins/{name}/status [get]
func (h *Handler) getPluginStatus(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	plugin, err := h.store.GetPlugin(r.Context(), name)
	if err != nil {
		h.logger.Errorf("Failed to get plugin: %v", err)
		http.Error(w, "Failed to get plugin", http.StatusInternalServerError)
		return
	}

	status, err := h.pm.GetPluginStatus(r.Context(), plugin)
	if err != nil {
		h.logger.Errorf("Failed to get plugin status: %v", err)
		http.Error(w, "Failed to get plugin status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Errorf("Failed to encode plugin status: %v", err)
		http.Error(w, "Failed to encode plugin status", http.StatusInternalServerError)
		return
	}
}
