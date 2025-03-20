package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/chainlaunch/chainlaunch/pkg/fabric/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type OrganizationHandler struct {
	service *service.OrganizationService
}

func NewOrganizationHandler(service *service.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		service: service,
	}
}

// RegisterRoutes registers the organization routes
func (h *OrganizationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/organizations", func(r chi.Router) {
		r.Post("/", h.CreateOrganization)
		r.Get("/", h.ListOrganizations)
		r.Get("/by-mspid/{mspid}", h.GetOrganizationByMspID)
		r.Get("/{id}", h.GetOrganization)
		r.Put("/{id}", h.UpdateOrganization)
		r.Delete("/{id}", h.DeleteOrganization)
	})
}

// @Summary Create a new Fabric organization
// @Description Create a new Fabric organization with the specified configuration
// @Tags organizations
// @Accept json
// @Produce json
// @Param request body CreateOrganizationRequest true "Organization creation request"
// @Success 201 {object} OrganizationResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /organizations [post]
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	params := service.CreateOrganizationParams{
		MspID:       req.MspID,
		Name:        req.Name,
		Description: req.Description,
		ProviderID:  req.ProviderID,
	}

	org, err := h.service.CreateOrganization(r.Context(), params)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, toOrganizationResponse(org))
}

// @Summary Get a Fabric organization
// @Description Get a Fabric organization by ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "Organization ID"
// @Success 200 {object} OrganizationResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid organization ID"})
		return
	}

	org, err := h.service.GetOrganization(r.Context(), id)
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.JSON(w, r, toOrganizationResponse(org))
}

// @Summary Get a Fabric organization by MSP ID
// @Description Get a Fabric organization by MSP ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param mspid path string true "MSP ID"
// @Success 200 {object} OrganizationResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /organizations/by-mspid/{mspid} [get]
func (h *OrganizationHandler) GetOrganizationByMspID(w http.ResponseWriter, r *http.Request) {
	mspid := chi.URLParam(r, "mspid")
	if mspid == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid MSP ID"})
		return
	}

	org, err := h.service.GetOrganizationByMspID(r.Context(), mspid)
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.JSON(w, r, toOrganizationResponse(org))
}

// @Summary Update a Fabric organization
// @Description Update an existing Fabric organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "Organization ID"
// @Param request body UpdateOrganizationRequest true "Organization update request"
// @Success 200 {object} OrganizationResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid organization ID"})
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	org, err := h.service.UpdateOrganization(r.Context(), id, service.UpdateOrganizationParams{
		Description: req.Description,
	})
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.JSON(w, r, toOrganizationResponse(org))
}

// @Summary Delete a Fabric organization
// @Description Delete a Fabric organization by ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path int true "Organization ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid organization ID"})
		return
	}

	if err := h.service.DeleteOrganization(r.Context(), id); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	render.Status(r, http.StatusNoContent)
}

// @Summary List all Fabric organizations
// @Description Get a list of all Fabric organizations
// @Tags organizations
// @Accept json
// @Produce json
// @Success 200 {array} OrganizationResponse
// @Failure 500 {object} map[string]string
// @Router /organizations [get]
func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.service.ListOrganizations(r.Context())
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}

	response := make([]*OrganizationResponse, len(orgs))
	for i, org := range orgs {
		response[i] = toOrganizationResponse(&org)
	}

	render.JSON(w, r, response)
}
