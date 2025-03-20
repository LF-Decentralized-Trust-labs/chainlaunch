package service

import (
	"time"
)

// BackupTargetType represents the type of backup target
type BackupTargetType string

const (
	BackupTargetTypeS3 BackupTargetType = "S3"
)

// BackupStatus represents the status of a backup
type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "PENDING"
	BackupStatusInProgress BackupStatus = "IN_PROGRESS"
	BackupStatusCompleted  BackupStatus = "COMPLETED"
	BackupStatusFailed     BackupStatus = "FAILED"
)

// S3Config represents the configuration for an S3 backup target
type S3Config struct {
	BucketName     string  `json:"bucketName"`
	Region         string  `json:"region"`
	Endpoint       *string `json:"endpoint,omitempty"`
	BucketPath     string  `json:"bucketPath"`
	AccessKeyID    string  `json:"accessKeyId"`
	SecretKey      string  `json:"secretKey"`
	ForcePathStyle *bool   `json:"forcePathStyle,omitempty"`
}

// ResticConfig represents the configuration for a Restic backup target
type ResticConfig struct {
	Password string `json:"password"`
	// Add other Restic-specific configuration options
}

// BackupTargetDTO represents a backup target
type BackupTargetDTO struct {
	ID             int64            `json:"id"`
	Name           string           `json:"name"`
	Type           BackupTargetType `json:"type"`
	BucketName     string           `json:"bucketName,omitempty"`
	Region         string           `json:"region,omitempty"`
	Endpoint       string           `json:"endpoint,omitempty"`
	BucketPath     string           `json:"bucketPath,omitempty"`
	AccessKeyID    string           `json:"accessKeyId,omitempty"`
	ForcePathStyle bool             `json:"forcePathStyle,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      *time.Time       `json:"updatedAt,omitempty"`
}

// BackupScheduleDTO represents a backup schedule
type BackupScheduleDTO struct {
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

// BackupDTO represents a backup
type BackupDTO struct {
	ID           int64        `json:"id"`
	ScheduleID   *int64       `json:"scheduleId,omitempty"`
	TargetID     int64        `json:"targetId"`
	Status       BackupStatus `json:"status"`
	SizeBytes    *int64       `json:"sizeBytes,omitempty"`
	StartedAt    time.Time    `json:"startedAt"`
	CompletedAt  *time.Time   `json:"completedAt,omitempty"`
	ErrorMessage *string      `json:"errorMessage,omitempty"`
	Metadata     interface{}  `json:"metadata,omitempty"`
	CreatedAt    time.Time    `json:"createdAt"`
}

// CreateBackupTargetParams represents parameters for creating a backup target
type CreateBackupTargetParams struct {
	Name           string           `validate:"required"`
	Type           BackupTargetType `validate:"required,oneof=S3 LOCAL"`
	BucketName     string           `validate:"required_if=Type S3"`
	BucketPath     string           `validate:"required_if=Type S3"`
	Endpoint       string           `validate:"required_if=Type S3,url"`
	AccessKeyID    string           `validate:"required_if=Type S3"`
	SecretKey      string           `validate:"required_if=Type S3"`
	Region         string           `validate:"required_if=Type S3"`
	ForcePathStyle bool
}

// CreateBackupScheduleParams represents parameters for creating a backup schedule
type CreateBackupScheduleParams struct {
	Name           string `validate:"required"`
	Description    string
	CronExpression string `validate:"required"`
	TargetID       int64  `validate:"required"`
	RetentionDays  int    `validate:"required,min=1"`
	Enabled        bool
}

// CreateBackupParams represents parameters for creating a backup
type CreateBackupParams struct {
	ScheduleID *int64
	TargetID   int64       `validate:"required"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

// UpdateBackupTargetParams represents parameters for updating a backup target
type UpdateBackupTargetParams struct {
	ID             int64            `validate:"required"`
	Name           string           `validate:"required"`
	Type           BackupTargetType `validate:"required,oneof=S3 LOCAL"`
	BucketName     string           `validate:"required_if=Type S3"`
	BucketPath     string           `validate:"required_if=Type S3"`
	Endpoint       string           `validate:"required_if=Type S3,url"`
	AccessKeyID    string           `validate:"required_if=Type S3"`
	SecretKey      string           `validate:"required_if=Type S3"`
	Region         string           `validate:"required_if=Type S3"`
	ForcePathStyle bool
}

// UpdateBackupScheduleParams represents parameters for updating a backup schedule
type UpdateBackupScheduleParams struct {
	ID             int64  `validate:"required"`
	Name           string `validate:"required"`
	Description    string
	CronExpression string `validate:"required"`
	TargetID       int64  `validate:"required"`
	RetentionDays  int    `validate:"required,min=1"`
	Enabled        bool
}
