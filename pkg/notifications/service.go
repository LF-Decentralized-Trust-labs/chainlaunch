package notifications

import (
	"context"
)

// Service defines the interface for notification services
type Service interface {
	// SendNotification sends a notification of the specified type with the provided data
	SendNotification(ctx context.Context, notificationType NotificationType, data interface{}) error

	// SendBackupSuccessNotification sends a notification about a successful backup
	SendBackupSuccessNotification(ctx context.Context, data BackupSuccessData) error

	// SendBackupFailureNotification sends a notification about a failed backup
	SendBackupFailureNotification(ctx context.Context, data BackupFailureData) error

	// SendS3ConnectionIssueNotification sends a notification about S3 connection issues
	SendS3ConnectionIssueNotification(ctx context.Context, data S3ConnectionIssueData) error

	// SendNodeDowntimeNotification sends a notification about node downtime
	SendNodeDowntimeNotification(ctx context.Context, data NodeDowntimeData) error
}
