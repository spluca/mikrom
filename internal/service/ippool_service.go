package service

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/spluca/mikrom/internal/models"
	"github.com/spluca/mikrom/internal/repository"
)

type IPPoolService struct {
	ipPoolRepo *repository.IPPoolRepository
}

func NewIPPoolService(ipPoolRepo *repository.IPPoolRepository) *IPPoolService {
	return &IPPoolService{
		ipPoolRepo: ipPoolRepo,
	}
}

// CreateIPPoolRequest represents the request to create a new IP pool
type CreateIPPoolRequest struct {
	Name    string `json:"name" binding:"required,min=1,max=50"`
	Network string `json:"network" binding:"required"`
	CIDR    string `json:"cidr" binding:"required"`
	Gateway string `json:"gateway" binding:"required"`
	StartIP string `json:"start_ip" binding:"required"`
	EndIP   string `json:"end_ip" binding:"required"`
}

// UpdateIPPoolRequest represents the request to update an IP pool
type UpdateIPPoolRequest struct {
	Name     *string `json:"name" binding:"omitempty,min=1,max=50"`
	IsActive *bool   `json:"is_active"`
}

// IPPoolResponse represents an IP pool in API responses
type IPPoolResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Network   string `json:"network"`
	CIDR      string `json:"cidr"`
	Gateway   string `json:"gateway"`
	StartIP   string `json:"start_ip"`
	EndIP     string `json:"end_ip"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// IPPoolStatsResponse represents IP pool statistics
type IPPoolStatsResponse struct {
	PoolID       int     `json:"pool_id"`
	PoolName     string  `json:"pool_name"`
	Total        int64   `json:"total"`
	Allocated    int64   `json:"allocated"`
	Available    int64   `json:"available"`
	UsagePercent float64 `json:"usage_percent"`
}

// CreateIPPool creates a new IP pool and allocates IP addresses
func (s *IPPoolService) CreateIPPool(req CreateIPPoolRequest) (*models.IPPool, error) {
	// Validate IP addresses and CIDR
	if err := s.validateIPPool(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create IP pool record
	pool := &models.IPPool{
		Name:     req.Name,
		Network:  req.Network,
		CIDR:     req.CIDR,
		Gateway:  req.Gateway,
		StartIP:  req.StartIP,
		EndIP:    req.EndIP,
		IsActive: true,
	}

	if err := s.ipPoolRepo.CreatePool(pool); err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Generate and create IP allocations
	ips, err := s.generateIPRange(req.StartIP, req.EndIP)
	if err != nil {
		return nil, fmt.Errorf("failed to generate IP range: %w", err)
	}

	allocations := make([]models.IPAllocation, len(ips))
	for i, ip := range ips {
		allocations[i] = models.IPAllocation{
			PoolID:    pool.ID,
			IPAddress: ip,
			VMID:      "",
			IsActive:  false,
		}
	}

	if err := s.ipPoolRepo.CreateAllocations(allocations); err != nil {
		return nil, fmt.Errorf("failed to create IP allocations: %w", err)
	}

	return pool, nil
}

// GetIPPool retrieves an IP pool by ID
func (s *IPPoolService) GetIPPool(id int) (*models.IPPool, error) {
	pool, err := s.ipPoolRepo.FindPoolByID(uint(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find pool: %w", err)
	}
	return pool, nil
}

// ListIPPools retrieves all IP pools with pagination
func (s *IPPoolService) ListIPPools(page, pageSize int) ([]IPPoolResponse, int64, error) {
	offset := (page - 1) * pageSize

	pools, total, err := s.ipPoolRepo.ListPools(offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list pools: %w", err)
	}

	responses := make([]IPPoolResponse, len(pools))
	for i, pool := range pools {
		responses[i] = IPPoolResponse{
			ID:        pool.ID,
			Name:      pool.Name,
			Network:   pool.Network,
			CIDR:      pool.CIDR,
			Gateway:   pool.Gateway,
			StartIP:   pool.StartIP,
			EndIP:     pool.EndIP,
			IsActive:  pool.IsActive,
			CreatedAt: pool.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: pool.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return responses, total, nil
}

// UpdateIPPool updates an IP pool
func (s *IPPoolService) UpdateIPPool(id int, req UpdateIPPoolRequest) (*models.IPPool, error) {
	pool, err := s.ipPoolRepo.FindPoolByID(uint(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find pool: %w", err)
	}

	if pool == nil {
		return nil, fmt.Errorf("pool not found")
	}

	// Update fields
	if req.Name != nil {
		pool.Name = *req.Name
	}
	if req.IsActive != nil {
		pool.IsActive = *req.IsActive
	}

	if err := s.ipPoolRepo.UpdatePool(pool); err != nil {
		return nil, fmt.Errorf("failed to update pool: %w", err)
	}

	return pool, nil
}

// DeleteIPPool deletes an IP pool
func (s *IPPoolService) DeleteIPPool(id int) error {
	// Check if pool has any allocated IPs
	_, allocated, _, err := s.ipPoolRepo.GetPoolStats(id)
	if err != nil {
		return fmt.Errorf("failed to get pool stats: %w", err)
	}

	if allocated > 0 {
		return fmt.Errorf("cannot delete pool with allocated IPs (%d IPs in use)", allocated)
	}

	if err := s.ipPoolRepo.DeletePool(uint(id)); err != nil {
		return fmt.Errorf("failed to delete pool: %w", err)
	}

	return nil
}

// GetPoolStats retrieves statistics for an IP pool
func (s *IPPoolService) GetPoolStats(id int) (*IPPoolStatsResponse, error) {
	pool, err := s.ipPoolRepo.FindPoolByID(uint(id))
	if err != nil {
		return nil, fmt.Errorf("failed to find pool: %w", err)
	}

	if pool == nil {
		return nil, fmt.Errorf("pool not found")
	}

	total, allocated, available, err := s.ipPoolRepo.GetPoolStats(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool stats: %w", err)
	}

	usagePercent := 0.0
	if total > 0 {
		usagePercent = float64(allocated) / float64(total) * 100
	}

	return &IPPoolStatsResponse{
		PoolID:       pool.ID,
		PoolName:     pool.Name,
		Total:        total,
		Allocated:    allocated,
		Available:    available,
		UsagePercent: math.Round(usagePercent*100) / 100,
	}, nil
}

// validateIPPool validates IP pool creation request
func (s *IPPoolService) validateIPPool(req CreateIPPoolRequest) error {
	// Validate CIDR format
	_, _, err := net.ParseCIDR(req.CIDR)
	if err != nil {
		return fmt.Errorf("invalid CIDR format: %w", err)
	}

	// Validate IP addresses
	if net.ParseIP(req.Gateway) == nil {
		return fmt.Errorf("invalid gateway IP address")
	}

	startIP := net.ParseIP(req.StartIP)
	if startIP == nil {
		return fmt.Errorf("invalid start IP address")
	}

	endIP := net.ParseIP(req.EndIP)
	if endIP == nil {
		return fmt.Errorf("invalid end IP address")
	}

	// Validate range
	if ipToInt(startIP) > ipToInt(endIP) {
		return fmt.Errorf("start IP must be less than or equal to end IP")
	}

	return nil
}

// generateIPRange generates a list of IP addresses between start and end
func (s *IPPoolService) generateIPRange(startIP, endIP string) ([]string, error) {
	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)

	if start == nil || end == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	startInt := ipToInt(start)
	endInt := ipToInt(end)

	if startInt > endInt {
		return nil, fmt.Errorf("start IP must be less than or equal to end IP")
	}

	count := endInt - startInt + 1
	if count > 1000 {
		return nil, fmt.Errorf("IP range too large (max 1000 IPs)")
	}

	ips := make([]string, count)
	for i := uint32(0); i < count; i++ {
		ips[i] = intToIP(startInt + i)
	}

	return ips, nil
}

// ipToInt converts an IP address to uint32
func ipToInt(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// intToIP converts uint32 to IP address string
func intToIP(n uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

// GetAllPoolStats retrieves statistics for all IP pools
func (s *IPPoolService) GetAllPoolStats() ([]IPPoolStatsResponse, error) {
	pools, _, err := s.ipPoolRepo.ListPools(0, 100) // Get all pools
	if err != nil {
		return nil, fmt.Errorf("failed to list pools: %w", err)
	}

	stats := make([]IPPoolStatsResponse, len(pools))
	for i, pool := range pools {
		poolStats, err := s.GetPoolStats(pool.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for pool %d: %w", pool.ID, err)
		}
		stats[i] = *poolStats
	}

	return stats, nil
}

// ParseCIDR parses a CIDR and returns network info
func ParseCIDR(cidr string) (network, firstIP, lastIP, broadcast string, totalHosts int, err error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", "", "", 0, err
	}

	// Calculate network address
	network = ipnet.IP.String()

	// Calculate first usable IP (network + 1)
	firstIPInt := ipToInt(ipnet.IP) + 1
	firstIP = intToIP(firstIPInt)

	// Calculate broadcast address
	mask := ipnet.Mask
	ones, bits := mask.Size()
	hostBits := bits - ones
	totalHosts = (1 << hostBits) - 2 // Subtract network and broadcast

	broadcastInt := ipToInt(ipnet.IP) | ^ipToInt(net.IP(mask))
	broadcast = intToIP(broadcastInt)

	// Last usable IP (broadcast - 1)
	lastIPInt := broadcastInt - 1
	lastIP = intToIP(lastIPInt)

	return network, firstIP, lastIP, broadcast, totalHosts, nil
}

// SuggestIPRange suggests a reasonable IP range for a CIDR
func (s *IPPoolService) SuggestIPRange(cidr string) (startIP, endIP string, err error) {
	_, firstIP, lastIP, _, totalHosts, err := ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}

	// Reserve first 10 IPs for infrastructure (gateway, DNS, etc.)
	startIPInt := ipToInt(net.ParseIP(firstIP)) + 9
	startIP = intToIP(startIPInt)

	// Use up to 90% of available IPs
	maxIPs := int(float64(totalHosts) * 0.9)
	if maxIPs > 254 {
		maxIPs = 254 // Reasonable limit
	}

	endIPInt := startIPInt + uint32(maxIPs) - 1
	lastIPInt := ipToInt(net.ParseIP(lastIP))

	if endIPInt > lastIPInt {
		endIPInt = lastIPInt
	}

	endIP = intToIP(endIPInt)

	return startIP, endIP, nil
}

// Helper to convert string to int
func atoiOrZero(s string) int {
	val, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return val
}
