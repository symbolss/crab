package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/family/crab-server/internal/models"
	memrepo "github.com/family/crab-server/internal/repository/memory"
	"github.com/family/crab-server/internal/services"
)

func TestFamilyService_TokenClaims(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	// Create a family
	resp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Validate token and check all claims
	claims, err := svc.ValidateToken(resp.ParentToken)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	// Check iat claim
	iat, ok := claims["iat"].(float64)
	if !ok {
		t.Fatal("iat claim should be a number")
	}
	iatTime := time.Unix(int64(iat), 0)
	if iatTime.After(time.Now()) {
		t.Error("iat should be in the past")
	}

	// Check exp claim
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("exp claim should be a number")
	}
	if int64(exp) <= int64(iat) {
		t.Error("exp should be greater than iat")
	}
}

func TestFamilyService_MultipleFamilies(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	// Create multiple families
	family1, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "Parent 1"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	family2, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "Parent 2"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Ensure families have different IDs
	if family1.FamilyID == family2.FamilyID {
		t.Error("Families should have different IDs")
	}

	// Pair children to different families
	child1, err := svc.PairChild(context.Background(), &models.PairChildRequest{
		PairingCode: family1.PairingCode,
		ChildName:   "Child 1",
	})
	if err != nil {
		t.Fatalf("PairChild() error = %v", err)
	}

	child2, err := svc.PairChild(context.Background(), &models.PairChildRequest{
		PairingCode: family2.PairingCode,
		ChildName:   "Child 2",
	})
	if err != nil {
		t.Fatalf("PairChild() error = %v", err)
	}

	// Ensure children belong to correct families
	if child1.FamilyID != family1.FamilyID {
		t.Error("Child 1 should belong to family 1")
	}
	if child2.FamilyID != family2.FamilyID {
		t.Error("Child 2 should belong to family 2")
	}
}

func TestFamilyService_ChildTokenValidation(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	// Create family and child
	familyResp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	childResp, err := svc.PairChild(context.Background(), &models.PairChildRequest{
		PairingCode: familyResp.PairingCode,
		ChildName:   "Little John",
	})
	if err != nil {
		t.Fatalf("PairChild() error = %v", err)
	}

	// Validate child token
	claims, err := svc.ValidateToken(childResp.ChildToken)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	// Check that role is child
	if claims["role"] != "child" {
		t.Errorf("role = %v, want child", claims["role"])
	}

	// Check that sub matches child ID
	if claims["sub"] != childResp.ChildID.String() {
		t.Errorf("sub = %v, want %v", claims["sub"], childResp.ChildID.String())
	}
}

func TestFamilyService_GetChildren_EmptyFamily(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	// Create family without children
	familyResp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Get children (should be empty)
	children, err := svc.GetChildrenByFamilyID(context.Background(), familyResp.FamilyID)
	if err != nil {
		t.Fatalf("GetChildrenByFamilyID() error = %v", err)
	}

	if len(children) != 0 {
		t.Errorf("len(children) = %d, want 0", len(children))
	}
}

func TestFamilyService_PairingCodeExpiry(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	// Create a family
	familyResp, err := svc.CreateFamily(context.Background(), &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Verify the family has an expiry time set
	family, err := familyRepo.GetFamilyByID(context.Background(), familyResp.FamilyID)
	if err != nil {
		t.Fatalf("GetFamilyByID() error = %v", err)
	}

	if family.PairingCodeExpiresAt.IsZero() {
		t.Error("PairingCodeExpiresAt should be set")
	}

	// Verify expiry is approximately 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	diff := family.PairingCodeExpiresAt.Sub(expectedExpiry)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("PairingCodeExpiresAt diff = %v, expected close to 0", diff)
	}
}
