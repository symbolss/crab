package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/family/crab-server/internal/models"
	memrepo "github.com/family/crab-server/internal/repository/memory"
	"github.com/family/crab-server/internal/services"
	"github.com/google/uuid"
)

const testJWTSecret = "test-secret-key-minimum-32-characters"

func TestFamilyService_CreateFamily(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()
	req := &models.CreateFamilyRequest{ParentName: "John Doe"}

	resp, err := svc.CreateFamily(ctx, req)
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Verify response
	if resp.FamilyID == uuid.Nil {
		t.Error("FamilyID should not be nil")
	}
	if len(resp.PairingCode) != 8 {
		t.Errorf("PairingCode length = %d, want 8", len(resp.PairingCode))
	}
	if resp.ParentToken == "" {
		t.Error("ParentToken should not be empty")
	}

	// Verify family was stored
	family, err := familyRepo.GetFamilyByID(ctx, resp.FamilyID)
	if err != nil {
		t.Fatalf("GetFamilyByID() error = %v", err)
	}
	if family.PairingCode != resp.PairingCode {
		t.Errorf("stored PairingCode = %v, want %v", family.PairingCode, resp.PairingCode)
	}
}

func TestFamilyService_CreateFamily_GeneratesUniquePairingCodes(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()

	resp1, err := svc.CreateFamily(ctx, &models.CreateFamilyRequest{ParentName: "Parent 1"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	resp2, err := svc.CreateFamily(ctx, &models.CreateFamilyRequest{ParentName: "Parent 2"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	if resp1.PairingCode == resp2.PairingCode {
		t.Error("Pairing codes should be unique")
	}
}

func TestFamilyService_PairChild(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()

	// First create a family
	familyResp, err := svc.CreateFamily(ctx, &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Pair a child
	req := &models.PairChildRequest{
		PairingCode: familyResp.PairingCode,
		ChildName:   "Little John",
	}

	resp, err := svc.PairChild(ctx, req)
	if err != nil {
		t.Fatalf("PairChild() error = %v", err)
	}

	// Verify response
	if resp.ChildID == uuid.Nil {
		t.Error("ChildID should not be nil")
	}
	if resp.FamilyID != familyResp.FamilyID {
		t.Errorf("FamilyID = %v, want %v", resp.FamilyID, familyResp.FamilyID)
	}
	if resp.ChildName != "Little John" {
		t.Errorf("ChildName = %v, want Little John", resp.ChildName)
	}
	if resp.ChildToken == "" {
		t.Error("ChildToken should not be empty")
	}

	// Verify child was stored
	children, err := childRepo.GetChildrenByFamilyID(ctx, familyResp.FamilyID)
	if err != nil {
		t.Fatalf("GetChildrenByFamilyID() error = %v", err)
	}
	if len(children) != 1 {
		t.Errorf("len(children) = %d, want 1", len(children))
	}
}

func TestFamilyService_PairChild_InvalidPairingCode(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()

	req := &models.PairChildRequest{
		PairingCode: "INVALID1",
		ChildName:   "Little John",
	}

	_, err := svc.PairChild(ctx, req)
	if !errors.Is(err, models.ErrPairingCodeInvalid) {
		t.Errorf("PairChild() error = %v, want %v", err, models.ErrPairingCodeInvalid)
	}
}

func TestFamilyService_GetChildrenByFamilyID(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()

	// Create a family
	familyResp, err := svc.CreateFamily(ctx, &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Pair multiple children
	for i := 0; i < 3; i++ {
		_, err := svc.PairChild(ctx, &models.PairChildRequest{
			PairingCode: familyResp.PairingCode,
			ChildName:   "Child " + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("PairChild() error = %v", err)
		}
	}

	// Get children
	children, err := svc.GetChildrenByFamilyID(ctx, familyResp.FamilyID)
	if err != nil {
		t.Fatalf("GetChildrenByFamilyID() error = %v", err)
	}

	if len(children) != 3 {
		t.Errorf("len(children) = %d, want 3", len(children))
	}

	// Verify children have correct structure
	for _, c := range children {
		if c.ID == uuid.Nil {
			t.Error("Child ID should not be nil")
		}
		if c.Name == "" {
			t.Error("Child Name should not be empty")
		}
		if c.Status != "active" {
			t.Errorf("Child Status = %v, want active", c.Status)
		}
		if c.CreatedAt.IsZero() {
			t.Error("Child CreatedAt should not be zero")
		}
	}
}

func TestFamilyService_GetChildrenByFamilyID_Empty(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()

	children, err := svc.GetChildrenByFamilyID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("GetChildrenByFamilyID() error = %v", err)
	}

	if len(children) != 0 {
		t.Errorf("len(children) = %d, want 0", len(children))
	}
}

func TestFamilyService_ValidateToken(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()

	// Create a family and get a token
	resp, err := svc.CreateFamily(ctx, &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Validate the token
	claims, err := svc.ValidateToken(resp.ParentToken)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	// Verify claims
	if claims["sub"] != resp.FamilyID.String() {
		t.Errorf("sub claim = %v, want %v", claims["sub"], resp.FamilyID.String())
	}
	if claims["role"] != "parent" {
		t.Errorf("role claim = %v, want parent", claims["role"])
	}
}

func TestFamilyService_ValidateToken_Invalid(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	// Test with invalid token
	_, err := svc.ValidateToken("invalid-token")
	if err == nil {
		t.Error("ValidateToken() should return error for invalid token")
	}
}

func TestFamilyService_ValidateToken_ChildToken(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()

	// Create a family
	familyResp, err := svc.CreateFamily(ctx, &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Pair a child
	childResp, err := svc.PairChild(ctx, &models.PairChildRequest{
		PairingCode: familyResp.PairingCode,
		ChildName:   "Little John",
	})
	if err != nil {
		t.Fatalf("PairChild() error = %v", err)
	}

	// Validate the child token
	claims, err := svc.ValidateToken(childResp.ChildToken)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	// Verify claims
	if claims["sub"] != childResp.ChildID.String() {
		t.Errorf("sub claim = %v, want %v", claims["sub"], childResp.ChildID.String())
	}
	if claims["role"] != "child" {
		t.Errorf("role claim = %v, want child", claims["role"])
	}
}

func TestFamilyService_TokenExpiration(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()

	// Create a family
	resp, err := svc.CreateFamily(ctx, &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Validate token and check expiration
	claims, err := svc.ValidateToken(resp.ParentToken)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("exp claim should be a number")
	}

	// Token should expire in approximately 24 hours
	expTime := time.Unix(int64(exp), 0)
	expectedExp := time.Now().Add(24 * time.Hour)

	// Allow 1 minute tolerance
	diff := expTime.Sub(expectedExp)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("token expiration diff = %v, expected close to 0", diff)
	}
}
