package service

import (
	"testing"

	"github.com/spluca/mikrom/internal/models"
	"github.com/spluca/mikrom/internal/repository"
	"github.com/spluca/mikrom/pkg/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWorkerClient simulates the worker client for testing
type MockWorkerClient struct {
	EnqueueCreateVMFunc  func(payload *worker.CreateVMPayload) error
	EnqueueDeleteVMFunc  func(payload *worker.DeleteVMPayload) error
	EnqueueStartVMFunc   func(payload *worker.StartVMPayload) error
	EnqueueStopVMFunc    func(payload *worker.StopVMPayload) error
	EnqueueRestartVMFunc func(payload *worker.RestartVMPayload) error

	CreateVMCalls  int
	DeleteVMCalls  int
	StartVMCalls   int
	StopVMCalls    int
	RestartVMCalls int
}

func (m *MockWorkerClient) EnqueueCreateVM(payload *worker.CreateVMPayload) error {
	m.CreateVMCalls++
	if m.EnqueueCreateVMFunc != nil {
		return m.EnqueueCreateVMFunc(payload)
	}
	return nil
}

func (m *MockWorkerClient) EnqueueDeleteVM(payload *worker.DeleteVMPayload) error {
	m.DeleteVMCalls++
	if m.EnqueueDeleteVMFunc != nil {
		return m.EnqueueDeleteVMFunc(payload)
	}
	return nil
}

func (m *MockWorkerClient) EnqueueStartVM(payload *worker.StartVMPayload) error {
	m.StartVMCalls++
	if m.EnqueueStartVMFunc != nil {
		return m.EnqueueStartVMFunc(payload)
	}
	return nil
}

func (m *MockWorkerClient) EnqueueStopVM(payload *worker.StopVMPayload) error {
	m.StopVMCalls++
	if m.EnqueueStopVMFunc != nil {
		return m.EnqueueStopVMFunc(payload)
	}
	return nil
}

func (m *MockWorkerClient) EnqueueRestartVM(payload *worker.RestartVMPayload) error {
	m.RestartVMCalls++
	if m.EnqueueRestartVMFunc != nil {
		return m.EnqueueRestartVMFunc(payload)
	}
	return nil
}

func (m *MockWorkerClient) Close() error {
	return nil
}

func TestCreateVM_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	req := models.CreateVMRequest{
		Name:        "Test VM",
		Description: "Test Description",
		VCPUCount:   2,
		MemoryMB:    1024,
		KernelPath:  "/path/to/kernel",
		RootfsPath:  "/path/to/rootfs",
	}

	vm, err := vmService.CreateVM(req, 1)

	assert.NoError(t, err)
	require.NotNil(t, vm)
	assert.NotZero(t, vm.ID)
	assert.NotEmpty(t, vm.VMID)
	assert.Equal(t, req.Name, vm.Name)
	assert.Equal(t, req.Description, vm.Description)
	assert.Equal(t, req.VCPUCount, vm.VCPUCount)
	assert.Equal(t, req.MemoryMB, vm.MemoryMB)
	assert.Equal(t, 1, vm.UserID)
	assert.Equal(t, models.VMStatusPending, vm.Status)

	// Verify worker was called
	assert.Equal(t, 1, mockWorker.CreateVMCalls)
}

func TestCreateVM_WorkerError(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{
		EnqueueCreateVMFunc: func(payload *worker.CreateVMPayload) error {
			return assert.AnError
		},
	}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	req := models.CreateVMRequest{
		Name:      "Test VM",
		VCPUCount: 2,
		MemoryMB:  1024,
	}

	// Should still succeed even if worker fails
	vm, err := vmService.CreateVM(req, 1)

	assert.NoError(t, err)
	require.NotNil(t, vm)
	assert.Equal(t, models.VMStatusPending, vm.Status)
}

func TestGetVM_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create a VM
	expectedVM := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
		Status:    models.VMStatusRunning,
	}
	db.Create(expectedVM)

	vm, err := vmService.GetVM("srv-test123", 1)

	assert.NoError(t, err)
	require.NotNil(t, vm)
	assert.Equal(t, expectedVM.VMID, vm.VMID)
	assert.Equal(t, expectedVM.Name, vm.Name)
}

func TestGetVM_NotFound(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	vm, err := vmService.GetVM("srv-nonexistent", 1)

	assert.NoError(t, err)
	assert.Nil(t, vm)
}

func TestGetVM_UnauthorizedAccess(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create VM owned by user 1
	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
	}
	db.Create(vm)

	// Try to access with different user
	result, err := vmService.GetVM("srv-test123", 2)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestListVMs_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create multiple VMs
	for i := 1; i <= 5; i++ {
		vm := &models.VM{
			VMID:      "srv-test" + string(rune(i)),
			Name:      "VM " + string(rune(i)),
			UserID:    1,
			VCPUCount: 1,
			MemoryMB:  512,
		}
		db.Create(vm)
	}

	response, err := vmService.ListVMs(1, 1, 10)

	assert.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, int64(5), response.Total)
	assert.Len(t, response.Items, 5)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
	assert.Equal(t, 1, response.TotalPages)
}

func TestListVMs_Pagination(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create 10 VMs
	for i := 1; i <= 10; i++ {
		vm := &models.VM{
			VMID:      "srv-test" + string(rune(i)),
			Name:      "VM " + string(rune(i)),
			UserID:    1,
			VCPUCount: 1,
			MemoryMB:  512,
		}
		db.Create(vm)
	}

	// Get first page
	response, err := vmService.ListVMs(1, 1, 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), response.Total)
	assert.Len(t, response.Items, 3)
	assert.Equal(t, 4, response.TotalPages)

	// Get second page
	response, err = vmService.ListVMs(1, 2, 3)
	assert.NoError(t, err)
	assert.Len(t, response.Items, 3)
}

func TestUpdateVM_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create VM
	vm := &models.VM{
		VMID:        "srv-test123",
		Name:        "Original Name",
		Description: "Original Description",
		UserID:      1,
		VCPUCount:   2,
		MemoryMB:    1024,
	}
	db.Create(vm)

	// Update VM
	newName := "Updated Name"
	newDesc := "Updated Description"
	req := models.UpdateVMRequest{
		Name:        &newName,
		Description: &newDesc,
	}

	updated, err := vmService.UpdateVM("srv-test123", 1, req)

	assert.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, newName, updated.Name)
	assert.Equal(t, newDesc, updated.Description)
}

func TestUpdateVM_NotFound(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	newName := "Updated Name"
	req := models.UpdateVMRequest{Name: &newName}

	updated, err := vmService.UpdateVM("srv-nonexistent", 1, req)

	assert.NoError(t, err)
	assert.Nil(t, updated)
}

func TestUpdateVM_UnauthorizedAccess(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create VM owned by user 1
	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
	}
	db.Create(vm)

	// Try to update with different user
	newName := "Updated Name"
	req := models.UpdateVMRequest{Name: &newName}
	updated, err := vmService.UpdateVM("srv-test123", 2, req)

	assert.Error(t, err)
	assert.Nil(t, updated)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestDeleteVM_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create VM
	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
		Status:    models.VMStatusRunning,
	}
	db.Create(vm)

	err := vmService.DeleteVM("srv-test123", 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, mockWorker.DeleteVMCalls)

	// Verify status updated
	var updated models.VM
	db.Where("vm_id = ?", "srv-test123").First(&updated)
	assert.Equal(t, models.VMStatusDeleting, updated.Status)
}

func TestDeleteVM_NotFound(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	err := vmService.DeleteVM("srv-nonexistent", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteVM_AlreadyDeleting(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create VM already in deleting status
	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
		Status:    models.VMStatusDeleting,
	}
	db.Create(vm)

	err := vmService.DeleteVM("srv-test123", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already being deleted")
}

func TestStartVM_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create stopped VM
	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
		Status:    models.VMStatusStopped,
	}
	db.Create(vm)

	err := vmService.StartVM("srv-test123", 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, mockWorker.StartVMCalls)

	// Verify status updated
	var updated models.VM
	db.Where("vm_id = ?", "srv-test123").First(&updated)
	assert.Equal(t, models.VMStatusStarting, updated.Status)
}

func TestStartVM_InvalidStatus(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create running VM
	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
		Status:    models.VMStatusRunning,
	}
	db.Create(vm)

	err := vmService.StartVM("srv-test123", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be started")
}

func TestStopVM_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create running VM
	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
		Status:    models.VMStatusRunning,
	}
	db.Create(vm)

	err := vmService.StopVM("srv-test123", 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, mockWorker.StopVMCalls)

	// Verify status updated
	var updated models.VM
	db.Where("vm_id = ?", "srv-test123").First(&updated)
	assert.Equal(t, models.VMStatusStopping, updated.Status)
}

func TestStopVM_InvalidStatus(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create stopped VM
	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
		Status:    models.VMStatusStopped,
	}
	db.Create(vm)

	err := vmService.StopVM("srv-test123", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestRestartVM_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create running VM
	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
		Status:    models.VMStatusRunning,
	}
	db.Create(vm)

	err := vmService.RestartVM("srv-test123", 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, mockWorker.RestartVMCalls)

	// Verify status updated
	var updated models.VM
	db.Where("vm_id = ?", "srv-test123").First(&updated)
	assert.Equal(t, models.VMStatusRestarting, updated.Status)
}

func TestRestartVM_InvalidStatus(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	vmRepo := repository.NewVMRepository(db)
	mockWorker := &MockWorkerClient{}
	vmService := &VMService{
		vmRepo:       vmRepo,
		workerClient: mockWorker,
	}

	// Create stopped VM
	vm := &models.VM{
		VMID:      "srv-test123",
		Name:      "Test VM",
		UserID:    1,
		VCPUCount: 2,
		MemoryMB:  1024,
		Status:    models.VMStatusStopped,
	}
	db.Create(vm)

	err := vmService.RestartVM("srv-test123", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}
