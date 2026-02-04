package service

import (
	"fmt"
	"log"
	"math"

	"github.com/spluca/mikrom/internal/models"
	"github.com/spluca/mikrom/internal/repository"
	"github.com/spluca/mikrom/pkg/utils"
	"github.com/spluca/mikrom/pkg/worker"
)

// WorkerClient defines the interface for enqueueing VM tasks
type WorkerClient interface {
	EnqueueCreateVM(payload *worker.CreateVMPayload) error
	EnqueueDeleteVM(payload *worker.DeleteVMPayload) error
	EnqueueStartVM(payload *worker.StartVMPayload) error
	EnqueueStopVM(payload *worker.StopVMPayload) error
	EnqueueRestartVM(payload *worker.RestartVMPayload) error
	Close() error
}

type VMService struct {
	vmRepo       *repository.VMRepository
	workerClient WorkerClient
}

func NewVMService(vmRepo *repository.VMRepository, workerClient WorkerClient) *VMService {
	return &VMService{
		vmRepo:       vmRepo,
		workerClient: workerClient,
	}
}

// CreateVM creates a new VM and queues it for provisioning
func (s *VMService) CreateVM(req models.CreateVMRequest, userID int) (*models.VM, error) {
	// Generate unique VM ID
	vmID := utils.GenerateVMID()

	// Create VM record with pending status
	vm := &models.VM{
		VMID:        vmID,
		Name:        req.Name,
		Description: req.Description,
		VCPUCount:   req.VCPUCount,
		MemoryMB:    req.MemoryMB,
		UserID:      userID,
		Status:      models.VMStatusPending,
		KernelPath:  req.KernelPath,
		RootfsPath:  req.RootfsPath,
	}

	if err := s.vmRepo.Create(vm); err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	// Queue background task for VM creation
	payload := &worker.CreateVMPayload{
		VMID:        vm.VMID,
		UserID:      uint(vm.UserID),
		Name:        vm.Name,
		VCPUCount:   vm.VCPUCount,
		MemoryMB:    vm.MemoryMB,
		KernelPath:  vm.KernelPath,
		RootfsPath:  vm.RootfsPath,
		Description: vm.Description,
	}

	if err := s.workerClient.EnqueueCreateVM(payload); err != nil {
		log.Printf("Failed to enqueue CreateVM task: %v", err)
		// Don't fail the request, but log the error
		// VM will remain in "pending" status
	}

	return vm, nil
}

// GetVM retrieves a VM by its vm_id
func (s *VMService) GetVM(vmID string, userID int) (*models.VM, error) {
	vm, err := s.vmRepo.FindByVMID(vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM: %w", err)
	}

	if vm == nil {
		return nil, nil
	}

	// Check ownership
	if vm.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to VM")
	}

	return vm, nil
}

// ListVMs retrieves all VMs for a user with pagination
func (s *VMService) ListVMs(userID int, page, pageSize int) (*models.VMListResponse, error) {
	offset := (page - 1) * pageSize

	vms, total, err := s.vmRepo.FindByUserID(userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	// Convert to response format
	vmResponses := make([]models.VMResponse, len(vms))
	for i, vm := range vms {
		vmResponses[i] = vmToResponse(&vm)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &models.VMListResponse{
		Items:      vmResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateVM updates VM metadata (name, description)
func (s *VMService) UpdateVM(vmID string, userID int, req models.UpdateVMRequest) (*models.VM, error) {
	vm, err := s.vmRepo.FindByVMID(vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM: %w", err)
	}

	if vm == nil {
		return nil, nil
	}

	// Check ownership
	if vm.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to VM")
	}

	// Update fields
	if req.Name != nil {
		vm.Name = *req.Name
	}
	if req.Description != nil {
		vm.Description = *req.Description
	}

	if err := s.vmRepo.Update(vm); err != nil {
		return nil, fmt.Errorf("failed to update VM: %w", err)
	}

	return vm, nil
}

// DeleteVM marks a VM for deletion and queues the deletion task
func (s *VMService) DeleteVM(vmID string, userID int) error {
	vm, err := s.vmRepo.FindByVMID(vmID)
	if err != nil {
		return fmt.Errorf("failed to find VM: %w", err)
	}

	if vm == nil {
		return fmt.Errorf("VM not found")
	}

	// Check ownership
	if vm.UserID != userID {
		return fmt.Errorf("unauthorized access to VM")
	}

	// Check if already deleting
	if vm.Status == models.VMStatusDeleting {
		return fmt.Errorf("VM is already being deleted")
	}

	// Update status to deleting
	if err := s.vmRepo.UpdateStatus(vmID, models.VMStatusDeleting, ""); err != nil {
		return fmt.Errorf("failed to update VM status: %w", err)
	}

	// Queue background task for VM deletion
	payload := &worker.DeleteVMPayload{
		VMID:      vm.VMID,
		UserID:    uint(vm.UserID),
		IPAddress: vm.IPAddress,
	}

	if err := s.workerClient.EnqueueDeleteVM(payload); err != nil {
		log.Printf("Failed to enqueue DeleteVM task: %v", err)
		// Rollback status update
		s.vmRepo.UpdateStatus(vmID, vm.Status, "")
		return fmt.Errorf("failed to enqueue deletion task: %w", err)
	}

	return nil
}

// StartVM queues a task to start a stopped VM
func (s *VMService) StartVM(vmID string, userID int) error {
	vm, err := s.vmRepo.FindByVMID(vmID)
	if err != nil {
		return fmt.Errorf("failed to find VM: %w", err)
	}

	if vm == nil {
		return fmt.Errorf("VM not found")
	}

	// Check ownership
	if vm.UserID != userID {
		return fmt.Errorf("unauthorized access to VM")
	}

	// Check if VM can be started
	if vm.Status != models.VMStatusStopped && vm.Status != models.VMStatusError {
		return fmt.Errorf("VM cannot be started from current status: %s", vm.Status)
	}

	// Update status to starting
	if err := s.vmRepo.UpdateStatus(vmID, models.VMStatusStarting, ""); err != nil {
		return fmt.Errorf("failed to update VM status: %w", err)
	}

	// Queue background task for VM start
	payload := &worker.StartVMPayload{
		VMID:      vm.VMID,
		UserID:    uint(vm.UserID),
		IPAddress: vm.IPAddress,
	}

	if err := s.workerClient.EnqueueStartVM(payload); err != nil {
		log.Printf("Failed to enqueue StartVM task: %v", err)
		// Rollback status update
		s.vmRepo.UpdateStatus(vmID, vm.Status, "")
		return fmt.Errorf("failed to enqueue start task: %w", err)
	}

	return nil
}

// StopVM queues a task to stop a running VM
func (s *VMService) StopVM(vmID string, userID int) error {
	vm, err := s.vmRepo.FindByVMID(vmID)
	if err != nil {
		return fmt.Errorf("failed to find VM: %w", err)
	}

	if vm == nil {
		return fmt.Errorf("VM not found")
	}

	// Check ownership
	if vm.UserID != userID {
		return fmt.Errorf("unauthorized access to VM")
	}

	// Check if VM is running
	if vm.Status != models.VMStatusRunning {
		return fmt.Errorf("VM is not running (current status: %s)", vm.Status)
	}

	// Update status to stopping
	if err := s.vmRepo.UpdateStatus(vmID, models.VMStatusStopping, ""); err != nil {
		return fmt.Errorf("failed to update VM status: %w", err)
	}

	// Queue background task for VM stop
	payload := &worker.StopVMPayload{
		VMID:      vm.VMID,
		UserID:    uint(vm.UserID),
		IPAddress: vm.IPAddress,
	}

	if err := s.workerClient.EnqueueStopVM(payload); err != nil {
		log.Printf("Failed to enqueue StopVM task: %v", err)
		// Rollback status update
		s.vmRepo.UpdateStatus(vmID, vm.Status, "")
		return fmt.Errorf("failed to enqueue stop task: %w", err)
	}

	return nil
}

// RestartVM queues a task to restart a running VM
func (s *VMService) RestartVM(vmID string, userID int) error {
	vm, err := s.vmRepo.FindByVMID(vmID)
	if err != nil {
		return fmt.Errorf("failed to find VM: %w", err)
	}

	if vm == nil {
		return fmt.Errorf("VM not found")
	}

	// Check ownership
	if vm.UserID != userID {
		return fmt.Errorf("unauthorized access to VM")
	}

	// Check if VM is running
	if vm.Status != models.VMStatusRunning {
		return fmt.Errorf("VM is not running (current status: %s)", vm.Status)
	}

	// Update status to restarting
	if err := s.vmRepo.UpdateStatus(vmID, models.VMStatusRestarting, ""); err != nil {
		return fmt.Errorf("failed to update VM status: %w", err)
	}

	// Queue background task for VM restart
	payload := &worker.RestartVMPayload{
		VMID:      vm.VMID,
		UserID:    uint(vm.UserID),
		IPAddress: vm.IPAddress,
	}

	if err := s.workerClient.EnqueueRestartVM(payload); err != nil {
		log.Printf("Failed to enqueue RestartVM task: %v", err)
		// Rollback status update
		s.vmRepo.UpdateStatus(vmID, vm.Status, "")
		return fmt.Errorf("failed to enqueue restart task: %w", err)
	}

	return nil
}

// Helper function to convert VM to VMResponse
func vmToResponse(vm *models.VM) models.VMResponse {
	return models.VMResponse{
		ID:           vm.ID,
		VMID:         vm.VMID,
		Name:         vm.Name,
		Description:  vm.Description,
		VCPUCount:    vm.VCPUCount,
		MemoryMB:     vm.MemoryMB,
		IPAddress:    vm.IPAddress,
		Status:       vm.Status,
		ErrorMessage: vm.ErrorMessage,
		Host:         vm.Host,
		UserID:       vm.UserID,
		CreatedAt:    vm.CreatedAt,
		UpdatedAt:    vm.UpdatedAt,
	}
}
