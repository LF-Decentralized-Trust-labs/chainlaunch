package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/chainlaunch/chainlaunch/pkg/backups/service"
	"github.com/chainlaunch/chainlaunch/pkg/errors"
	"github.com/chainlaunch/chainlaunch/pkg/http/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  *service.BackupService
	validate *validator.Validate
}

func NewHandler(service *service.BackupService) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

// RegisterRoutes registers the backup routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/backups", func(r chi.Router) {
		// Backup targets
		r.Post("/targets", response.Middleware(h.CreateBackupTarget))
		r.Get("/targets", response.Middleware(h.ListBackupTargets))
		r.Get("/targets/{id}", response.Middleware(h.GetBackupTarget))
		r.Delete("/targets/{id}", response.Middleware(h.DeleteBackupTarget))
		r.Put("/targets/{id}", response.Middleware(h.UpdateBackupTarget))

		// Backup schedules
		r.Post("/schedules", response.Middleware(h.CreateBackupSchedule))
		r.Get("/schedules", response.Middleware(h.ListBackupSchedules))
		r.Get("/schedules/{id}", response.Middleware(h.GetBackupSchedule))
		r.Put("/schedules/{id}/enable", response.Middleware(h.EnableBackupSchedule))
		r.Put("/schedules/{id}/disable", response.Middleware(h.DisableBackupSchedule))
		r.Delete("/schedules/{id}", response.Middleware(h.DeleteBackupSchedule))
		r.Put("/schedules/{id}", response.Middleware(h.UpdateBackupSchedule))

		// Backups
		r.Get("/", response.Middleware(h.ListBackups))
		r.Post("/", response.Middleware(h.CreateBackup))
		r.Get("/{id}", response.Middleware(h.GetBackup))
		r.Delete("/{id}", response.Middleware(h.DeleteBackup))
	})
}

// CreateBackupTarget godoc
// @Summary Create a new backup target
// @Description Create a new backup target with the specified configuration
// @Tags backup-targets
// @Accept json
// @Produce json
// @Param request body CreateBackupTargetRequest true "Backup target creation request"
// @Success 201 {object} BackupTargetResponse
// @Failure 400 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/targets [post]
func (h *Handler) CreateBackupTarget(w http.ResponseWriter, r *http.Request) error {
	var req CreateBackupTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.NewValidationError("invalid request body", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_REQUEST_BODY",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors[err.Field()] = err.Tag()
		}
		return errors.NewValidationError("validation failed", map[string]interface{}{
			"detail": "Request validation failed",
			"code":   "VALIDATION_ERROR",
			"errors": validationErrors,
		})
	}

	target, err := h.service.CreateBackupTarget(r.Context(), service.CreateBackupTargetParams{
		Name:           req.Name,
		Type:           service.BackupTargetType(req.Type),
		BucketName:     req.BucketName,
		Region:         req.Region,
		Endpoint:       req.Endpoint,
		BucketPath:     req.BucketPath,
		AccessKeyID:    req.AccessKeyID,
		SecretKey:      req.SecretKey,
		ForcePathStyle: req.ForcePathStyle,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return errors.NewConflictError("backup target already exists", map[string]interface{}{
				"detail": err.Error(),
				"code":   "TARGET_ALREADY_EXISTS",
			})
		}
		return errors.NewInternalError("failed to create backup target", err, nil)
	}

	return response.WriteJSON(w, http.StatusCreated, toBackupTargetResponse(target))
}

// ListBackupTargets godoc
// @Summary List all backup targets
// @Description Get a list of all backup targets
// @Tags backup-targets
// @Accept json
// @Produce json
// @Success 200 {array} BackupTargetResponse
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/targets [get]
func (h *Handler) ListBackupTargets(w http.ResponseWriter, r *http.Request) error {
	targets, err := h.service.ListBackupTargets(r.Context())
	if err != nil {
		return errors.NewInternalError("failed to list backup targets", err, nil)
	}

	responses := make([]BackupTargetResponse, len(targets))
	for i, target := range targets {
		responses[i] = toBackupTargetResponse(target)
	}

	return response.WriteJSON(w, http.StatusOK, responses)
}

// GetBackupTarget godoc
// @Summary Get a backup target by ID
// @Description Get detailed information about a specific backup target
// @Tags backup-targets
// @Accept json
// @Produce json
// @Param id path int true "Backup Target ID"
// @Success 200 {object} BackupTargetResponse
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Target not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/targets/{id} [get]
func (h *Handler) GetBackupTarget(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid backup target ID", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_ID_FORMAT",
		})
	}

	target, err := h.service.GetBackupTarget(r.Context(), id)
	if err != nil {
		if err == service.ErrTargetNotFound {
			return errors.NewNotFoundError("backup target not found", map[string]interface{}{
				"detail":    "The requested backup target does not exist",
				"code":      "TARGET_NOT_FOUND",
				"target_id": id,
			})
		}
		return errors.NewInternalError("failed to get backup target", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toBackupTargetResponse(target))
}

// DeleteBackupTarget godoc
// @Summary Delete a backup target
// @Description Delete a backup target and all associated backups
// @Tags backup-targets
// @Accept json
// @Produce json
// @Param id path int true "Backup Target ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Target not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/targets/{id} [delete]
func (h *Handler) DeleteBackupTarget(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid backup target ID", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_ID_FORMAT",
		})
	}

	if err := h.service.DeleteBackupTarget(r.Context(), id); err != nil {
		if err == service.ErrTargetNotFound {
			return errors.NewNotFoundError("backup target not found", map[string]interface{}{
				"detail":    "The requested backup target does not exist",
				"code":      "TARGET_NOT_FOUND",
				"target_id": id,
			})
		}
		return errors.NewInternalError("failed to delete backup target", err, nil)
	}

	return response.WriteJSON(w, http.StatusNoContent, nil)
}

// CreateBackupSchedule godoc
// @Summary Create a new backup schedule
// @Description Create a new backup schedule with the specified configuration
// @Tags backup-schedules
// @Accept json
// @Produce json
// @Param request body CreateBackupScheduleRequest true "Backup schedule creation request"
// @Success 201 {object} BackupScheduleResponse
// @Failure 400 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/schedules [post]
func (h *Handler) CreateBackupSchedule(w http.ResponseWriter, r *http.Request) error {
	var req CreateBackupScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.NewValidationError("invalid request body", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_REQUEST_BODY",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors[err.Field()] = err.Tag()
		}
		return errors.NewValidationError("validation failed", map[string]interface{}{
			"detail": "Request validation failed",
			"code":   "VALIDATION_ERROR",
			"errors": validationErrors,
		})
	}

	schedule, err := h.service.CreateBackupSchedule(r.Context(), service.CreateBackupScheduleParams{
		Name:           req.Name,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		TargetID:       req.TargetID,
		RetentionDays:  req.RetentionDays,
		Enabled:        req.Enabled,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid cron expression") {
			return errors.NewValidationError("invalid cron expression", map[string]interface{}{
				"detail": err.Error(),
				"code":   "INVALID_CRON_EXPRESSION",
			})
		}
		return errors.NewInternalError("failed to create backup schedule", err, nil)
	}

	return response.WriteJSON(w, http.StatusCreated, toBackupScheduleResponse(schedule))
}

// ListBackupSchedules godoc
// @Summary List all backup schedules
// @Description Get a list of all backup schedules
// @Tags backup-schedules
// @Accept json
// @Produce json
// @Success 200 {array} BackupScheduleResponse
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/schedules [get]
func (h *Handler) ListBackupSchedules(w http.ResponseWriter, r *http.Request) error {
	schedules, err := h.service.ListBackupSchedules(r.Context())
	if err != nil {
		return errors.NewInternalError("failed to list backup schedules", err, nil)
	}

	responses := make([]BackupScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = toBackupScheduleResponse(schedule)
	}

	return response.WriteJSON(w, http.StatusOK, responses)
}

// GetBackupSchedule godoc
// @Summary Get a backup schedule by ID
// @Description Get detailed information about a specific backup schedule
// @Tags backup-schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 200 {object} BackupScheduleResponse
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Schedule not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/schedules/{id} [get]
func (h *Handler) GetBackupSchedule(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid schedule ID", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_ID_FORMAT",
		})
	}

	schedule, err := h.service.GetBackupSchedule(r.Context(), id)
	if err != nil {
		if err == service.ErrScheduleNotFound {
			return errors.NewNotFoundError("backup schedule not found", map[string]interface{}{
				"detail":      "The requested backup schedule does not exist",
				"code":        "SCHEDULE_NOT_FOUND",
				"schedule_id": id,
			})
		}
		return errors.NewInternalError("failed to get backup schedule", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toBackupScheduleResponse(schedule))
}

// EnableBackupSchedule godoc
// @Summary Enable a backup schedule
// @Description Enable a backup schedule to start running
// @Tags backup-schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 200 {object} BackupScheduleResponse
// @Failure 400 {object} response.Response "Validation error"
// @Failure 404 {object} response.Response "Schedule not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/schedules/{id}/enable [put]
func (h *Handler) EnableBackupSchedule(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid schedule ID", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_ID_FORMAT",
		})
	}

	schedule, err := h.service.EnableBackupSchedule(r.Context(), id)
	if err != nil {
		if err == service.ErrScheduleNotFound {
			return errors.NewNotFoundError("backup schedule not found", map[string]interface{}{
				"detail":      "The requested backup schedule does not exist",
				"code":        "SCHEDULE_NOT_FOUND",
				"schedule_id": id,
			})
		}
		return errors.NewInternalError("failed to enable backup schedule", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toBackupScheduleResponse(schedule))
}

// DisableBackupSchedule godoc
// @Summary Disable a backup schedule
// @Description Disable a backup schedule to stop it from running
// @Tags backup-schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 200 {object} BackupScheduleResponse
// @Failure 400 {object} response.Response "Validation error"
// @Failure 404 {object} response.Response "Schedule not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/schedules/{id}/disable [put]
func (h *Handler) DisableBackupSchedule(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid schedule ID", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_ID_FORMAT",
		})
	}

	schedule, err := h.service.DisableBackupSchedule(r.Context(), id)
	if err != nil {
		if err == service.ErrScheduleNotFound {
			return errors.NewNotFoundError("backup schedule not found", map[string]interface{}{
				"detail":      "The requested backup schedule does not exist",
				"code":        "SCHEDULE_NOT_FOUND",
				"schedule_id": id,
			})
		}
		return errors.NewInternalError("failed to disable backup schedule", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toBackupScheduleResponse(schedule))
}

// DeleteBackupSchedule godoc
// @Summary Delete a backup schedule
// @Description Delete a backup schedule and stop its execution
// @Tags backup-schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Schedule not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/schedules/{id} [delete]
func (h *Handler) DeleteBackupSchedule(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid schedule ID", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_ID_FORMAT",
		})
	}

	if err := h.service.DeleteBackupSchedule(r.Context(), id); err != nil {
		if err == service.ErrScheduleNotFound {
			return errors.NewNotFoundError("backup schedule not found", map[string]interface{}{
				"detail":      "The requested backup schedule does not exist",
				"code":        "SCHEDULE_NOT_FOUND",
				"schedule_id": id,
			})
		}
		return errors.NewInternalError("failed to delete backup schedule", err, nil)
	}

	return response.WriteJSON(w, http.StatusNoContent, nil)
}

// ListBackups godoc
// @Summary List all backups
// @Description Get a list of all backups
// @Tags backups
// @Accept json
// @Produce json
// @Success 200 {array} BackupResponse
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups [get]
func (h *Handler) ListBackups(w http.ResponseWriter, r *http.Request) error {
	backups, err := h.service.ListBackups(r.Context())
	if err != nil {
		return errors.NewInternalError("failed to list backups", err, nil)
	}

	responses := make([]BackupResponse, len(backups))
	for i, backup := range backups {
		responses[i] = toBackupResponse(backup)
	}

	return response.WriteJSON(w, http.StatusOK, responses)
}

// CreateBackup godoc
// @Summary Create a new backup
// @Description Create a new backup with the specified configuration
// @Tags backups
// @Accept json
// @Produce json
// @Param request body CreateBackupRequest true "Backup creation request"
// @Success 201 {object} BackupResponse
// @Failure 400 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups [post]
func (h *Handler) CreateBackup(w http.ResponseWriter, r *http.Request) error {
	var req CreateBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.NewValidationError("invalid request body", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_REQUEST_BODY",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors[err.Field()] = err.Tag()
		}
		return errors.NewValidationError("validation failed", map[string]interface{}{
			"detail": "Request validation failed",
			"code":   "VALIDATION_ERROR",
			"errors": validationErrors,
		})
	}

	// Convert metadata to string if present
	var metadataStr *string
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return errors.NewValidationError("invalid metadata format", map[string]interface{}{
				"detail": err.Error(),
				"code":   "INVALID_METADATA_FORMAT",
			})
		}
		str := string(metadataBytes)
		metadataStr = &str
	}

	backup, err := h.service.CreateBackup(r.Context(), service.CreateBackupParams{
		ScheduleID: req.ScheduleID,
		TargetID:   req.TargetID,
		Metadata:   metadataStr,
	})
	if err != nil {
		return errors.NewInternalError("failed to create backup", err, nil)
	}

	return response.WriteJSON(w, http.StatusCreated, toBackupResponse(backup))
}

// GetBackup godoc
// @Summary Get a backup by ID
// @Description Get detailed information about a specific backup
// @Tags backups
// @Accept json
// @Produce json
// @Param id path int true "Backup ID"
// @Success 200 {object} BackupResponse
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Backup not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/{id} [get]
func (h *Handler) GetBackup(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid backup ID", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_ID_FORMAT",
		})
	}

	backup, err := h.service.GetBackup(r.Context(), id)
	if err != nil {
		if err == service.ErrBackupNotFound {
			return errors.NewNotFoundError("backup not found", map[string]interface{}{
				"detail":    "The requested backup does not exist",
				"code":      "BACKUP_NOT_FOUND",
				"backup_id": id,
			})
		}
		return errors.NewInternalError("failed to get backup", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toBackupResponse(backup))
}

// DeleteBackup godoc
// @Summary Delete a backup
// @Description Delete a backup and its associated files
// @Tags backups
// @Accept json
// @Produce json
// @Param id path int true "Backup ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Backup not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/{id} [delete]
func (h *Handler) DeleteBackup(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid backup ID", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_ID_FORMAT",
		})
	}

	if err := h.service.DeleteBackup(r.Context(), id); err != nil {
		if err == service.ErrBackupNotFound {
			return errors.NewNotFoundError("backup not found", map[string]interface{}{
				"detail":    "The requested backup does not exist",
				"code":      "BACKUP_NOT_FOUND",
				"backup_id": id,
			})
		}
		return errors.NewInternalError("failed to delete backup", err, nil)
	}

	return response.WriteJSON(w, http.StatusNoContent, nil)
}

// UpdateBackupTarget godoc
// @Summary Update a backup target
// @Description Update an existing backup target with new configuration
// @Tags backup-targets
// @Accept json
// @Produce json
// @Param id path int true "Backup Target ID"
// @Param request body UpdateBackupTargetRequest true "Backup target update request"
// @Success 200 {object} BackupTargetResponse
// @Failure 400 {object} response.Response "Validation error"
// @Failure 404 {object} response.Response "Target not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/targets/{id} [put]
func (h *Handler) UpdateBackupTarget(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid backup target ID", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_ID_FORMAT",
		})
	}

	var req UpdateBackupTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.NewValidationError("invalid request body", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_REQUEST_BODY",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors[err.Field()] = err.Tag()
		}
		return errors.NewValidationError("validation failed", map[string]interface{}{
			"detail": "Request validation failed",
			"code":   "VALIDATION_ERROR",
			"errors": validationErrors,
		})
	}

	target, err := h.service.UpdateBackupTarget(r.Context(), service.UpdateBackupTargetParams{
		ID:             id,
		Name:           req.Name,
		Type:           service.BackupTargetType(req.Type),
		BucketName:     req.BucketName,
		Region:         req.Region,
		Endpoint:       req.Endpoint,
		BucketPath:     req.BucketPath,
		AccessKeyID:    req.AccessKeyID,
		SecretKey:      req.SecretKey,
		ForcePathStyle: req.ForcePathStyle,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.NewNotFoundError("backup target not found", map[string]interface{}{
				"detail":    "The requested backup target does not exist",
				"code":      "TARGET_NOT_FOUND",
				"target_id": id,
			})
		}
		return errors.NewInternalError("failed to update backup target", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toBackupTargetResponse(target))
}

// UpdateBackupSchedule godoc
// @Summary Update a backup schedule
// @Description Update an existing backup schedule with new configuration
// @Tags backup-schedules
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Param request body UpdateBackupScheduleRequest true "Backup schedule update request"
// @Success 200 {object} BackupScheduleResponse
// @Failure 400 {object} response.Response "Validation error"
// @Failure 404 {object} response.Response "Schedule not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /backups/schedules/{id} [put]
func (h *Handler) UpdateBackupSchedule(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return errors.NewValidationError("invalid schedule ID", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_ID_FORMAT",
		})
	}

	var req UpdateBackupScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.NewValidationError("invalid request body", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_REQUEST_BODY",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors[err.Field()] = err.Tag()
		}
		return errors.NewValidationError("validation failed", map[string]interface{}{
			"detail": "Request validation failed",
			"code":   "VALIDATION_ERROR",
			"errors": validationErrors,
		})
	}

	schedule, err := h.service.UpdateBackupSchedule(r.Context(), service.UpdateBackupScheduleParams{
		ID:             id,
		Name:           req.Name,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		TargetID:       req.TargetID,
		RetentionDays:  req.RetentionDays,
		Enabled:        req.Enabled,
	})
	if err != nil {
		if err == service.ErrScheduleNotFound {
			return errors.NewNotFoundError("backup schedule not found", map[string]interface{}{
				"detail":      "The requested backup schedule does not exist",
				"code":        "SCHEDULE_NOT_FOUND",
				"schedule_id": id,
			})
		}
		return errors.NewInternalError("failed to update backup schedule", err, nil)
	}

	return response.WriteJSON(w, http.StatusOK, toBackupScheduleResponse(schedule))
}

// Helper functions to convert service DTOs to HTTP responses
func toBackupTargetResponse(target *service.BackupTargetDTO) BackupTargetResponse {
	return BackupTargetResponse{
		ID:             target.ID,
		Name:           target.Name,
		Type:           string(target.Type),
		BucketName:     target.BucketName,
		Region:         target.Region,
		Endpoint:       target.Endpoint,
		BucketPath:     target.BucketPath,
		AccessKeyID:    target.AccessKeyID,
		ForcePathStyle: target.ForcePathStyle,
		CreatedAt:      target.CreatedAt,
		UpdatedAt:      target.UpdatedAt,
	}
}

func toBackupScheduleResponse(schedule *service.BackupScheduleDTO) BackupScheduleResponse {
	return BackupScheduleResponse{
		ID:             schedule.ID,
		Name:           schedule.Name,
		Description:    schedule.Description,
		CronExpression: schedule.CronExpression,
		TargetID:       schedule.TargetID,
		RetentionDays:  schedule.RetentionDays,
		Enabled:        schedule.Enabled,
		CreatedAt:      schedule.CreatedAt,
		UpdatedAt:      schedule.UpdatedAt,
		LastRunAt:      schedule.LastRunAt,
		NextRunAt:      schedule.NextRunAt,
	}
}

func toBackupResponse(backup *service.BackupDTO) BackupResponse {
	return BackupResponse{
		ID:           backup.ID,
		ScheduleID:   backup.ScheduleID,
		TargetID:     backup.TargetID,
		Status:       string(backup.Status),
		SizeBytes:    backup.SizeBytes,
		StartedAt:    backup.StartedAt,
		CompletedAt:  backup.CompletedAt,
		ErrorMessage: backup.ErrorMessage,
		Metadata:     backup.Metadata,
		CreatedAt:    backup.CreatedAt,
	}
}
