package firecracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Client handles Firecracker VM operations via Ansible
type Client struct {
	deployPath  string // Path to firecracker-deploy
	defaultHost string // Default Ansible host
}

// NewClient creates a new Firecracker client
func NewClient(deployPath, defaultHost string) *Client {
	return &Client{
		deployPath:  deployPath,
		defaultHost: defaultHost,
	}
}

// CreateVMParams contains parameters for creating a VM
type CreateVMParams struct {
	VMName     string
	VCPUCount  int
	MemoryMB   int
	IPAddress  string
	KernelPath string
	RootfsPath string
	SSHKeyPath string
}

// AnsibleResult represents the result of an Ansible playbook execution
type AnsibleResult struct {
	Success bool
	Output  string
	Error   string
}

// CreateVM creates a new Firecracker VM using Ansible
func (c *Client) CreateVM(ctx context.Context, params CreateVMParams) (*AnsibleResult, error) {
	// Build ansible-playbook command
	// Example: ansible-playbook -i inventory create_vm.yml -e "vm_name=srv-xxx vcpu_count=2 memory_mb=1024 ip_address=10.0.0.10"

	extraVars := map[string]interface{}{
		"vm_name":     params.VMName,
		"vcpu_count":  params.VCPUCount,
		"memory_mb":   params.MemoryMB,
		"ip_address":  params.IPAddress,
		"kernel_path": params.KernelPath,
		"rootfs_path": params.RootfsPath,
	}

	if params.SSHKeyPath != "" {
		extraVars["ssh_key_path"] = params.SSHKeyPath
	}

	extraVarsJSON, err := json.Marshal(extraVars)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extra vars: %w", err)
	}

	args := []string{
		"-i", c.defaultHost + ",", // Inventory (host with comma for inline inventory)
		c.deployPath + "/create_vm.yml",
		"-e", string(extraVarsJSON),
	}

	return c.runAnsiblePlaybook(ctx, args, 5*time.Minute)
}

// StartVM starts a Firecracker VM using Ansible
func (c *Client) StartVM(ctx context.Context, vmName, ipAddress string) (*AnsibleResult, error) {
	extraVars := map[string]interface{}{
		"vm_name":    vmName,
		"ip_address": ipAddress,
	}

	extraVarsJSON, err := json.Marshal(extraVars)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extra vars: %w", err)
	}

	args := []string{
		"-i", c.defaultHost + ",",
		c.deployPath + "/start_vm.yml",
		"-e", string(extraVarsJSON),
	}

	return c.runAnsiblePlaybook(ctx, args, 2*time.Minute)
}

// StopVM stops a Firecracker VM using Ansible
func (c *Client) StopVM(ctx context.Context, vmName, ipAddress string) (*AnsibleResult, error) {
	extraVars := map[string]interface{}{
		"vm_name":    vmName,
		"ip_address": ipAddress,
	}

	extraVarsJSON, err := json.Marshal(extraVars)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extra vars: %w", err)
	}

	args := []string{
		"-i", c.defaultHost + ",",
		c.deployPath + "/stop_vm.yml",
		"-e", string(extraVarsJSON),
	}

	return c.runAnsiblePlaybook(ctx, args, 2*time.Minute)
}

// CleanupVM deletes a Firecracker VM using Ansible
func (c *Client) CleanupVM(ctx context.Context, vmName, ipAddress string) (*AnsibleResult, error) {
	extraVars := map[string]interface{}{
		"vm_name":    vmName,
		"ip_address": ipAddress,
	}

	extraVarsJSON, err := json.Marshal(extraVars)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extra vars: %w", err)
	}

	args := []string{
		"-i", c.defaultHost + ",",
		c.deployPath + "/cleanup_vm.yml",
		"-e", string(extraVarsJSON),
	}

	return c.runAnsiblePlaybook(ctx, args, 3*time.Minute)
}

// runAnsiblePlaybook executes an ansible-playbook command with timeout
func (c *Client) runAnsiblePlaybook(ctx context.Context, args []string, timeout time.Duration) (*AnsibleResult, error) {
	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build command
	cmd := exec.CommandContext(execCtx, "ansible-playbook", args...)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err := cmd.Run()

	result := &AnsibleResult{
		Success: err == nil,
		Output:  stdout.String(),
		Error:   stderr.String(),
	}

	if err != nil {
		// Check if it was a timeout
		if execCtx.Err() == context.DeadlineExceeded {
			return result, fmt.Errorf("ansible playbook execution timed out after %v", timeout)
		}
		return result, fmt.Errorf("ansible playbook execution failed: %w (stderr: %s)", err, stderr.String())
	}

	return result, nil
}

// CheckHealth checks if Ansible is available and accessible
func (c *Client) CheckHealth(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "ansible-playbook", "--version")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ansible-playbook not available: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}
