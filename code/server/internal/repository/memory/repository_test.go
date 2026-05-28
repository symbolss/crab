package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/family/crab-server/internal/models"
	"github.com/family/crab-server/internal/repository/memory"
	"github.com/google/uuid"
)

func TestInMemoryFamilyRepository_CreateFamily(t *testing.T) {
	repo := memory.NewInMemoryFamilyRepository()
	ctx := context.Background()

	family := &models.Family{
		ID:          uuid.New(),
		PairingCode: "ABCD1234",
		CreatedAt:   time.Now(),
	}

	err := repo.CreateFamily(ctx, family)
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}
}

func TestInMemoryFamilyRepository_GetFamilyByID(t *testing.T) {
	repo := memory.NewInMemoryFamilyRepository()
	ctx := context.Background()

	family := &models.Family{
		ID:          uuid.New(),
		PairingCode: "ABCD1234",
		CreatedAt:   time.Now(),
	}

	_ = repo.CreateFamily(ctx, family)

	retrieved, err := repo.GetFamilyByID(ctx, family.ID)
	if err != nil {
		t.Fatalf("GetFamilyByID() error = %v", err)
	}

	if retrieved.ID != family.ID {
		t.Errorf("retrieved.ID = %v, want %v", retrieved.ID, family.ID)
	}
}

func TestInMemoryFamilyRepository_GetFamilyByID_NotFound(t *testing.T) {
	repo := memory.NewInMemoryFamilyRepository()
	ctx := context.Background()

	_, err := repo.GetFamilyByID(ctx, uuid.New())
	if err != models.ErrFamilyNotFound {
		t.Errorf("GetFamilyByID() error = %v, want %v", err, models.ErrFamilyNotFound)
	}
}

func TestInMemoryFamilyRepository_GetFamilyByPairingCode(t *testing.T) {
	repo := memory.NewInMemoryFamilyRepository()
	ctx := context.Background()

	family := &models.Family{
		ID:          uuid.New(),
		PairingCode: "TEST5678",
		CreatedAt:   time.Now(),
	}

	_ = repo.CreateFamily(ctx, family)

	retrieved, err := repo.GetFamilyByPairingCode(ctx, "TEST5678")
	if err != nil {
		t.Fatalf("GetFamilyByPairingCode() error = %v", err)
	}

	if retrieved.ID != family.ID {
		t.Errorf("retrieved.ID = %v, want %v", retrieved.ID, family.ID)
	}
}

func TestInMemoryFamilyRepository_GetFamilyByPairingCode_Invalid(t *testing.T) {
	repo := memory.NewInMemoryFamilyRepository()
	ctx := context.Background()

	_, err := repo.GetFamilyByPairingCode(ctx, "INVALID1")
	if err != models.ErrPairingCodeInvalid {
		t.Errorf("GetFamilyByPairingCode() error = %v, want %v", err, models.ErrPairingCodeInvalid)
	}
}

func TestInMemoryFamilyRepository_UpdatePairingCode(t *testing.T) {
	repo := memory.NewInMemoryFamilyRepository()
	ctx := context.Background()

	family := &models.Family{
		ID:          uuid.New(),
		PairingCode: "OLD1234",
		CreatedAt:   time.Now(),
	}

	_ = repo.CreateFamily(ctx, family)

	err := repo.UpdatePairingCode(ctx, family.ID, "NEW5678")
	if err != nil {
		t.Fatalf("UpdatePairingCode() error = %v", err)
	}

	// Verify old code no longer works
	_, err = repo.GetFamilyByPairingCode(ctx, "OLD1234")
	if err != models.ErrPairingCodeInvalid {
		t.Errorf("GetFamilyByPairingCode(old) error = %v, want %v", err, models.ErrPairingCodeInvalid)
	}

	// Verify new code works
	retrieved, err := repo.GetFamilyByPairingCode(ctx, "NEW5678")
	if err != nil {
		t.Fatalf("GetFamilyByPairingCode(new) error = %v", err)
	}

	if retrieved.ID != family.ID {
		t.Errorf("retrieved.ID = %v, want %v", retrieved.ID, family.ID)
	}
}

func TestInMemoryFamilyRepository_UpdatePairingCode_NotFound(t *testing.T) {
	repo := memory.NewInMemoryFamilyRepository()
	ctx := context.Background()

	err := repo.UpdatePairingCode(ctx, uuid.New(), "NEW5678")
	if err != models.ErrFamilyNotFound {
		t.Errorf("UpdatePairingCode() error = %v, want %v", err, models.ErrFamilyNotFound)
	}
}

func TestInMemoryChildRepository_CreateChild(t *testing.T) {
	repo := memory.NewInMemoryChildRepository()
	ctx := context.Background()
	familyID := uuid.New()

	child := &models.Child{
		ID:        uuid.New(),
		FamilyID:  familyID,
		Name:      "Test Child",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	err := repo.CreateChild(ctx, child)
	if err != nil {
		t.Fatalf("CreateChild() error = %v", err)
	}
}

func TestInMemoryChildRepository_GetChildByID(t *testing.T) {
	repo := memory.NewInMemoryChildRepository()
	ctx := context.Background()

	child := &models.Child{
		ID:        uuid.New(),
		FamilyID:  uuid.New(),
		Name:      "Test Child",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	_ = repo.CreateChild(ctx, child)

	retrieved, err := repo.GetChildByID(ctx, child.ID)
	if err != nil {
		t.Fatalf("GetChildByID() error = %v", err)
	}

	if retrieved.ID != child.ID {
		t.Errorf("retrieved.ID = %v, want %v", retrieved.ID, child.ID)
	}
}

func TestInMemoryChildRepository_GetChildByID_NotFound(t *testing.T) {
	repo := memory.NewInMemoryChildRepository()
	ctx := context.Background()

	_, err := repo.GetChildByID(ctx, uuid.New())
	if err != models.ErrChildNotFound {
		t.Errorf("GetChildByID() error = %v, want %v", err, models.ErrChildNotFound)
	}
}

func TestInMemoryChildRepository_GetChildrenByFamilyID(t *testing.T) {
	repo := memory.NewInMemoryChildRepository()
	ctx := context.Background()
	familyID := uuid.New()

	for i := 0; i < 3; i++ {
		child := &models.Child{
			ID:        uuid.New(),
			FamilyID:  familyID,
			Name:      "Test Child",
			Status:    "active",
			CreatedAt: time.Now(),
		}
		_ = repo.CreateChild(ctx, child)
	}

	children, err := repo.GetChildrenByFamilyID(ctx, familyID)
	if err != nil {
		t.Fatalf("GetChildrenByFamilyID() error = %v", err)
	}

	if len(children) != 3 {
		t.Errorf("len(children) = %d, want 3", len(children))
	}
}

func TestInMemoryChildRepository_GetChildrenByFamilyID_Empty(t *testing.T) {
	repo := memory.NewInMemoryChildRepository()
	ctx := context.Background()

	children, err := repo.GetChildrenByFamilyID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("GetChildrenByFamilyID() error = %v", err)
	}

	if len(children) != 0 {
		t.Errorf("len(children) = %d, want 0", len(children))
	}
}

func TestInMemoryChildRepository_UpdateChildStatus(t *testing.T) {
	repo := memory.NewInMemoryChildRepository()
	ctx := context.Background()

	child := &models.Child{
		ID:        uuid.New(),
		FamilyID:  uuid.New(),
		Name:      "Test Child",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	_ = repo.CreateChild(ctx, child)

	err := repo.UpdateChildStatus(ctx, child.ID, "inactive")
	if err != nil {
		t.Fatalf("UpdateChildStatus() error = %v", err)
	}

	retrieved, err := repo.GetChildByID(ctx, child.ID)
	if err != nil {
		t.Fatalf("GetChildByID() error = %v", err)
	}

	if retrieved.Status != "inactive" {
		t.Errorf("retrieved.Status = %v, want inactive", retrieved.Status)
	}
}

func TestInMemoryChildRepository_UpdateChildStatus_NotFound(t *testing.T) {
	repo := memory.NewInMemoryChildRepository()
	ctx := context.Background()

	err := repo.UpdateChildStatus(ctx, uuid.New(), "inactive")
	if err != models.ErrChildNotFound {
		t.Errorf("UpdateChildStatus() error = %v, want %v", err, models.ErrChildNotFound)
	}
}

func TestInMemoryFamilyRepository_Isolation(t *testing.T) {
	repo := memory.NewInMemoryFamilyRepository()
	ctx := context.Background()

	original := &models.Family{
		ID:          uuid.New(),
		PairingCode: "ABCD1234",
		CreatedAt:   time.Now(),
	}

	_ = repo.CreateFamily(ctx, original)

	// Mutate the original
	original.PairingCode = "MUTATED!"

	// Verify stored data is unchanged
	retrieved, err := repo.GetFamilyByID(ctx, original.ID)
	if err != nil {
		t.Fatalf("GetFamilyByID() error = %v", err)
	}

	if retrieved.PairingCode != "ABCD1234" {
		t.Errorf("retrieved.PairingCode = %v, want ABCD1234", retrieved.PairingCode)
	}
}

func TestInMemoryChildRepository_Isolation(t *testing.T) {
	repo := memory.NewInMemoryChildRepository()
	ctx := context.Background()

	original := &models.Child{
		ID:        uuid.New(),
		FamilyID:  uuid.New(),
		Name:      "Test Child",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	_ = repo.CreateChild(ctx, original)

	// Mutate the original
	original.Name = "MUTATED!"

	// Verify stored data is unchanged
	retrieved, err := repo.GetChildByID(ctx, original.ID)
	if err != nil {
		t.Fatalf("GetChildByID() error = %v", err)
	}

	if retrieved.Name != "Test Child" {
		t.Errorf("retrieved.Name = %v, want Test Child", retrieved.Name)
	}
}
