package models

import (
	"time"

	"github.com/google/uuid"
)

// Family represents a family unit in the system
type Family struct {
	ID                    uuid.UUID `json:"id"`
	PairingCode           string    `json:"pairingCode"`
	PairingCodeExpiresAt  time.Time `json:"pairingCodeExpiresAt"`
	CreatedAt             time.Time `json:"createdAt"`
}

// Child represents a child account linked to a family
type Child struct {
	ID               uuid.UUID `json:"id"`
	FamilyID         uuid.UUID `json:"familyId"`
	Name             string    `json:"name"`
	AvatarURL        string    `json:"avatarUrl,omitempty"`
	DeviceID         string    `json:"deviceId,omitempty"`
	DailyTimeLimitSec int      `json:"dailyTimeLimitSec"` // Daily time limit in seconds, default 3600
	Status           string    `json:"status"`            // active, inactive
	CreatedAt        time.Time `json:"createdAt"`
}

// CreateFamilyRequest represents the request body for creating a family
type CreateFamilyRequest struct {
	// ParentName is optional, used for display purposes
	ParentName string `json:"parentName,omitempty"`
}

// CreateFamilyResponse represents the response for creating a family
type CreateFamilyResponse struct {
	FamilyID    uuid.UUID `json:"familyId"`
	PairingCode string    `json:"pairingCode"`
	ParentToken string    `json:"parentToken"`
}

// PairChildRequest represents the request body for pairing a child
type PairChildRequest struct {
	PairingCode string `json:"pairingCode"`
	ChildName   string `json:"childName"`
	DeviceID    string `json:"deviceId,omitempty"` // Optional device identifier
}

// PairChildResponse represents the response for pairing a child
type PairChildResponse struct {
	ChildID    uuid.UUID `json:"childId"`
	FamilyID   uuid.UUID `json:"familyId"`
	ChildName  string    `json:"childName"`
	ChildToken string    `json:"childToken"`
}

// ChildListResponse represents a child in the list response
type ChildListResponse struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	AvatarURL        string    `json:"avatarUrl,omitempty"`
	DeviceID         string    `json:"deviceId,omitempty"`
	DailyTimeLimitSec int      `json:"dailyTimeLimitSec"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
}

// Validate validates the CreateFamilyRequest - ParentName is optional
func (r *CreateFamilyRequest) Validate() error {
	if len(r.ParentName) > 100 {
		return ErrParentNameTooLong
	}
	return nil
}

// Validate validates the PairChildRequest
func (r *PairChildRequest) Validate() error {
	if r.PairingCode == "" {
		return ErrPairingCodeRequired
	}
	if len(r.PairingCode) != 8 {
		return ErrInvalidPairingCode
	}
	if r.ChildName == "" {
		return ErrChildNameRequired
	}
	if len(r.ChildName) > 100 {
		return ErrChildNameTooLong
	}
	return nil
}

// IsPairingCodeExpired checks if the pairing code has expired
func (f *Family) IsPairingCodeExpired() bool {
	return time.Now().After(f.PairingCodeExpiresAt)
}
