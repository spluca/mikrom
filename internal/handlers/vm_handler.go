package handlers

import (
	"net/http"
	"strconv"

	"github.com/apardo/mikrom-go/internal/models"
	"github.com/apardo/mikrom-go/internal/service"
	"github.com/gin-gonic/gin"
)

type VMHandler struct {
	vmService *service.VMService
}

func NewVMHandler(vmService *service.VMService) *VMHandler {
	return &VMHandler{
		vmService: vmService,
	}
}

// CreateVM godoc
// @Summary Create a new VM
// @Description Create a new Firecracker microVM
// @Tags VMs
// @Accept json
// @Produce json
// @Param vm body models.CreateVMRequest true "VM creation request"
// @Success 201 {object} models.VMResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /vms [post]
func (h *VMHandler) CreateVM(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	var req models.CreateVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	vm, err := h.vmService.CreateVM(req, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, vm)
}

// ListVMs godoc
// @Summary List VMs
// @Description List all VMs owned by the current user
// @Tags VMs
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} models.VMListResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /vms [get]
func (h *VMHandler) ListVMs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	result, err := h.vmService.ListVMs(userID.(int), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetVM godoc
// @Summary Get VM details
// @Description Get details of a specific VM
// @Tags VMs
// @Produce json
// @Param vm_id path string true "VM ID"
// @Success 200 {object} models.VMResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /vms/{vm_id} [get]
func (h *VMHandler) GetVM(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	vmID := c.Param("vm_id")
	vm, err := h.vmService.GetVM(vmID, userID.(int))
	if err != nil {
		if err.Error() == "unauthorized access to VM" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if vm == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "VM not found",
		})
		return
	}

	c.JSON(http.StatusOK, vm)
}

// UpdateVM godoc
// @Summary Update VM
// @Description Update VM metadata (name, description)
// @Tags VMs
// @Accept json
// @Produce json
// @Param vm_id path string true "VM ID"
// @Param vm body models.UpdateVMRequest true "VM update request"
// @Success 200 {object} models.VMResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /vms/{vm_id} [patch]
func (h *VMHandler) UpdateVM(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	vmID := c.Param("vm_id")

	var req models.UpdateVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	vm, err := h.vmService.UpdateVM(vmID, userID.(int), req)
	if err != nil {
		if err.Error() == "unauthorized access to VM" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if vm == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "VM not found",
		})
		return
	}

	c.JSON(http.StatusOK, vm)
}

// DeleteVM godoc
// @Summary Delete VM
// @Description Delete a VM
// @Tags VMs
// @Produce json
// @Param vm_id path string true "VM ID"
// @Success 202 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /vms/{vm_id} [delete]
func (h *VMHandler) DeleteVM(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	vmID := c.Param("vm_id")

	err := h.vmService.DeleteVM(vmID, userID.(int))
	if err != nil {
		if err.Error() == "unauthorized access to VM" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Access denied",
			})
			return
		}
		if err.Error() == "VM not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "VM not found",
			})
			return
		}
		if err.Error() == "VM is already being deleted" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: "VM is already being deleted",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "VM deletion queued",
		"vm_id":   vmID,
		"status":  "deleting",
	})
}

// StartVM godoc
// @Summary Start VM
// @Description Start a stopped VM
// @Tags VMs
// @Produce json
// @Param vm_id path string true "VM ID"
// @Success 202 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /vms/{vm_id}/start [post]
func (h *VMHandler) StartVM(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	vmID := c.Param("vm_id")

	err := h.vmService.StartVM(vmID, userID.(int))
	if err != nil {
		if err.Error() == "unauthorized access to VM" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Access denied",
			})
			return
		}
		if err.Error() == "VM not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "VM not found",
			})
			return
		}
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "VM start queued",
		"vm_id":   vmID,
		"status":  "starting",
	})
}

// StopVM godoc
// @Summary Stop VM
// @Description Stop a running VM
// @Tags VMs
// @Produce json
// @Param vm_id path string true "VM ID"
// @Success 202 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /vms/{vm_id}/stop [post]
func (h *VMHandler) StopVM(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	vmID := c.Param("vm_id")

	err := h.vmService.StopVM(vmID, userID.(int))
	if err != nil {
		if err.Error() == "unauthorized access to VM" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Access denied",
			})
			return
		}
		if err.Error() == "VM not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "VM not found",
			})
			return
		}
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "VM stop queued",
		"vm_id":   vmID,
		"status":  "stopping",
	})
}

// RestartVM godoc
// @Summary Restart VM
// @Description Restart a running VM
// @Tags VMs
// @Produce json
// @Param vm_id path string true "VM ID"
// @Success 202 {object} map[string]string
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /vms/{vm_id}/restart [post]
func (h *VMHandler) RestartVM(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Unauthorized",
		})
		return
	}

	vmID := c.Param("vm_id")

	err := h.vmService.RestartVM(vmID, userID.(int))
	if err != nil {
		if err.Error() == "unauthorized access to VM" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: "Access denied",
			})
			return
		}
		if err.Error() == "VM not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "VM not found",
			})
			return
		}
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "VM restart queued",
		"vm_id":   vmID,
		"status":  "restarting",
	})
}
