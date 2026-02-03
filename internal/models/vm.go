package models

import (
	"time"
)

// VMStatus represents the current state of a VM
type VMStatus string

const (
	VMStatusPending      VMStatus = "pending"
	VMStatusProvisioning VMStatus = "provisioning"
	VMStatusStarting     VMStatus = "starting"
	VMStatusRunning      VMStatus = "running"
	VMStatusStopping     VMStatus = "stopping"
	VMStatusStopped      VMStatus = "stopped"
	VMStatusRestarting   VMStatus = "restarting"
	VMStatusError        VMStatus = "error"
	VMStatusDeleting     VMStatus = "deleting"
)

// VM represents a Firecracker microVM
type VM struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	VMID        string `json:"vm_id" gorm:"uniqueIndex;not null;size:50"`
	Name        string `json:"name" gorm:"not null;size:64"`
	Description string `json:"description" gorm:"size:500"`

	// Resources
	VCPUCount int `json:"vcpu_count" gorm:"not null;default:1"`
	MemoryMB  int `json:"memory_mb" gorm:"not null;default:512"`

	// Network
	IPAddress string `json:"ip_address" gorm:"index;size:15"`

	// State
	Status       VMStatus `json:"status" gorm:"not null;default:'pending';size:20"`
	ErrorMessage string   `json:"error_message,omitempty" gorm:"type:text"`

	// Infrastructure
	Host       string `json:"host,omitempty" gorm:"size:100"`
	KernelPath string `json:"kernel_path,omitempty" gorm:"type:text"`
	RootfsPath string `json:"rootfs_path,omitempty" gorm:"type:text"`

	// Ownership
	UserID int  `json:"user_id" gorm:"not null;index"`
	User   User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for VM
func (VM) TableName() string {
	return "vms"
}

// CreateVMRequest represents the request to create a new VM
type CreateVMRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=64"`
	Description string `json:"description" binding:"max=500"`
	VCPUCount   int    `json:"vcpu_count" binding:"required,min=1,max=32"`
	MemoryMB    int    `json:"memory_mb" binding:"required,min=128,max=32768"`
	KernelPath  string `json:"kernel_path" binding:"omitempty"`
	RootfsPath  string `json:"rootfs_path" binding:"omitempty"`
}

// UpdateVMRequest represents the request to update a VM
type UpdateVMRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=64"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// VMResponse represents a VM in API responses
type VMResponse struct {
	ID           int       `json:"id"`
	VMID         string    `json:"vm_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	VCPUCount    int       `json:"vcpu_count"`
	MemoryMB     int       `json:"memory_mb"`
	IPAddress    string    `json:"ip_address,omitempty"`
	Status       VMStatus  `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	Host         string    `json:"host,omitempty"`
	UserID       int       `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// VMListResponse represents paginated VMs
type VMListResponse struct {
	Items      []VMResponse `json:"items"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}
