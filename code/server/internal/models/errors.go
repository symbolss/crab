package models

import "errors"

// Validation errors
var (
	ErrParentNameTooLong   = errors.New("parent name must be less than 100 characters")
	ErrPairingCodeRequired = errors.New("pairing code is required")
	ErrInvalidPairingCode  = errors.New("pairing code must be 8 characters")
	ErrChildNameRequired   = errors.New("child name is required")
	ErrChildNameTooLong    = errors.New("child name must be less than 100 characters")
	ErrSourceURLRequired   = errors.New("source url is required")
	ErrSourceURLTooLong    = errors.New("source url must be less than 2048 characters")
	ErrInvalidChildID      = errors.New("invalid child id")
	ErrLinkUnsupported     = errors.New("link source is not supported")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrUsageExceeded       = errors.New("daily usage limit exceeded")
)

// Business errors
var (
	ErrFamilyNotFound          = errors.New("family not found")
	ErrPairingCodeInvalid      = errors.New("invalid or expired pairing code")
	ErrPairingCodeExpired      = errors.New("pairing code has expired")
	ErrChildNotFound           = errors.New("child not found")
	ErrUnauthorized            = errors.New("unauthorized access")
	ErrItemNotFound            = errors.New("item not found")
	ErrDuplicateURL            = errors.New("content with this url already exists")
	ErrChildNotInFamily        = errors.New("child does not belong to this family")
)
