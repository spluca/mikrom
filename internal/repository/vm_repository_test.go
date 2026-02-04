package repository

import (
	"fmt"
	"testing"

	"github.com/spluca/mikrom/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVMRepository_Create_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create a user first (FK requirement)
	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hash",
		Name:         "Test User",
	}
	db.Create(user)

	repo := NewVMRepository(db)

	vm := &models.VM{
		VMID:        "srv-test123",
		Name:        "Test VM",
		Description: "Test Description",
		VCPUCount:   2,
		MemoryMB:    1024,
		Status:      models.VMStatusPending,
		UserID:      int(user.ID),
	}

	err := repo.Create(vm)

	assert.NoError(t, err)
	assert.NotZero(t, vm.ID)
	assert.NotZero(t, vm.CreatedAt)
	assert.NotZero(t, vm.UpdatedAt)

	// Verify VM was created
	var found models.VM
	db.First(&found, vm.ID)
	assert.Equal(t, vm.VMID, found.VMID)
	assert.Equal(t, vm.Name, found.Name)
	assert.Equal(t, vm.VCPUCount, found.VCPUCount)
	assert.Equal(t, vm.MemoryMB, found.MemoryMB)
}

func TestVMRepository_Create_Error(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewVMRepository(db)

	// Close DB to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 1,
		MemoryMB:  512,
	}

	err := repo.Create(vm)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating VM")
}

func TestVMRepository_FindByID_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create user and VM
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User"}
	db.Create(user)

	expected := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		VCPUCount: 2,
		MemoryMB:  1024,
		UserID:    int(user.ID),
		Status:    models.VMStatusRunning,
	}
	db.Create(expected)

	repo := NewVMRepository(db)
	vm, err := repo.FindByID(int(expected.ID))

	assert.NoError(t, err)
	require.NotNil(t, vm)
	assert.Equal(t, expected.VMID, vm.VMID)
	assert.Equal(t, expected.Name, vm.Name)
	assert.NotNil(t, vm.User) // Preloaded
	assert.Equal(t, user.Email, vm.User.Email)
}

func TestVMRepository_FindByID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewVMRepository(db)
	vm, err := repo.FindByID(999)

	assert.NoError(t, err)
	assert.Nil(t, vm)
}

func TestVMRepository_FindByID_Error(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewVMRepository(db)

	// Close DB to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	vm, err := repo.FindByID(1)

	assert.Error(t, err)
	assert.Nil(t, vm)
	assert.Contains(t, err.Error(), "error finding VM")
}

func TestVMRepository_FindByVMID_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create user and VM
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User"}
	db.Create(user)

	expected := &models.VM{
		VMID:      "srv-unique123",
		Name:      "Test VM",
		VCPUCount: 2,
		MemoryMB:  1024,
		UserID:    int(user.ID),
	}
	db.Create(expected)

	repo := NewVMRepository(db)
	vm, err := repo.FindByVMID("srv-unique123")

	assert.NoError(t, err)
	require.NotNil(t, vm)
	assert.Equal(t, expected.VMID, vm.VMID)
	assert.Equal(t, expected.Name, vm.Name)
	assert.NotNil(t, vm.User) // Preloaded
}

func TestVMRepository_FindByVMID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewVMRepository(db)
	vm, err := repo.FindByVMID("srv-nonexistent")

	assert.NoError(t, err)
	assert.Nil(t, vm)
}

func TestVMRepository_FindByUserID_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User"}
	db.Create(user)

	// Create multiple VMs for this user
	for i := 1; i <= 5; i++ {
		vm := &models.VM{
			VMID:      fmt.Sprintf("srv-test%d", i),
			Name:      fmt.Sprintf("VM %d", i),
			VCPUCount: 1,
			MemoryMB:  512,
			UserID:    int(user.ID),
		}
		db.Create(vm)
	}

	repo := NewVMRepository(db)
	vms, total, err := repo.FindByUserID(int(user.ID), 0, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, vms, 5)
}

func TestVMRepository_FindByUserID_Pagination(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User"}
	db.Create(user)

	// Create 10 VMs
	for i := 1; i <= 10; i++ {
		vm := &models.VM{
			VMID:      fmt.Sprintf("srv-test%d", i),
			Name:      fmt.Sprintf("VM %d", i),
			VCPUCount: 1,
			MemoryMB:  512,
			UserID:    int(user.ID),
		}
		db.Create(vm)
	}

	repo := NewVMRepository(db)

	// Test first page
	vms, total, err := repo.FindByUserID(int(user.ID), 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, vms, 5)

	// Test second page
	vms, total, err = repo.FindByUserID(int(user.ID), 5, 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, vms, 5)
}

func TestVMRepository_FindByUserID_NoVMs(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewVMRepository(db)
	vms, total, err := repo.FindByUserID(999, 0, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, vms)
}

func TestVMRepository_FindAll_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create users
	user1 := &models.User{Email: "user1@example.com", PasswordHash: "hash", Name: "User 1"}
	user2 := &models.User{Email: "user2@example.com", PasswordHash: "hash", Name: "User 2"}
	db.Create(user1)
	db.Create(user2)

	// Create VMs for different users
	for i := 1; i <= 3; i++ {
		db.Create(&models.VM{
			VMID:      fmt.Sprintf("srv-user1-%d", i),
			Name:      fmt.Sprintf("VM %d", i),
			VCPUCount: 1,
			MemoryMB:  512,
			UserID:    int(user1.ID),
		})
	}

	for i := 1; i <= 2; i++ {
		db.Create(&models.VM{
			VMID:      fmt.Sprintf("srv-user2-%d", i),
			Name:      fmt.Sprintf("VM %d", i),
			VCPUCount: 1,
			MemoryMB:  512,
			UserID:    int(user2.ID),
		})
	}

	repo := NewVMRepository(db)
	vms, total, err := repo.FindAll(0, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, vms, 5)

	// Verify User preloaded
	for _, vm := range vms {
		assert.NotZero(t, vm.User.ID)
	}
}

func TestVMRepository_FindAll_Pagination(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create user
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User"}
	db.Create(user)

	// Create 8 VMs
	for i := 1; i <= 8; i++ {
		db.Create(&models.VM{
			VMID:      fmt.Sprintf("srv-test%d", i),
			Name:      fmt.Sprintf("VM %d", i),
			VCPUCount: 1,
			MemoryMB:  512,
			UserID:    int(user.ID),
		})
	}

	repo := NewVMRepository(db)

	// Test pagination
	vms, total, err := repo.FindAll(0, 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), total)
	assert.Len(t, vms, 3)

	vms, total, err = repo.FindAll(3, 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), total)
	assert.Len(t, vms, 3)

	vms, total, err = repo.FindAll(6, 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(8), total)
	assert.Len(t, vms, 2) // Last page has only 2
}

func TestVMRepository_Update_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create user and VM
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User"}
	db.Create(user)

	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Original Name",
		VCPUCount: 1,
		MemoryMB:  512,
		UserID:    int(user.ID),
	}
	db.Create(vm)

	repo := NewVMRepository(db)

	// Update VM
	vm.Name = "Updated Name"
	vm.Description = "New Description"
	vm.IPAddress = "192.168.1.10"

	err := repo.Update(vm)

	assert.NoError(t, err)

	// Verify update
	var found models.VM
	db.First(&found, vm.ID)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "New Description", found.Description)
	assert.Equal(t, "192.168.1.10", found.IPAddress)
}

func TestVMRepository_Update_Error(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewVMRepository(db)

	// Close DB to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	vm := &models.VM{ID: 1, Name: "Test"}
	err := repo.Update(vm)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error updating VM")
}

func TestVMRepository_Delete_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create user and VM
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User"}
	db.Create(user)

	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		VCPUCount: 1,
		MemoryMB:  512,
		UserID:    int(user.ID),
	}
	db.Create(vm)

	repo := NewVMRepository(db)
	err := repo.Delete(vm)

	assert.NoError(t, err)

	// Verify deletion
	var found models.VM
	result := db.First(&found, vm.ID)
	assert.Error(t, result.Error) // Should not find
}

func TestVMRepository_Delete_Error(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewVMRepository(db)

	// Close DB to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	vm := &models.VM{ID: 1}
	err := repo.Delete(vm)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error deleting VM")
}

func TestVMRepository_UpdateStatus_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create user and VM
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User"}
	db.Create(user)

	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		VCPUCount: 1,
		MemoryMB:  512,
		UserID:    int(user.ID),
		Status:    models.VMStatusPending,
	}
	db.Create(vm)

	repo := NewVMRepository(db)

	// Update status to running
	err := repo.UpdateStatus("srv-test123", models.VMStatusRunning, "")

	assert.NoError(t, err)

	// Verify update
	var found models.VM
	db.Where("vm_id = ?", "srv-test123").First(&found)
	assert.Equal(t, models.VMStatusRunning, found.Status)
	assert.Empty(t, found.ErrorMessage)
}

func TestVMRepository_UpdateStatus_WithError(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create user and VM
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User"}
	db.Create(user)

	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		VCPUCount: 1,
		MemoryMB:  512,
		UserID:    int(user.ID),
		Status:    models.VMStatusProvisioning,
	}
	db.Create(vm)

	repo := NewVMRepository(db)

	// Update status to error with message
	errorMsg := "Failed to provision VM"
	err := repo.UpdateStatus("srv-test123", models.VMStatusError, errorMsg)

	assert.NoError(t, err)

	// Verify update
	var found models.VM
	db.Where("vm_id = ?", "srv-test123").First(&found)
	assert.Equal(t, models.VMStatusError, found.Status)
	assert.Equal(t, errorMsg, found.ErrorMessage)
}

func TestVMRepository_UpdateStatus_ClearError(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create user and VM with error
	user := &models.User{Email: "test@example.com", PasswordHash: "hash", Name: "Test User"}
	db.Create(user)

	vm := &models.VM{
		VMID:         "srv-test123",
		Name:         "Test VM",
		VCPUCount:    1,
		MemoryMB:     512,
		UserID:       int(user.ID),
		Status:       models.VMStatusError,
		ErrorMessage: "Previous error",
	}
	db.Create(vm)

	repo := NewVMRepository(db)

	// Update status to running (should clear error)
	err := repo.UpdateStatus("srv-test123", models.VMStatusRunning, "")

	assert.NoError(t, err)

	// Verify error was cleared
	var found models.VM
	db.Where("vm_id = ?", "srv-test123").First(&found)
	assert.Equal(t, models.VMStatusRunning, found.Status)
	assert.Empty(t, found.ErrorMessage)
}

func TestVMRepository_UpdateStatus_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewVMRepository(db)

	// Try to update non-existent VM
	err := repo.UpdateStatus("srv-nonexistent", models.VMStatusRunning, "")

	// GORM doesn't return error if no rows affected
	assert.NoError(t, err)
}
