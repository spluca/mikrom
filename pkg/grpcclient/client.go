package grpcclient

import (
	"context"
	"fmt"
	"time"

	pb "github.com/apardo/mikrom-go/api/proto/firecracker/v1"
	"github.com/apardo/mikrom-go/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client handles gRPC communication with firecracker-agent
type Client struct {
	conn   *grpc.ClientConn
	client pb.FirecrackerAgentClient
	addr   string
}

// NewClient creates a new gRPC client for firecracker-agent
func NewClient(addr string) (*Client, error) {
	// Create gRPC connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to firecracker-agent at %s: %w", addr, err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewFirecrackerAgentClient(conn),
		addr:   addr,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateVMParams contains parameters for creating a VM
type CreateVMParams struct {
	VMName     string
	VCPUCount  int32
	MemoryMB   int32
	IPAddress  string
	KernelPath string
	RootfsPath string
	Metadata   map[string]string
}

// VMResponse represents a generic VM operation response
type VMResponse struct {
	Success bool
	VMID    string
	State   string
	Error   string
}

// MapVMStateToStatus maps gRPC VMState to mikrom-go VMStatus
func MapVMStateToStatus(state pb.VMState) models.VMStatus {
	switch state {
	case pb.VMState_VM_STATE_CREATING:
		return models.VMStatusProvisioning
	case pb.VMState_VM_STATE_RUNNING:
		return models.VMStatusRunning
	case pb.VMState_VM_STATE_STOPPING:
		return models.VMStatusStopping
	case pb.VMState_VM_STATE_STOPPED:
		return models.VMStatusStopped
	case pb.VMState_VM_STATE_DELETING:
		return models.VMStatusDeleting
	case pb.VMState_VM_STATE_ERROR:
		return models.VMStatusError
	default:
		return models.VMStatusError
	}
}

// GetVMStatus extracts VMStatus from VMResponse
func (r *VMResponse) GetVMStatus() models.VMStatus {
	// Parse the string state back to enum
	switch r.State {
	case "VM_STATE_CREATING":
		return models.VMStatusProvisioning
	case "VM_STATE_RUNNING":
		return models.VMStatusRunning
	case "VM_STATE_STOPPING":
		return models.VMStatusStopping
	case "VM_STATE_STOPPED":
		return models.VMStatusStopped
	case "VM_STATE_DELETING":
		return models.VMStatusDeleting
	case "VM_STATE_ERROR":
		return models.VMStatusError
	default:
		return models.VMStatusError
	}
}

// CreateVM creates a new Firecracker VM using gRPC
func (c *Client) CreateVM(ctx context.Context, params CreateVMParams) (*VMResponse, error) {
	req := &pb.CreateVMRequest{
		VmId:       params.VMName,
		VcpuCount:  params.VCPUCount,
		MemoryMb:   params.MemoryMB,
		IpAddress:  params.IPAddress,
		KernelPath: params.KernelPath,
		RootfsPath: params.RootfsPath,
		Metadata:   params.Metadata,
	}

	resp, err := c.client.CreateVM(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	return &VMResponse{
		Success: resp.ErrorMessage == "",
		VMID:    resp.VmId,
		State:   resp.State.String(),
		Error:   resp.ErrorMessage,
	}, nil
}

// StartVM starts a Firecracker VM using gRPC
func (c *Client) StartVM(ctx context.Context, vmID string) (*VMResponse, error) {
	req := &pb.StartVMRequest{
		VmId: vmID,
	}

	resp, err := c.client.StartVM(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start VM: %w", err)
	}

	return &VMResponse{
		Success: resp.ErrorMessage == "",
		VMID:    resp.VmId,
		State:   resp.State.String(),
		Error:   resp.ErrorMessage,
	}, nil
}

// StopVM stops a Firecracker VM using gRPC
func (c *Client) StopVM(ctx context.Context, vmID string, force bool) (*VMResponse, error) {
	req := &pb.StopVMRequest{
		VmId:  vmID,
		Force: force,
	}

	resp, err := c.client.StopVM(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to stop VM: %w", err)
	}

	return &VMResponse{
		Success: resp.ErrorMessage == "",
		VMID:    resp.VmId,
		State:   resp.State.String(),
		Error:   resp.ErrorMessage,
	}, nil
}

// DeleteVM deletes a Firecracker VM using gRPC
func (c *Client) DeleteVM(ctx context.Context, vmID string) (*VMResponse, error) {
	req := &pb.DeleteVMRequest{
		VmId: vmID,
	}

	resp, err := c.client.DeleteVM(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete VM: %w", err)
	}

	return &VMResponse{
		Success: resp.Success,
		VMID:    resp.VmId,
		Error:   resp.ErrorMessage,
	}, nil
}

// GetVM retrieves VM information using gRPC
func (c *Client) GetVM(ctx context.Context, vmID string) (*pb.VMInfo, error) {
	req := &pb.GetVMRequest{
		VmId: vmID,
	}

	resp, err := c.client.GetVM(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	return resp.Vm, nil
}

// HealthCheck checks if the firecracker-agent is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	req := &pb.HealthCheckRequest{}

	resp, err := c.client.HealthCheck(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !resp.Healthy {
		return fmt.Errorf("firecracker-agent is not healthy")
	}

	return nil
}
