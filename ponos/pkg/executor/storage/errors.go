package storage

import "errors"

var (
	// ErrNotFound is returned when a requested item is not found in storage
	ErrNotFound = errors.New("item not found")

	// ErrAlreadyExists is returned when attempting to create an item that already exists
	ErrAlreadyExists = errors.New("item already exists")

	// ErrStoreClosed is returned when attempting to use a closed storage instance
	ErrStoreClosed = errors.New("storage is closed")

	// ErrInvalidDeploymentStatus is returned when an invalid deployment status transition is attempted
	ErrInvalidDeploymentStatus = errors.New("invalid deployment status")

	// ErrPerformerNotFound is returned when a performer is not found
	ErrPerformerNotFound = errors.New("performer not found")

	// ErrTaskNotFound is returned when a task is not found
	ErrTaskNotFound = errors.New("task not found")
)
