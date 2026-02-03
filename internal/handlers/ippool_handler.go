package handlers

import (
	"net/http"
	"strconv"

	"github.com/apardo/mikrom-go/internal/service"
	"github.com/gin-gonic/gin"
)

type IPPoolHandler struct {
	ipPoolService *service.IPPoolService
}

func NewIPPoolHandler(ipPoolService *service.IPPoolService) *IPPoolHandler {
	return &IPPoolHandler{
		ipPoolService: ipPoolService,
	}
}

// CreateIPPool creates a new IP pool
// POST /api/v1/ippools
func (h *IPPoolHandler) CreateIPPool(c *gin.Context) {
	var req service.CreateIPPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pool, err := h.ipPoolService.CreateIPPool(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pool)
}

// ListIPPools lists all IP pools with pagination
// GET /api/v1/ippools
func (h *IPPoolHandler) ListIPPools(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	pools, total, err := h.ipPoolService.ListIPPools(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       pools,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetIPPool retrieves an IP pool by ID
// GET /api/v1/ippools/:id
func (h *IPPoolHandler) GetIPPool(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool ID"})
		return
	}

	pool, err := h.ipPoolService.GetIPPool(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pool not found"})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// UpdateIPPool updates an IP pool
// PATCH /api/v1/ippools/:id
func (h *IPPoolHandler) UpdateIPPool(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool ID"})
		return
	}

	var req service.UpdateIPPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pool, err := h.ipPoolService.UpdateIPPool(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// DeleteIPPool deletes an IP pool
// DELETE /api/v1/ippools/:id
func (h *IPPoolHandler) DeleteIPPool(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool ID"})
		return
	}

	if err := h.ipPoolService.DeleteIPPool(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetPoolStats retrieves statistics for an IP pool
// GET /api/v1/ippools/:id/stats
func (h *IPPoolHandler) GetPoolStats(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool ID"})
		return
	}

	stats, err := h.ipPoolService.GetPoolStats(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAllPoolStats retrieves statistics for all IP pools
// GET /api/v1/ippools/stats
func (h *IPPoolHandler) GetAllPoolStats(c *gin.Context) {
	stats, err := h.ipPoolService.GetAllPoolStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pools": stats,
	})
}

// SuggestIPRange suggests an IP range for a given CIDR
// POST /api/v1/ippools/suggest-range
func (h *IPPoolHandler) SuggestIPRange(c *gin.Context) {
	var req struct {
		CIDR string `json:"cidr" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startIP, endIP, err := h.ipPoolService.SuggestIPRange(req.CIDR)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	network, firstIP, lastIP, broadcast, totalHosts, err := service.ParseCIDR(req.CIDR)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cidr":              req.CIDR,
		"network_address":   network,
		"first_usable_ip":   firstIP,
		"last_usable_ip":    lastIP,
		"broadcast_address": broadcast,
		"total_hosts":       totalHosts,
		"suggested_start":   startIP,
		"suggested_end":     endIP,
	})
}
