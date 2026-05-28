package services_test

import (
	"context"
	"testing"

	"github.com/family/crab-server/internal/models"
	memrepo "github.com/family/crab-server/internal/repository/memory"
	"github.com/family/crab-server/internal/services"
	"github.com/google/uuid"
)

func TestFamilyService_ValidateToken_WrongSecret(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	// Create a token with a different service
	otherSvc := services.NewFamilyService(familyRepo, childRepo, "different-secret-key-32-chars-minimum")
	resp, err := otherSvc.CreateFamily(nil, &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Validate with original service (should fail due to different secret)
	_, err = svc.ValidateToken(resp.ParentToken)
	if err == nil {
		t.Error("ValidateToken() should fail with different secret")
	}
}

func TestFamilyService_GetChildrenByFamilyID_WithChildren(t *testing.T) {
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
	for i := 0; i < 5; i++ {
		_, err := svc.PairChild(ctx, &models.PairChildRequest{
			PairingCode: familyResp.PairingCode,
			ChildName:   "Child",
		})
		if err != nil {
			t.Fatalf("PairChild() error = %v", err)
		}
	}

	children, err := svc.GetChildrenByFamilyID(ctx, familyResp.FamilyID)
	if err != nil {
		t.Fatalf("GetChildrenByFamilyID() error = %v", err)
	}

	if len(children) != 5 {
		t.Errorf("len(children) = %d, want 5", len(children))
	}
}

func TestFamilyService_CreateFamily_LongParentName(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()

	// Create a name that is exactly 100 characters
	longName := make([]byte, 100)
	for i := range longName {
		longName[i] = 'A'
	}

	req := &models.CreateFamilyRequest{ParentName: string(longName)}

	resp, err := svc.CreateFamily(ctx, req)
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	if resp.FamilyID == uuid.Nil {
		t.Error("FamilyID should not be nil")
	}
}

func TestFamilyService_PairChild_LongChildName(t *testing.T) {
	familyRepo := memrepo.NewInMemoryFamilyRepository()
	childRepo := memrepo.NewInMemoryChildRepository()
	svc := services.NewFamilyService(familyRepo, childRepo, testJWTSecret)

	ctx := context.Background()

	// Create a family
	familyResp, err := svc.CreateFamily(ctx, &models.CreateFamilyRequest{ParentName: "John Doe"})
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	// Create a name that is exactly 100 characters
	longName := make([]byte, 100)
	for i := range longName {
		longName[i] = 'B'
	}

	req := &models.PairChildRequest{
		PairingCode: familyResp.PairingCode,
		ChildName:   string(longName),
	}

	resp, err := svc.PairChild(ctx, req)
	if err != nil {
		t.Fatalf("PairChild() error = %v", err)
	}

	if resp.ChildName != string(longName) {
		t.Errorf("ChildName = %v, want %v", resp.ChildName, string(longName))
	}
}
