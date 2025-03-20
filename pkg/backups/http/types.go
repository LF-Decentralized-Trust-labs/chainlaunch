package http

import (
	"time"
)

// CreateBackupTargetRequest represents the HTTP request for creating a backup target
// @Description Request body for creating a new backup target
type CreateBackupTargetRequest struct {
	// Name of the backup target
	// @Example "daily-backup-s3"
	Name string `json:"name" validate:"required"`
	// Type of backup target (S3 or LOCAL)
	// @Example "S3"
	Type string `json:"type" validate:"required,oneof=S3 LOCAL"`
	// S3 bucket name (required for S3 type)
	// @Example "my-backup-bucket"
	BucketName string `json:"bucketName,omitempty" validate:"required_if=Type S3"`
	// AWS region (required for S3 type)
	// @Example "us-east-1"
	Region string `json:"region,omitempty" validate:"required_if=Type S3"`
	// Custom S3 endpoint (optional)
	// @Example "https://s3.custom-domain.com"
	Endpoint string `json:"endpoint,omitempty"`
	// Path within the bucket (required for S3 type)
	// @Example "backups/app1"
	BucketPath string `json:"bucketPath,omitempty" validate:"required_if=Type S3"`
	// AWS access key ID (required for S3 type)
	// @Example "AKIAXXXXXXXXXXXXXXXX"
	AccessKeyID string `json:"accessKeyId,omitempty" validate:"required_if=Type S3"`
	// AWS secret key (required for S3 type)
	SecretKey string `json:"secretKey,omitempty" validate:"required_if=Type S3"`
	// Use path-style S3 URLs
	// @Example false
	ForcePathStyle bool `json:"forcePathStyle,omitempty"`
}

// CreateBackupScheduleRequest represents the HTTP request for creating a backup schedule
// @Description Request body for creating a new backup schedule
type CreateBackupScheduleRequest struct {
	// Name of the backup schedule
	// @Example "daily-backup"
	Name string `json:"name" validate:"required"`
	// Optional description
	// @Example "Daily backup at midnight"
	Description string `json:"description"`
	// Cron expression for schedule
	// @Example "0 0 * * *"
	CronExpression string `json:"cronExpression" validate:"required"`
	// ID of the backup target to use
	// @Example 1
	TargetID int64 `json:"targetId" validate:"required"`
	// Number of days to retain backups
	// @Example 30
	RetentionDays int `json:"retentionDays" validate:"required,min=1"`
	// Whether the schedule is enabled
	// @Example true
	Enabled bool `json:"enabled"`
}

// BackupTargetResponse represents the HTTP response for a backup target
type BackupTargetResponse struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	BucketName     string     `json:"bucketName,omitempty"`
	Region         string     `json:"region,omitempty"`
	Endpoint       string     `json:"endpoint,omitempty"`
	BucketPath     string     `json:"bucketPath,omitempty"`
	AccessKeyID    string     `json:"accessKeyId,omitempty"`
	ForcePathStyle bool       `json:"forcePathStyle,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
}

// BackupScheduleResponse represents the HTTP response for a backup schedule
type BackupScheduleResponse struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	CronExpression string     `json:"cronExpression"`
	TargetID       int64      `json:"targetId"`
	RetentionDays  int        `json:"retentionDays"`
	Enabled        bool       `json:"enabled"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
	LastRunAt      *time.Time `json:"lastRunAt,omitempty"`
	NextRunAt      *time.Time `json:"nextRunAt,omitempty"`
}

// Add S3Config type to match the diesel schema
type S3Config struct {
	BucketName     string  `json:"bucketName" validate:"required"`
	Region         string  `json:"region" validate:"required"`
	Endpoint       *string `json:"endpoint,omitempty"`
	BucketPath     string  `json:"bucketPath" validate:"required"`
	AccessKeyID    string  `json:"accessKeyId" validate:"required"`
	SecretKey      string  `json:"secretKey" validate:"required"`
	ForcePathStyle *bool   `json:"forcePathStyle,omitempty"`
}

// ErrorResponse represents an error response
// @Description Error response from the API
type ErrorResponse struct {
	// Error message
	// @Example "Invalid request parameters"
	Error string `json:"error"`
}

// Update CreateBackupRequest
type CreateBackupRequest struct {
	ScheduleID *int64      `json:"scheduleId,omitempty"`
	TargetID   int64       `json:"targetId" validate:"required"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

// Update BackupResponse
type BackupResponse struct {
	ID           int64       `json:"id"`
	ScheduleID   *int64      `json:"scheduleId,omitempty"`
	TargetID     int64       `json:"targetId"`
	Status       string      `json:"status"`
	SizeBytes    *int64      `json:"sizeBytes,omitempty"`
	StartedAt    time.Time   `json:"startedAt"`
	CompletedAt  *time.Time  `json:"completedAt,omitempty"`
	ErrorMessage *string     `json:"errorMessage,omitempty"`
	Metadata     interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time   `json:"createdAt"`
}

// UpdateBackupTargetRequest represents the HTTP request for updating a backup target
type UpdateBackupTargetRequest struct {
	Name           string `json:"name" validate:"required"`
	Type           string `json:"type" validate:"required,oneof=S3 LOCAL"`
	BucketName     string `json:"bucketName,omitempty" validate:"required_if=Type S3"`
	Region         string `json:"region,omitempty" validate:"required_if=Type S3"`
	Endpoint       string `json:"endpoint,omitempty"`
	BucketPath     string `json:"bucketPath,omitempty" validate:"required_if=Type S3"`
	AccessKeyID    string `json:"accessKeyId,omitempty" validate:"required_if=Type S3"`
	SecretKey      string `json:"secretKey,omitempty" validate:"required_if=Type S3"`
	ForcePathStyle bool   `json:"forcePathStyle,omitempty"`
}

// UpdateBackupScheduleRequest represents the HTTP request for updating a backup schedule
type UpdateBackupScheduleRequest struct {
	Name           string `json:"name" validate:"required"`
	Description    string `json:"description"`
	CronExpression string `json:"cronExpression" validate:"required"`
	TargetID       int64  `json:"targetId" validate:"required"`
	RetentionDays  int    `json:"retentionDays" validate:"required,min=1"`
	Enabled        bool   `json:"enabled"`
}
