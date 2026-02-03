package worker

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateVMPayload_JSONMarshal(t *testing.T) {
	payload := &CreateVMPayload{
		VMID:        "srv-12345678",
		UserID:      1,
		Name:        "Test VM",
		VCPUCount:   2,
		MemoryMB:    1024,
		KernelPath:  "/path/to/kernel",
		RootfsPath:  "/path/to/rootfs",
		Description: "Test Description",
	}

	data, err := json.Marshal(payload)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal and verify
	var decoded CreateVMPayload
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, payload.VMID, decoded.VMID)
	assert.Equal(t, payload.UserID, decoded.UserID)
	assert.Equal(t, payload.Name, decoded.Name)
	assert.Equal(t, payload.VCPUCount, decoded.VCPUCount)
	assert.Equal(t, payload.MemoryMB, decoded.MemoryMB)
	assert.Equal(t, payload.KernelPath, decoded.KernelPath)
	assert.Equal(t, payload.RootfsPath, decoded.RootfsPath)
	assert.Equal(t, payload.Description, decoded.Description)
}

func TestCreateVMPayload_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"vm_id": "srv-12345678",
		"user_id": 1,
		"name": "Test VM",
		"vcpu_count": 2,
		"memory_mb": 1024,
		"kernel_path": "/path/to/kernel",
		"rootfs_path": "/path/to/rootfs",
		"description": "Test Description"
	}`

	var payload CreateVMPayload
	err := json.Unmarshal([]byte(jsonData), &payload)

	assert.NoError(t, err)
	assert.Equal(t, "srv-12345678", payload.VMID)
	assert.Equal(t, uint(1), payload.UserID)
	assert.Equal(t, "Test VM", payload.Name)
	assert.Equal(t, 2, payload.VCPUCount)
	assert.Equal(t, 1024, payload.MemoryMB)
	assert.Equal(t, "/path/to/kernel", payload.KernelPath)
	assert.Equal(t, "/path/to/rootfs", payload.RootfsPath)
	assert.Equal(t, "Test Description", payload.Description)
}

func TestCreateVMPayload_OmitEmptyDescription(t *testing.T) {
	payload := &CreateVMPayload{
		VMID:       "srv-12345678",
		UserID:     1,
		Name:       "Test VM",
		VCPUCount:  2,
		MemoryMB:   1024,
		KernelPath: "/path/to/kernel",
		RootfsPath: "/path/to/rootfs",
		// Description is empty
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	// Description should not be in JSON when empty (omitempty)
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	_, hasDescription := decoded["description"]
	assert.False(t, hasDescription, "Empty description should be omitted from JSON")
}

func TestDeleteVMPayload_JSONMarshal(t *testing.T) {
	payload := &DeleteVMPayload{
		VMID:       "srv-12345678",
		UserID:     1,
		IPAddress:  "192.168.1.10",
		DeployPath: "/deploy/path",
	}

	data, err := json.Marshal(payload)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded DeleteVMPayload
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, payload.VMID, decoded.VMID)
	assert.Equal(t, payload.UserID, decoded.UserID)
	assert.Equal(t, payload.IPAddress, decoded.IPAddress)
	assert.Equal(t, payload.DeployPath, decoded.DeployPath)
}

func TestDeleteVMPayload_OmitEmptyFields(t *testing.T) {
	payload := &DeleteVMPayload{
		VMID:   "srv-12345678",
		UserID: 1,
		// IPAddress and DeployPath are empty
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	_, hasIP := decoded["ip_address"]
	_, hasPath := decoded["deploy_path"]
	assert.False(t, hasIP, "Empty ip_address should be omitted")
	assert.False(t, hasPath, "Empty deploy_path should be omitted")
}

func TestStartVMPayload_JSONMarshal(t *testing.T) {
	payload := &StartVMPayload{
		VMID:       "srv-12345678",
		UserID:     1,
		IPAddress:  "192.168.1.10",
		DeployPath: "/deploy/path",
	}

	data, err := json.Marshal(payload)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded StartVMPayload
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, payload.VMID, decoded.VMID)
	assert.Equal(t, payload.UserID, decoded.UserID)
	assert.Equal(t, payload.IPAddress, decoded.IPAddress)
	assert.Equal(t, payload.DeployPath, decoded.DeployPath)
}

func TestStopVMPayload_JSONMarshal(t *testing.T) {
	payload := &StopVMPayload{
		VMID:       "srv-12345678",
		UserID:     1,
		IPAddress:  "192.168.1.10",
		DeployPath: "/deploy/path",
	}

	data, err := json.Marshal(payload)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded StopVMPayload
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, payload.VMID, decoded.VMID)
	assert.Equal(t, payload.UserID, decoded.UserID)
	assert.Equal(t, payload.IPAddress, decoded.IPAddress)
	assert.Equal(t, payload.DeployPath, decoded.DeployPath)
}

func TestRestartVMPayload_JSONMarshal(t *testing.T) {
	payload := &RestartVMPayload{
		VMID:       "srv-12345678",
		UserID:     1,
		IPAddress:  "192.168.1.10",
		DeployPath: "/deploy/path",
	}

	data, err := json.Marshal(payload)

	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded RestartVMPayload
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, payload.VMID, decoded.VMID)
	assert.Equal(t, payload.UserID, decoded.UserID)
	assert.Equal(t, payload.IPAddress, decoded.IPAddress)
	assert.Equal(t, payload.DeployPath, decoded.DeployPath)
}

func TestTaskTypeConstants(t *testing.T) {
	// Verify task type constants are correctly defined
	assert.Equal(t, "vm:create", TypeCreateVM)
	assert.Equal(t, "vm:delete", TypeDeleteVM)
	assert.Equal(t, "vm:start", TypeStartVM)
	assert.Equal(t, "vm:stop", TypeStopVM)
	assert.Equal(t, "vm:restart", TypeRestartVM)

	// Verify they're all unique
	types := []string{TypeCreateVM, TypeDeleteVM, TypeStartVM, TypeStopVM, TypeRestartVM}
	uniqueTypes := make(map[string]bool)
	for _, taskType := range types {
		assert.False(t, uniqueTypes[taskType], "Task type %s is duplicated", taskType)
		uniqueTypes[taskType] = true
	}
	assert.Equal(t, 5, len(uniqueTypes))
}

func TestCreateVMPayload_EmptyFields(t *testing.T) {
	payload := &CreateVMPayload{}

	data, err := json.Marshal(payload)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded CreateVMPayload
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
}

func TestPayload_RoundTrip(t *testing.T) {
	// Test round-trip for all payload types
	tests := []struct {
		name    string
		payload interface{}
	}{
		{"CreateVM", &CreateVMPayload{VMID: "srv-test", UserID: 1}},
		{"DeleteVM", &DeleteVMPayload{VMID: "srv-test", UserID: 1}},
		{"StartVM", &StartVMPayload{VMID: "srv-test", UserID: 1}},
		{"StopVM", &StopVMPayload{VMID: "srv-test", UserID: 1}},
		{"RestartVM", &RestartVMPayload{VMID: "srv-test", UserID: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.payload)
			assert.NoError(t, err)
			assert.NotEmpty(t, data)

			// Unmarshal
			var decoded interface{}
			switch tt.payload.(type) {
			case *CreateVMPayload:
				decoded = &CreateVMPayload{}
			case *DeleteVMPayload:
				decoded = &DeleteVMPayload{}
			case *StartVMPayload:
				decoded = &StartVMPayload{}
			case *StopVMPayload:
				decoded = &StopVMPayload{}
			case *RestartVMPayload:
				decoded = &RestartVMPayload{}
			}

			err = json.Unmarshal(data, decoded)
			assert.NoError(t, err)

			// Re-marshal and compare
			data2, err := json.Marshal(decoded)
			assert.NoError(t, err)
			assert.JSONEq(t, string(data), string(data2))
		})
	}
}

func TestPayload_InvalidJSON(t *testing.T) {
	invalidJSON := `{"invalid json`

	var payload CreateVMPayload
	err := json.Unmarshal([]byte(invalidJSON), &payload)
	assert.Error(t, err)
}

func TestPayload_WrongTypes(t *testing.T) {
	// Test with wrong types in JSON
	jsonData := `{
		"vm_id": 12345,
		"user_id": "not-a-number",
		"vcpu_count": "two"
	}`

	var payload CreateVMPayload
	err := json.Unmarshal([]byte(jsonData), &payload)
	assert.Error(t, err)
}
