package repository

import (
	"fmt"

	"github.com/apardo/mikrom-go/internal/models"
	"gorm.io/gorm"
)

// IPPoolRepository handles database operations for IP pools and allocations
type IPPoolRepository struct {
	db *gorm.DB
}

// NewIPPoolRepository creates a new IPPoolRepository
func NewIPPoolRepository(db *gorm.DB) *IPPoolRepository {
	return &IPPoolRepository{db: db}
}

// CreatePool creates a new IP pool
func (r *IPPoolRepository) CreatePool(pool *models.IPPool) error {
	return r.db.Create(pool).Error
}

// FindPoolByID finds an IP pool by ID
func (r *IPPoolRepository) FindPoolByID(id uint) (*models.IPPool, error) {
	var pool models.IPPool
	err := r.db.First(&pool, id).Error
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

// FindActivePool finds an active IP pool (for simple MVP, just get the first active one)
func (r *IPPoolRepository) FindActivePool() (*models.IPPool, error) {
	var pool models.IPPool
	err := r.db.Where("is_active = ?", true).First(&pool).Error
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

// ListPools lists all IP pools with pagination
func (r *IPPoolRepository) ListPools(offset, limit int) ([]models.IPPool, int64, error) {
	var pools []models.IPPool
	var total int64

	// Get total count
	if err := r.db.Model(&models.IPPool{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&pools).Error

	if err != nil {
		return nil, 0, err
	}

	return pools, total, nil
}

// UpdatePool updates an IP pool
func (r *IPPoolRepository) UpdatePool(pool *models.IPPool) error {
	return r.db.Save(pool).Error
}

// DeletePool deletes an IP pool
func (r *IPPoolRepository) DeletePool(id uint) error {
	return r.db.Delete(&models.IPPool{}, id).Error
}

// AllocateIP allocates an available IP from a pool to a VM
// This uses a transaction to prevent race conditions
func (r *IPPoolRepository) AllocateIP(poolID int, vmID string) (*models.IPAllocation, error) {
	var allocation models.IPAllocation

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Find the pool
		var pool models.IPPool
		if err := tx.First(&pool, poolID).Error; err != nil {
			return fmt.Errorf("pool not found: %w", err)
		}

		if !pool.IsActive {
			return fmt.Errorf("pool is not active")
		}

		// Find an available IP (VMID is empty means not allocated)
		if err := tx.Where("pool_id = ? AND (vm_id = ? OR vm_id IS NULL)", poolID, "").
			First(&allocation).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("no available IPs in pool")
			}
			return fmt.Errorf("failed to find available IP: %w", err)
		}

		// Mark as allocated
		allocation.VMID = vmID
		allocation.IsActive = true

		if err := tx.Save(&allocation).Error; err != nil {
			return fmt.Errorf("failed to allocate IP: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &allocation, nil
}

// ReleaseIP releases an allocated IP back to the pool
func (r *IPPoolRepository) ReleaseIP(ipAddress string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var allocation models.IPAllocation

		if err := tx.Where("ip_address = ?", ipAddress).First(&allocation).Error; err != nil {
			return fmt.Errorf("IP allocation not found: %w", err)
		}

		// Mark as available
		allocation.VMID = ""
		allocation.IsActive = false

		if err := tx.Save(&allocation).Error; err != nil {
			return fmt.Errorf("failed to release IP: %w", err)
		}

		return nil
	})
}

// FindAllocationByVMID finds an IP allocation by VM ID
func (r *IPPoolRepository) FindAllocationByVMID(vmID string) (*models.IPAllocation, error) {
	var allocation models.IPAllocation
	err := r.db.Where("vm_id = ? AND is_active = ?", vmID, true).First(&allocation).Error
	if err != nil {
		return nil, err
	}
	return &allocation, nil
}

// FindAllocationByIP finds an IP allocation by IP address
func (r *IPPoolRepository) FindAllocationByIP(ipAddress string) (*models.IPAllocation, error) {
	var allocation models.IPAllocation
	err := r.db.Where("ip_address = ?", ipAddress).First(&allocation).Error
	if err != nil {
		return nil, err
	}
	return &allocation, nil
}

// CreateAllocations creates multiple IP allocations for a pool
// This is useful when setting up a new pool with a range of IPs
func (r *IPPoolRepository) CreateAllocations(allocations []models.IPAllocation) error {
	return r.db.Create(&allocations).Error
}

// GetPoolStats gets statistics for a pool (total, allocated, available)
func (r *IPPoolRepository) GetPoolStats(poolID int) (total, allocated, available int64, err error) {
	// Total IPs
	if err = r.db.Model(&models.IPAllocation{}).
		Where("pool_id = ?", poolID).
		Count(&total).Error; err != nil {
		return
	}

	// Allocated IPs
	if err = r.db.Model(&models.IPAllocation{}).
		Where("pool_id = ? AND vm_id != ? AND vm_id IS NOT NULL", poolID, "").
		Count(&allocated).Error; err != nil {
		return
	}

	available = total - allocated
	return
}
