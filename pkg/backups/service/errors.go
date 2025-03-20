package service

import "errors"

var (
	// ErrTargetNotFound is returned when a backup target is not found
	ErrTargetNotFound = errors.New("backup target not found")

	// ErrScheduleNotFound is returned when a backup schedule is not found
	ErrScheduleNotFound = errors.New("backup schedule not found")

	// ErrBackupNotFound is returned when a backup is not found
	ErrBackupNotFound = errors.New("backup not found")

	// ErrInvalidBackupType is returned when an invalid backup type is provided
	ErrInvalidBackupType = errors.New("invalid backup type")

	// ErrScheduleAlreadyEnabled is returned when trying to enable an already enabled schedule
	ErrScheduleAlreadyEnabled = errors.New("schedule is already enabled")

	// ErrScheduleAlreadyDisabled is returned when trying to disable an already disabled schedule
	ErrScheduleAlreadyDisabled = errors.New("schedule is already disabled")
)
