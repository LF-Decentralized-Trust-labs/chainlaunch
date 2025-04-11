package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
)

// Setting represents a setting in the service layer
type Setting struct {
	ID        int64                  `json:"id"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// CreateSettingParams represents the parameters for creating a setting
type CreateSettingParams struct {
	Config map[string]interface{} `json:"config"`
}

// UpdateSettingParams represents the parameters for updating a setting
type UpdateSettingParams struct {
	Config map[string]interface{} `json:"config"`
}

// SettingsService handles operations for settings
type SettingsService struct {
	queries *db.Queries
	logger  *logger.Logger
}

// NewSettingsService creates a new settings service
func NewSettingsService(queries *db.Queries, logger *logger.Logger) *SettingsService {
	return &SettingsService{
		queries: queries,
		logger:  logger,
	}
}

// CreateSetting creates a new setting
func (s *SettingsService) CreateSetting(ctx context.Context, params CreateSettingParams) (*Setting, error) {
	configJSON, err := json.Marshal(params.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	dbSetting, err := s.queries.CreateSetting(ctx, string(configJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create setting: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(dbSetting.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &Setting{
		ID:        dbSetting.ID,
		Config:    config,
		CreatedAt: dbSetting.CreatedAt.Time,
		UpdatedAt: dbSetting.UpdatedAt.Time,
	}, nil
}

// GetSetting retrieves a setting by ID
func (s *SettingsService) GetSetting(ctx context.Context, id int64) (*Setting, error) {
	dbSetting, err := s.queries.GetSetting(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(dbSetting.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &Setting{
		ID:        dbSetting.ID,
		Config:    config,
		CreatedAt: dbSetting.CreatedAt.Time,
		UpdatedAt: dbSetting.UpdatedAt.Time,
	}, nil
}

// ListSettings retrieves all settings
func (s *SettingsService) ListSettings(ctx context.Context) ([]*Setting, error) {
	dbSettings, err := s.queries.ListSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list settings: %w", err)
	}

	settings := make([]*Setting, len(dbSettings))
	for i, dbSetting := range dbSettings {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(dbSetting.Config), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		settings[i] = &Setting{
			ID:        dbSetting.ID,
			Config:    config,
			CreatedAt: dbSetting.CreatedAt.Time,
			UpdatedAt: dbSetting.UpdatedAt.Time,
		}
	}

	return settings, nil
}

// UpdateSetting updates a setting
func (s *SettingsService) UpdateSetting(ctx context.Context, id int64, params UpdateSettingParams) (*Setting, error) {
	configJSON, err := json.Marshal(params.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	dbSetting, err := s.queries.UpdateSetting(ctx, &db.UpdateSettingParams{
		ID:     id,
		Config: string(configJSON),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update setting: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(dbSetting.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &Setting{
		ID:        dbSetting.ID,
		Config:    config,
		CreatedAt: dbSetting.CreatedAt.Time,
		UpdatedAt: dbSetting.UpdatedAt.Time,
	}, nil
}

// DeleteSetting deletes a setting
func (s *SettingsService) DeleteSetting(ctx context.Context, id int64) error {
	err := s.queries.DeleteSetting(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}
	return nil
}
