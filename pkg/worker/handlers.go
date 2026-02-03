package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/apardo/mikrom-go/internal/models"
	"github.com/apardo/mikrom-go/internal/repository"
	"github.com/apardo/mikrom-go/pkg/firecracker"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// TaskHandler handles background tasks for VM operations
type TaskHandler struct {
	db              *gorm.DB
	vmRepo          *repository.VMRepository
	ipPoolRepo      *repository.IPPoolRepository
	fcClient        *firecracker.Client
	firecrackerPath string
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(
	db *gorm.DB,
	vmRepo *repository.VMRepository,
	ipPoolRepo *repository.IPPoolRepository,
	fcClient *firecracker.Client,
	firecrackerPath string,
) *TaskHandler {
	return &TaskHandler{
		db:              db,
		vmRepo:          vmRepo,
		ipPoolRepo:      ipPoolRepo,
		fcClient:        fcClient,
		firecrackerPath: firecrackerPath,
	}
}

// HandleCreateVM handles the VM creation task
func (h *TaskHandler) HandleCreateVM(ctx context.Context, t *asynq.Task) error {
	var payload CreateVMPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[CreateVM] Starting VM creation for %s", payload.VMID)

	// Update VM status to provisioning
	if err := h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusProvisioning, ""); err != nil {
		log.Printf("[CreateVM] Failed to update status to provisioning: %v", err)
		return err
	}

	// Step 1: Allocate IP from pool
	log.Printf("[CreateVM] Allocating IP for VM %s", payload.VMID)
	activePool, err := h.ipPoolRepo.FindActivePool()
	if err != nil {
		errMsg := fmt.Sprintf("No active IP pool available: %v", err)
		log.Printf("[CreateVM] %s", errMsg)
		h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusError, errMsg)
		return fmt.Errorf(errMsg)
	}

	allocation, err := h.ipPoolRepo.AllocateIP(int(activePool.ID), payload.VMID)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to allocate IP: %v", err)
		log.Printf("[CreateVM] %s", errMsg)
		h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusError, errMsg)
		return fmt.Errorf(errMsg)
	}

	log.Printf("[CreateVM] Allocated IP %s for VM %s", allocation.IPAddress, payload.VMID)

	// Step 2: Update VM with IP address
	vm, err := h.vmRepo.FindByVMID(payload.VMID)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to find VM: %v", err)
		log.Printf("[CreateVM] %s", errMsg)
		// Release IP on failure
		h.ipPoolRepo.ReleaseIP(allocation.IPAddress)
		h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusError, errMsg)
		return fmt.Errorf(errMsg)
	}

	vm.IPAddress = allocation.IPAddress
	if err := h.vmRepo.Update(vm); err != nil {
		errMsg := fmt.Sprintf("Failed to update VM with IP: %v", err)
		log.Printf("[CreateVM] %s", errMsg)
		// Release IP on failure
		h.ipPoolRepo.ReleaseIP(allocation.IPAddress)
		h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusError, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Step 3: Execute Ansible playbook to create VM
	log.Printf("[CreateVM] Executing Ansible playbook for VM %s", payload.VMID)

	fcParams := firecracker.CreateVMParams{
		VMName:     payload.VMID,
		VCPUCount:  payload.VCPUCount,
		MemoryMB:   payload.MemoryMB,
		IPAddress:  allocation.IPAddress,
		KernelPath: payload.KernelPath,
		RootfsPath: payload.RootfsPath,
	}

	result, err := h.fcClient.CreateVM(ctx, fcParams)
	if err != nil || !result.Success {
		errMsg := fmt.Sprintf("Ansible playbook failed: %v (output: %s)", err, result.Error)
		log.Printf("[CreateVM] %s", errMsg)
		// Release IP on failure
		h.ipPoolRepo.ReleaseIP(allocation.IPAddress)
		h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusError, errMsg)
		return fmt.Errorf(errMsg)
	}

	log.Printf("[CreateVM] Ansible playbook succeeded for VM %s", payload.VMID)

	// Step 4: Update VM status to running
	if err := h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusRunning, ""); err != nil {
		log.Printf("[CreateVM] Failed to update status to running: %v", err)
		return err
	}

	log.Printf("[CreateVM] VM %s created successfully", payload.VMID)
	return nil
}

// HandleDeleteVM handles the VM deletion task
func (h *TaskHandler) HandleDeleteVM(ctx context.Context, t *asynq.Task) error {
	var payload DeleteVMPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[DeleteVM] Starting VM deletion for %s", payload.VMID)

	// Get VM to check IP address if not provided
	vm, err := h.vmRepo.FindByVMID(payload.VMID)
	if err != nil {
		log.Printf("[DeleteVM] Failed to find VM: %v", err)
		return err
	}

	ipAddress := payload.IPAddress
	if ipAddress == "" {
		ipAddress = vm.IPAddress
	}

	// Step 1: Execute Ansible cleanup playbook
	if ipAddress != "" {
		log.Printf("[DeleteVM] Executing Ansible cleanup for VM %s", payload.VMID)

		result, err := h.fcClient.CleanupVM(ctx, payload.VMID, ipAddress)
		if err != nil || !result.Success {
			errMsg := fmt.Sprintf("Ansible cleanup failed: %v (output: %s)", err, result.Error)
			log.Printf("[DeleteVM] %s", errMsg)
			h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusError, errMsg)
			return fmt.Errorf(errMsg)
		}

		log.Printf("[DeleteVM] Ansible cleanup succeeded for VM %s", payload.VMID)

		// Step 2: Release IP address
		log.Printf("[DeleteVM] Releasing IP %s for VM %s", ipAddress, payload.VMID)
		if err := h.ipPoolRepo.ReleaseIP(ipAddress); err != nil {
			log.Printf("[DeleteVM] Failed to release IP: %v", err)
			// Continue with deletion even if IP release fails
		}
	}

	// Step 3: Delete VM from database
	log.Printf("[DeleteVM] Deleting VM %s from database", payload.VMID)
	if err := h.vmRepo.Delete(vm); err != nil {
		log.Printf("[DeleteVM] Failed to delete VM: %v", err)
		return err
	}

	log.Printf("[DeleteVM] VM %s deleted successfully", payload.VMID)
	return nil
}

// HandleStartVM handles the VM start task
func (h *TaskHandler) HandleStartVM(ctx context.Context, t *asynq.Task) error {
	var payload StartVMPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[StartVM] Starting VM %s", payload.VMID)

	// Execute Ansible start playbook
	result, err := h.fcClient.StartVM(ctx, payload.VMID, payload.IPAddress)
	if err != nil || !result.Success {
		errMsg := fmt.Sprintf("Ansible start failed: %v (output: %s)", err, result.Error)
		log.Printf("[StartVM] %s", errMsg)
		h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusError, errMsg)
		return fmt.Errorf(errMsg)
	}

	log.Printf("[StartVM] Ansible start succeeded for VM %s", payload.VMID)

	// Update VM status to running
	if err := h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusRunning, ""); err != nil {
		log.Printf("[StartVM] Failed to update status to running: %v", err)
		return err
	}

	log.Printf("[StartVM] VM %s started successfully", payload.VMID)
	return nil
}

// HandleStopVM handles the VM stop task
func (h *TaskHandler) HandleStopVM(ctx context.Context, t *asynq.Task) error {
	var payload StopVMPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[StopVM] Stopping VM %s", payload.VMID)

	// Execute Ansible stop playbook
	result, err := h.fcClient.StopVM(ctx, payload.VMID, payload.IPAddress)
	if err != nil || !result.Success {
		errMsg := fmt.Sprintf("Ansible stop failed: %v (output: %s)", err, result.Error)
		log.Printf("[StopVM] %s", errMsg)
		h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusError, errMsg)
		return fmt.Errorf(errMsg)
	}

	log.Printf("[StopVM] Ansible stop succeeded for VM %s", payload.VMID)

	// Update VM status to stopped
	if err := h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusStopped, ""); err != nil {
		log.Printf("[StopVM] Failed to update status to stopped: %v", err)
		return err
	}

	log.Printf("[StopVM] VM %s stopped successfully", payload.VMID)
	return nil
}

// HandleRestartVM handles the VM restart task
func (h *TaskHandler) HandleRestartVM(ctx context.Context, t *asynq.Task) error {
	var payload RestartVMPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[RestartVM] Restarting VM %s", payload.VMID)

	// Step 1: Stop the VM
	log.Printf("[RestartVM] Stopping VM %s", payload.VMID)
	h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusStopping, "")

	result, err := h.fcClient.StopVM(ctx, payload.VMID, payload.IPAddress)
	if err != nil || !result.Success {
		errMsg := fmt.Sprintf("Ansible stop failed during restart: %v (output: %s)", err, result.Error)
		log.Printf("[RestartVM] %s", errMsg)
		h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusError, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Step 2: Start the VM
	log.Printf("[RestartVM] Starting VM %s", payload.VMID)
	h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusStarting, "")

	result, err = h.fcClient.StartVM(ctx, payload.VMID, payload.IPAddress)
	if err != nil || !result.Success {
		errMsg := fmt.Sprintf("Ansible start failed during restart: %v (output: %s)", err, result.Error)
		log.Printf("[RestartVM] %s", errMsg)
		h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusError, errMsg)
		return fmt.Errorf(errMsg)
	}

	log.Printf("[RestartVM] Ansible restart succeeded for VM %s", payload.VMID)

	// Update VM status to running
	if err := h.vmRepo.UpdateStatus(payload.VMID, models.VMStatusRunning, ""); err != nil {
		log.Printf("[RestartVM] Failed to update status to running: %v", err)
		return err
	}

	log.Printf("[RestartVM] VM %s restarted successfully", payload.VMID)
	return nil
}
