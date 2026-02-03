package worker

// Task type constants
const (
	TypeCreateVM  = "vm:create"
	TypeDeleteVM  = "vm:delete"
	TypeStartVM   = "vm:start"
	TypeStopVM    = "vm:stop"
	TypeRestartVM = "vm:restart"
)

// CreateVMPayload represents the payload for creating a VM
type CreateVMPayload struct {
	VMID        string `json:"vm_id"`
	UserID      uint   `json:"user_id"`
	Name        string `json:"name"`
	VCPUCount   int    `json:"vcpu_count"`
	MemoryMB    int    `json:"memory_mb"`
	KernelPath  string `json:"kernel_path"`
	RootfsPath  string `json:"rootfs_path"`
	Description string `json:"description,omitempty"`
}

// DeleteVMPayload represents the payload for deleting a VM
type DeleteVMPayload struct {
	VMID       string `json:"vm_id"`
	UserID     uint   `json:"user_id"`
	IPAddress  string `json:"ip_address,omitempty"`  // To release IP
	DeployPath string `json:"deploy_path,omitempty"` // Firecracker deploy path
}

// StartVMPayload represents the payload for starting a VM
type StartVMPayload struct {
	VMID       string `json:"vm_id"`
	UserID     uint   `json:"user_id"`
	IPAddress  string `json:"ip_address"`
	DeployPath string `json:"deploy_path"`
}

// StopVMPayload represents the payload for stopping a VM
type StopVMPayload struct {
	VMID       string `json:"vm_id"`
	UserID     uint   `json:"user_id"`
	IPAddress  string `json:"ip_address"`
	DeployPath string `json:"deploy_path"`
}

// RestartVMPayload represents the payload for restarting a VM
type RestartVMPayload struct {
	VMID       string `json:"vm_id"`
	UserID     uint   `json:"user_id"`
	IPAddress  string `json:"ip_address"`
	DeployPath string `json:"deploy_path"`
}
