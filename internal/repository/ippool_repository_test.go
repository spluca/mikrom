package repository

import (
	"testing"

	"github.com/spluca/mikrom/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPPoolRepository_CreatePool_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewIPPoolRepository(db)

	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}

	err := repo.CreatePool(pool)

	assert.NoError(t, err)
	assert.NotZero(t, pool.ID)
	assert.NotZero(t, pool.CreatedAt)
}

func TestIPPoolRepository_CreatePool_Error(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewIPPoolRepository(db)

	// Close DB to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	pool := &models.IPPool{Name: "test-pool"}
	err := repo.CreatePool(pool)

	assert.Error(t, err)
}

func TestIPPoolRepository_FindPoolByID_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create a pool
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	repo := NewIPPoolRepository(db)
	found, err := repo.FindPoolByID(uint(pool.ID))

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, pool.Name, found.Name)
	assert.Equal(t, pool.Network, found.Network)
}

func TestIPPoolRepository_FindPoolByID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewIPPoolRepository(db)
	found, err := repo.FindPoolByID(999)

	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestIPPoolRepository_FindActivePool_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create first pool (will be returned first due to ID order)
	firstPool := &models.IPPool{
		Name:     "first-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(firstPool)

	// Create second active pool
	db.Create(&models.IPPool{
		Name:     "second-pool",
		Network:  "192.168.2.0/24",
		CIDR:     "24",
		Gateway:  "192.168.2.1",
		StartIP:  "192.168.2.10",
		EndIP:    "192.168.2.100",
		IsActive: true,
	})

	repo := NewIPPoolRepository(db)
	found, err := repo.FindActivePool()

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.True(t, found.IsActive)
	// Should return the first one (by ID)
	assert.Equal(t, firstPool.Name, found.Name)
}

func TestIPPoolRepository_FindActivePool_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Don't create any pools
	repo := NewIPPoolRepository(db)
	found, err := repo.FindActivePool()

	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestIPPoolRepository_ListPools_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create multiple pools
	for i := 1; i <= 5; i++ {
		db.Create(&models.IPPool{
			Name:     "pool-" + string(rune(i)),
			Network:  "192.168.1.0/24",
			CIDR:     "24",
			Gateway:  "192.168.1.1",
			StartIP:  "192.168.1.10",
			EndIP:    "192.168.1.100",
			IsActive: true,
		})
	}

	repo := NewIPPoolRepository(db)
	pools, total, err := repo.ListPools(0, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, pools, 5)
}

func TestIPPoolRepository_ListPools_Pagination(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create 10 pools
	for i := 1; i <= 10; i++ {
		db.Create(&models.IPPool{
			Name:     "pool-" + string(rune(i)),
			Network:  "192.168.1.0/24",
			CIDR:     "24",
			Gateway:  "192.168.1.1",
			StartIP:  "192.168.1.10",
			EndIP:    "192.168.1.100",
			IsActive: true,
		})
	}

	repo := NewIPPoolRepository(db)

	// First page
	pools, total, err := repo.ListPools(0, 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, pools, 5)

	// Second page
	pools, total, err = repo.ListPools(5, 5)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, pools, 5)
}

func TestIPPoolRepository_UpdatePool_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool
	pool := &models.IPPool{
		Name:     "original-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	repo := NewIPPoolRepository(db)

	// Update pool
	pool.Name = "updated-pool"
	pool.IsActive = false

	err := repo.UpdatePool(pool)

	assert.NoError(t, err)

	// Verify update
	var found models.IPPool
	db.First(&found, pool.ID)
	assert.Equal(t, "updated-pool", found.Name)
	assert.False(t, found.IsActive)
}

func TestIPPoolRepository_UpdatePool_Error(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewIPPoolRepository(db)

	// Close DB to simulate error
	sqlDB, _ := db.DB()
	sqlDB.Close()

	pool := &models.IPPool{ID: 1, Name: "test"}
	err := repo.UpdatePool(pool)

	assert.Error(t, err)
}

func TestIPPoolRepository_DeletePool_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	repo := NewIPPoolRepository(db)
	err := repo.DeletePool(uint(pool.ID))

	assert.NoError(t, err)

	// Verify deletion
	var found models.IPPool
	result := db.First(&found, pool.ID)
	assert.Error(t, result.Error)
}

func TestIPPoolRepository_AllocateIP_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	// Create available IP allocations
	allocations := []models.IPAllocation{
		{PoolID: pool.ID, IPAddress: "192.168.1.10", VMID: "", IsActive: false},
		{PoolID: pool.ID, IPAddress: "192.168.1.11", VMID: "", IsActive: false},
		{PoolID: pool.ID, IPAddress: "192.168.1.12", VMID: "", IsActive: false},
	}
	db.Create(&allocations)

	repo := NewIPPoolRepository(db)
	allocation, err := repo.AllocateIP(pool.ID, "srv-test123")

	assert.NoError(t, err)
	require.NotNil(t, allocation)
	assert.Equal(t, "srv-test123", allocation.VMID)
	assert.True(t, allocation.IsActive)
	assert.NotEmpty(t, allocation.IPAddress)
}

func TestIPPoolRepository_AllocateIP_NoAvailableIPs(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	// Create only allocated IPs
	allocations := []models.IPAllocation{
		{PoolID: pool.ID, IPAddress: "192.168.1.10", VMID: "srv-existing1", IsActive: true},
		{PoolID: pool.ID, IPAddress: "192.168.1.11", VMID: "srv-existing2", IsActive: true},
	}
	db.Create(&allocations)

	repo := NewIPPoolRepository(db)
	allocation, err := repo.AllocateIP(pool.ID, "srv-test123")

	assert.Error(t, err)
	assert.Nil(t, allocation)
	assert.Contains(t, err.Error(), "no available IPs")
}

func TestIPPoolRepository_AllocateIP_InactivePool(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool (will be active by default)
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	// Now update it to inactive
	pool.IsActive = false
	db.Save(pool)

	// Verify pool is now inactive
	var checkPool models.IPPool
	db.First(&checkPool, pool.ID)
	require.False(t, checkPool.IsActive, "Pool should be inactive")

	// Create available IP
	db.Create(&models.IPAllocation{
		PoolID:    pool.ID,
		IPAddress: "192.168.1.10",
		VMID:      "",
		IsActive:  false,
	})

	repo := NewIPPoolRepository(db)
	allocation, err := repo.AllocateIP(pool.ID, "srv-test123")

	assert.Error(t, err)
	assert.Nil(t, allocation)
	assert.Contains(t, err.Error(), "not active")
}

func TestIPPoolRepository_ReleaseIP_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	// Create allocated IP
	allocation := &models.IPAllocation{
		PoolID:    pool.ID,
		IPAddress: "192.168.1.10",
		VMID:      "srv-test123",
		IsActive:  true,
	}
	db.Create(allocation)

	repo := NewIPPoolRepository(db)
	err := repo.ReleaseIP("192.168.1.10")

	assert.NoError(t, err)

	// Verify release
	var found models.IPAllocation
	db.Where("ip_address = ?", "192.168.1.10").First(&found)
	assert.Empty(t, found.VMID)
	assert.False(t, found.IsActive)
}

func TestIPPoolRepository_ReleaseIP_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewIPPoolRepository(db)
	err := repo.ReleaseIP("192.168.99.99")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIPPoolRepository_FindAllocationByVMID_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool and allocation
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	allocation := &models.IPAllocation{
		PoolID:    pool.ID,
		IPAddress: "192.168.1.10",
		VMID:      "srv-test123",
		IsActive:  true,
	}
	db.Create(allocation)

	repo := NewIPPoolRepository(db)
	found, err := repo.FindAllocationByVMID("srv-test123")

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "192.168.1.10", found.IPAddress)
	assert.Equal(t, "srv-test123", found.VMID)
}

func TestIPPoolRepository_FindAllocationByVMID_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewIPPoolRepository(db)
	found, err := repo.FindAllocationByVMID("srv-nonexistent")

	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestIPPoolRepository_FindAllocationByIP_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool and allocation
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	allocation := &models.IPAllocation{
		PoolID:    pool.ID,
		IPAddress: "192.168.1.10",
		VMID:      "srv-test123",
		IsActive:  true,
	}
	db.Create(allocation)

	repo := NewIPPoolRepository(db)
	found, err := repo.FindAllocationByIP("192.168.1.10")

	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "192.168.1.10", found.IPAddress)
	assert.Equal(t, "srv-test123", found.VMID)
}

func TestIPPoolRepository_FindAllocationByIP_NotFound(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := NewIPPoolRepository(db)
	found, err := repo.FindAllocationByIP("192.168.99.99")

	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestIPPoolRepository_CreateAllocations_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	// Create multiple allocations
	allocations := []models.IPAllocation{
		{PoolID: pool.ID, IPAddress: "192.168.1.10", VMID: "", IsActive: false},
		{PoolID: pool.ID, IPAddress: "192.168.1.11", VMID: "", IsActive: false},
		{PoolID: pool.ID, IPAddress: "192.168.1.12", VMID: "", IsActive: false},
	}

	repo := NewIPPoolRepository(db)
	err := repo.CreateAllocations(allocations)

	assert.NoError(t, err)

	// Verify all were created
	var count int64
	db.Model(&models.IPAllocation{}).Where("pool_id = ?", pool.ID).Count(&count)
	assert.Equal(t, int64(3), count)
}

func TestIPPoolRepository_GetPoolStats_Success(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	// Create allocations (3 allocated, 2 available)
	allocations := []models.IPAllocation{
		{PoolID: pool.ID, IPAddress: "192.168.1.10", VMID: "srv-test1", IsActive: true},
		{PoolID: pool.ID, IPAddress: "192.168.1.11", VMID: "srv-test2", IsActive: true},
		{PoolID: pool.ID, IPAddress: "192.168.1.12", VMID: "srv-test3", IsActive: true},
		{PoolID: pool.ID, IPAddress: "192.168.1.13", VMID: "", IsActive: false},
		{PoolID: pool.ID, IPAddress: "192.168.1.14", VMID: "", IsActive: false},
	}
	db.Create(&allocations)

	repo := NewIPPoolRepository(db)
	total, allocated, available, err := repo.GetPoolStats(pool.ID)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, int64(3), allocated)
	assert.Equal(t, int64(2), available)
}

func TestIPPoolRepository_GetPoolStats_EmptyPool(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	// Create pool with no allocations
	pool := &models.IPPool{
		Name:     "test-pool",
		Network:  "192.168.1.0/24",
		CIDR:     "24",
		Gateway:  "192.168.1.1",
		StartIP:  "192.168.1.10",
		EndIP:    "192.168.1.100",
		IsActive: true,
	}
	db.Create(pool)

	repo := NewIPPoolRepository(db)
	total, allocated, available, err := repo.GetPoolStats(pool.ID)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, int64(0), allocated)
	assert.Equal(t, int64(0), available)
}
