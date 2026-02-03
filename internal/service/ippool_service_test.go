package service

import (
	"net"
	"testing"

	"github.com/apardo/mikrom-go/internal/models"
	"github.com/apardo/mikrom-go/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIPPool_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	req := CreateIPPoolRequest{
		Name:    "Test Pool",
		Network: "192.168.1.0",
		CIDR:    "192.168.1.0/24",
		Gateway: "192.168.1.1",
		StartIP: "192.168.1.10",
		EndIP:   "192.168.1.20",
	}

	pool, err := ipPoolService.CreateIPPool(req)

	assert.NoError(t, err)
	require.NotNil(t, pool)
	assert.NotZero(t, pool.ID)
	assert.Equal(t, req.Name, pool.Name)
	assert.Equal(t, req.Network, pool.Network)
	assert.Equal(t, req.CIDR, pool.CIDR)
	assert.Equal(t, req.Gateway, pool.Gateway)
	assert.Equal(t, req.StartIP, pool.StartIP)
	assert.Equal(t, req.EndIP, pool.EndIP)
	assert.True(t, pool.IsActive)

	// Verify IP allocations were created (11 IPs: 10 through 20 inclusive)
	var allocationCount int64
	db.Model(&models.IPAllocation{}).Where("pool_id = ?", pool.ID).Count(&allocationCount)
	assert.Equal(t, int64(11), allocationCount)
}

func TestCreateIPPool_InvalidCIDR(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	req := CreateIPPoolRequest{
		Name:    "Test Pool",
		Network: "192.168.1.0",
		CIDR:    "invalid-cidr",
		Gateway: "192.168.1.1",
		StartIP: "192.168.1.10",
		EndIP:   "192.168.1.20",
	}

	pool, err := ipPoolService.CreateIPPool(req)

	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestCreateIPPool_InvalidGateway(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	req := CreateIPPoolRequest{
		Name:    "Test Pool",
		Network: "192.168.1.0",
		CIDR:    "192.168.1.0/24",
		Gateway: "invalid-ip",
		StartIP: "192.168.1.10",
		EndIP:   "192.168.1.20",
	}

	pool, err := ipPoolService.CreateIPPool(req)

	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "invalid gateway")
}

func TestCreateIPPool_InvalidStartIP(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	req := CreateIPPoolRequest{
		Name:    "Test Pool",
		Network: "192.168.1.0",
		CIDR:    "192.168.1.0/24",
		Gateway: "192.168.1.1",
		StartIP: "invalid-ip",
		EndIP:   "192.168.1.20",
	}

	pool, err := ipPoolService.CreateIPPool(req)

	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "invalid start IP")
}

func TestCreateIPPool_StartIPGreaterThanEndIP(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	req := CreateIPPoolRequest{
		Name:    "Test Pool",
		Network: "192.168.1.0",
		CIDR:    "192.168.1.0/24",
		Gateway: "192.168.1.1",
		StartIP: "192.168.1.50",
		EndIP:   "192.168.1.20",
	}

	pool, err := ipPoolService.CreateIPPool(req)

	assert.Error(t, err)
	assert.Nil(t, pool)
	assert.Contains(t, err.Error(), "start IP must be less than")
}

func TestGetIPPool_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	// Create a pool
	expectedPool := &models.IPPool{
		Name:     "Test Pool",
		Network:  "192.168.1.0",
		CIDR:     "192.168.1.0/24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.20",
		IsActive: true,
	}
	db.Create(expectedPool)

	pool, err := ipPoolService.GetIPPool(expectedPool.ID)

	assert.NoError(t, err)
	require.NotNil(t, pool)
	assert.Equal(t, expectedPool.ID, pool.ID)
	assert.Equal(t, expectedPool.Name, pool.Name)
}

func TestGetIPPool_NotFound(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	pool, err := ipPoolService.GetIPPool(999)

	assert.Error(t, err)
	assert.Nil(t, pool)
}

func TestListIPPools_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	// Create multiple pools
	for i := 1; i <= 5; i++ {
		pool := &models.IPPool{
			Name:     "Pool " + string(rune('0'+i)),
			Network:  "192.168.1.0",
			CIDR:     "192.168.1.0/24",
			Gateway:  "192.168.1.1",
			StartIP:  "192.168.1.10",
			EndIP:    "192.168.1.20",
			IsActive: true,
		}
		db.Create(pool)
	}

	responses, total, err := ipPoolService.ListIPPools(1, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, responses, 5)
}

func TestListIPPools_Pagination(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	// Create 10 pools
	for i := 1; i <= 10; i++ {
		pool := &models.IPPool{
			Name:     "Pool " + string(rune('0'+i)),
			Network:  "192.168.1.0",
			CIDR:     "192.168.1.0/24",
			Gateway:  "192.168.1.1",
			StartIP:  "192.168.1.10",
			EndIP:    "192.168.1.20",
			IsActive: true,
		}
		db.Create(pool)
	}

	// Get first page
	responses, total, err := ipPoolService.ListIPPools(1, 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, responses, 3)

	// Get second page
	responses, total, err = ipPoolService.ListIPPools(2, 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, responses, 3)
}

func TestUpdateIPPool_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	// Create a pool
	pool := &models.IPPool{
		Name:     "Original Name",
		Network:  "192.168.1.0",
		CIDR:     "192.168.1.0/24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.20",
		IsActive: true,
	}
	db.Create(pool)

	// Update pool
	newName := "Updated Name"
	isActive := false
	req := UpdateIPPoolRequest{
		Name:     &newName,
		IsActive: &isActive,
	}

	updated, err := ipPoolService.UpdateIPPool(pool.ID, req)

	assert.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, newName, updated.Name)
	assert.Equal(t, isActive, updated.IsActive)
}

func TestUpdateIPPool_NotFound(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	newName := "Updated Name"
	req := UpdateIPPoolRequest{Name: &newName}

	updated, err := ipPoolService.UpdateIPPool(999, req)

	assert.Error(t, err)
	assert.Nil(t, updated)
}

func TestDeleteIPPool_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	// Create a pool
	pool := &models.IPPool{
		Name:     "Test Pool",
		Network:  "192.168.1.0",
		CIDR:     "192.168.1.0/24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.20",
		IsActive: true,
	}
	db.Create(pool)

	// Create some unallocated IP addresses
	db.Create(&models.IPAllocation{
		PoolID:    pool.ID,
		IPAddress: "192.168.1.10",
		IsActive:  false,
	})

	err := ipPoolService.DeleteIPPool(pool.ID)

	assert.NoError(t, err)

	// Verify pool was deleted
	var count int64
	db.Model(&models.IPPool{}).Where("id = ?", pool.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDeleteIPPool_WithAllocatedIPs(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	// Create a pool
	pool := &models.IPPool{
		Name:     "Test Pool",
		Network:  "192.168.1.0",
		CIDR:     "192.168.1.0/24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.20",
		IsActive: true,
	}
	db.Create(pool)

	// Create an allocated IP
	db.Create(&models.IPAllocation{
		PoolID:    pool.ID,
		IPAddress: "192.168.1.10",
		VMID:      "srv-test123",
		IsActive:  true,
	})

	err := ipPoolService.DeleteIPPool(pool.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete pool with allocated IPs")
}

func TestGetPoolStats_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	// Create a pool
	pool := &models.IPPool{
		Name:     "Test Pool",
		Network:  "192.168.1.0",
		CIDR:     "192.168.1.0/24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.20",
		IsActive: true,
	}
	db.Create(pool)

	// Create IP allocations (total: 5, allocated: 2)
	db.Create(&models.IPAllocation{PoolID: pool.ID, IPAddress: "192.168.1.10", IsActive: false})
	db.Create(&models.IPAllocation{PoolID: pool.ID, IPAddress: "192.168.1.11", IsActive: true, VMID: "vm1"})
	db.Create(&models.IPAllocation{PoolID: pool.ID, IPAddress: "192.168.1.12", IsActive: true, VMID: "vm2"})
	db.Create(&models.IPAllocation{PoolID: pool.ID, IPAddress: "192.168.1.13", IsActive: false})
	db.Create(&models.IPAllocation{PoolID: pool.ID, IPAddress: "192.168.1.14", IsActive: false})

	stats, err := ipPoolService.GetPoolStats(pool.ID)

	assert.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, pool.ID, stats.PoolID)
	assert.Equal(t, pool.Name, stats.PoolName)
	assert.Equal(t, int64(5), stats.Total)
	assert.Equal(t, int64(2), stats.Allocated)
	assert.Equal(t, int64(3), stats.Available)
	assert.Equal(t, 40.0, stats.UsagePercent)
}

func TestGetPoolStats_NotFound(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	stats, err := ipPoolService.GetPoolStats(999)

	assert.Error(t, err)
	assert.Nil(t, stats)
}

func TestGetAllPoolStats_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	// Create multiple pools with allocations
	for i := 1; i <= 3; i++ {
		pool := &models.IPPool{
			Name:     "Pool " + string(rune('0'+i)),
			Network:  "192.168.1.0",
			CIDR:     "192.168.1.0/24",
			Gateway:  "192.168.1.1",
			StartIP:  "192.168.1.10",
			EndIP:    "192.168.1.20",
			IsActive: true,
		}
		db.Create(pool)

		// Add some allocations
		db.Create(&models.IPAllocation{PoolID: pool.ID, IPAddress: "192.168.1.10", IsActive: false})
		db.Create(&models.IPAllocation{PoolID: pool.ID, IPAddress: "192.168.1.11", IsActive: true, VMID: "vm1"})
	}

	stats, err := ipPoolService.GetAllPoolStats()

	assert.NoError(t, err)
	assert.Len(t, stats, 3)
	for _, stat := range stats {
		assert.Equal(t, int64(2), stat.Total)
		assert.Equal(t, int64(1), stat.Allocated)
		assert.Equal(t, int64(1), stat.Available)
	}
}

func TestGenerateIPRange_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	ips, err := ipPoolService.generateIPRange("192.168.1.10", "192.168.1.15")

	assert.NoError(t, err)
	assert.Len(t, ips, 6) // 10, 11, 12, 13, 14, 15
	assert.Equal(t, "192.168.1.10", ips[0])
	assert.Equal(t, "192.168.1.15", ips[5])
}

func TestGenerateIPRange_SingleIP(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	ips, err := ipPoolService.generateIPRange("192.168.1.10", "192.168.1.10")

	assert.NoError(t, err)
	assert.Len(t, ips, 1)
	assert.Equal(t, "192.168.1.10", ips[0])
}

func TestGenerateIPRange_TooLarge(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	// Try to create a range with > 1000 IPs
	ips, err := ipPoolService.generateIPRange("192.168.1.1", "192.168.10.255")

	assert.Error(t, err)
	assert.Nil(t, ips)
	assert.Contains(t, err.Error(), "too large")
}

func TestSuggestIPRange_Success(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	startIP, endIP, err := ipPoolService.SuggestIPRange("192.168.1.0/24")

	assert.NoError(t, err)
	assert.NotEmpty(t, startIP)
	assert.NotEmpty(t, endIP)
	// Should start at .10 (first IP .1 + 9)
	assert.Equal(t, "192.168.1.10", startIP)
}

func TestSuggestIPRange_InvalidCIDR(t *testing.T) {
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := NewIPPoolService(ipPoolRepo)

	startIP, endIP, err := ipPoolService.SuggestIPRange("invalid-cidr")

	assert.Error(t, err)
	assert.Empty(t, startIP)
	assert.Empty(t, endIP)
}

func TestParseCIDR_Success(t *testing.T) {
	network, firstIP, lastIP, broadcast, totalHosts, err := ParseCIDR("192.168.1.0/24")

	assert.NoError(t, err)
	assert.Equal(t, "192.168.1.0", network)
	assert.Equal(t, "192.168.1.1", firstIP)
	assert.Equal(t, "192.168.1.254", lastIP)
	assert.Equal(t, "192.168.1.255", broadcast)
	assert.Equal(t, 254, totalHosts)
}

func TestParseCIDR_InvalidCIDR(t *testing.T) {
	_, _, _, _, _, err := ParseCIDR("invalid")

	assert.Error(t, err)
}

func TestIPConversion(t *testing.T) {
	// Test ipToInt and intToIP conversion
	ip := "192.168.1.100"
	ipInt := ipToInt(net.ParseIP(ip))
	convertedIP := intToIP(ipInt)

	assert.Equal(t, ip, convertedIP)
}
