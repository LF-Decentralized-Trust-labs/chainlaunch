package service

import "time"

// Template represents a node configuration template
type Template struct {
	ID          int64                  `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Template    map[string]interface{} `json:"template"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// CreateTemplateParams represents parameters for creating a template
type CreateTemplateParams struct {
	Type        string                 `json:"type" validate:"required"`
	Name        string                 `json:"name" validate:"required"`
	Description string                 `json:"description"`
	Template    map[string]interface{} `json:"template" validate:"required"`
}

// UpdateTemplateParams represents parameters for updating a template
type UpdateTemplateParams struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Template    map[string]interface{} `json:"template,omitempty"`
}
