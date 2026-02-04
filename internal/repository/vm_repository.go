package repository

import (
	"errors"
	"fmt"

	"github.com/spluca/mikrom/internal/models"
	"gorm.io/gorm"
)

type VMRepository struct {
	db *gorm.DB
}

func NewVMRepository(db *gorm.DB) *VMRepository {
	return &VMRepository{db: db}
}

// Create creates a new VM
func (r *VMRepository) Create(vm *models.VM) error {
	if err := r.db.Create(vm).Error; err != nil {
		return fmt.Errorf("error creating VM: %w", err)
	}
	return nil
}

// FindByID finds a VM by its database ID
func (r *VMRepository) FindByID(id int) (*models.VM, error) {
	var vm models.VM
	err := r.db.Preload("User").First(&vm, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding VM: %w", err)
	}

	return &vm, nil
}

// FindByVMID finds a VM by its vm_id (srv-xxxxxxxx)
func (r *VMRepository) FindByVMID(vmID string) (*models.VM, error) {
	var vm models.VM
	err := r.db.Preload("User").Where("vm_id = ?", vmID).First(&vm).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding VM: %w", err)
	}

	return &vm, nil
}

// FindByUserID finds all VMs owned by a user with pagination
func (r *VMRepository) FindByUserID(userID int, offset, limit int) ([]models.VM, int64, error) {
	var vms []models.VM
	var total int64

	// Count total
	if err := r.db.Model(&models.VM{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("error counting VMs: %w", err)
	}

	// Get VMs with pagination
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&vms).Error

	if err != nil {
		return nil, 0, fmt.Errorf("error finding VMs: %w", err)
	}

	return vms, total, nil
}

// FindAll returns all VMs with pagination (for superuser)
func (r *VMRepository) FindAll(offset, limit int) ([]models.VM, int64, error) {
	var vms []models.VM
	var total int64

	// Count total
	if err := r.db.Model(&models.VM{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("error counting VMs: %w", err)
	}

	// Get VMs with pagination
	err := r.db.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&vms).Error

	if err != nil {
		return nil, 0, fmt.Errorf("error finding VMs: %w", err)
	}

	return vms, total, nil
}

// Update updates a VM
func (r *VMRepository) Update(vm *models.VM) error {
	if err := r.db.Save(vm).Error; err != nil {
		return fmt.Errorf("error updating VM: %w", err)
	}
	return nil
}

// Delete deletes a VM
func (r *VMRepository) Delete(vm *models.VM) error {
	if err := r.db.Delete(vm).Error; err != nil {
		return fmt.Errorf("error deleting VM: %w", err)
	}
	return nil
}

// UpdateStatus updates the status of a VM
func (r *VMRepository) UpdateStatus(vmID string, status models.VMStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	} else {
		updates["error_message"] = ""
	}

	err := r.db.Model(&models.VM{}).Where("vm_id = ?", vmID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("error updating VM status: %w", err)
	}

	return nil
}
