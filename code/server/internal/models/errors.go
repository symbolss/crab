package models

import "errors"

// Validation errors
var (
	ErrParentNameRequired  = errors.New("parent name is required")
	ErrParentNameTooLong   = errors.New("parent name must be less than 100 characters")
	ErrPairingCodeRequired = errors.New("pairing code is required")
	ErrInvalidPairingCode  = errors.New("pairing code must be 8 characters")
	ErrChildNameRequired   = errors.New("child name is required")
	ErrChildNameTooLong    = errors.New("child name must be less than 100 characters")
)

// Business errors
var (
	ErrFamilyNotFound          = errors.New("family not found")
	ErrPairingCodeInvalid      = errors.New("invalid or expired pairing code")
	ErrPairingCodeExpired      = errors.New("pairing code has expired")
	ErrChildNotFound           = errors.New("child not found")
	ErrUnauthorized            = errors.New("unauthorized access")
)
