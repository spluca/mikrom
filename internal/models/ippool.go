package models

import (
	"time"
)

// IPPool represents a pool of IP addresses for VMs
type IPPool struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex;not null;size:50"`
	Network   string    `json:"network" gorm:"not null;size:50"`
	CIDR      string    `json:"cidr" gorm:"not null;size:20"`
	Gateway   string    `json:"gateway" gorm:"not null;size:15"`
	StartIP   string    `json:"start_ip" gorm:"not null;size:15"`
	EndIP     string    `json:"end_ip" gorm:"not null;size:15"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for IPPool
func (IPPool) TableName() string {
	return "ip_pools"
}

// IPAllocation represents an IP address allocation to a VM
type IPAllocation struct {
	ID          int        `json:"id" gorm:"primaryKey"`
	PoolID      int        `json:"pool_id" gorm:"not null;index"`
	Pool        IPPool     `json:"-" gorm:"foreignKey:PoolID"`
	VMID        string     `json:"vm_id" gorm:"not null;index;size:50"`
	IPAddress   string     `json:"ip_address" gorm:"not null;size:15"`
	IsActive    bool       `json:"is_active" gorm:"default:true;index"`
	AllocatedAt time.Time  `json:"allocated_at"`
	ReleasedAt  *time.Time `json:"released_at,omitempty"`
}

// TableName specifies the table name for IPAllocation
func (IPAllocation) TableName() string {
	return "ip_allocations"
}
