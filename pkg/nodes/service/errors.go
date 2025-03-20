package service

import "errors"

var (
	// ErrNotFound is returned when a node is not found
	ErrNotFound = errors.New("node not found")

	// ErrInvalidNodeType is returned when an invalid node type is provided
	ErrInvalidNodeType = errors.New("invalid node type")

	// ErrInvalidPlatform is returned when an invalid blockchain platform is provided
	ErrInvalidPlatform = errors.New("invalid blockchain platform")

	// ErrNodeAlreadyExists is returned when a node with the same name already exists
	ErrNodeAlreadyExists = errors.New("node already exists")

	// ErrNodeNotRunning is returned when trying to perform an operation on a non-running node
	ErrNodeNotRunning = errors.New("node is not running")

	// ErrNodeAlreadyRunning is returned when trying to start an already running node
	ErrNodeAlreadyRunning = errors.New("node is already running")
)
