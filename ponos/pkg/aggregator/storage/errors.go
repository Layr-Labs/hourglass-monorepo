package storage

import "errors"

var (
	// ErrNotFound is returned when a requested item is not found in storage
	ErrNotFound = errors.New("item not found")

	// ErrAlreadyExists is returned when attempting to create an item that already exists
	ErrAlreadyExists = errors.New("item already exists")

	// ErrStoreClosed is returned when attempting to use a closed storage instance
	ErrStoreClosed = errors.New("storage is closed")

	// ErrInvalidTaskStatus is returned when an invalid task status transition is attempted
	ErrInvalidTaskStatus = errors.New("invalid task status")

	// ErrInvalidChainId is returned when an invalid chain ID is provided
	ErrInvalidChainId = errors.New("invalid chain ID")
)