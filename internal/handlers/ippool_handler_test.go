package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apardo/mikrom-go/internal/models"
	"github.com/apardo/mikrom-go/internal/repository"
	"github.com/apardo/mikrom-go/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateIPPool_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.POST("/ippools", handler.CreateIPPool)

	reqBody := service.CreateIPPoolRequest{
		Name:    "Test Pool",
		Network: "192.168.1.0",
		CIDR:    "192.168.1.0/24",
		Gateway: "192.168.1.1",
		StartIP: "192.168.1.10",
		EndIP:   "192.168.1.20",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/ippools", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.IPPool
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, reqBody.Name, response.Name)
	assert.Equal(t, reqBody.CIDR, response.CIDR)
}

func TestCreateIPPool_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.POST("/ippools", handler.CreateIPPool)

	req := httptest.NewRequest(http.MethodPost, "/ippools", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateIPPool_InvalidCIDR(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.POST("/ippools", handler.CreateIPPool)

	reqBody := service.CreateIPPoolRequest{
		Name:    "Test Pool",
		Network: "192.168.1.0",
		CIDR:    "invalid-cidr",
		Gateway: "192.168.1.1",
		StartIP: "192.168.1.10",
		EndIP:   "192.168.1.20",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/ippools", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListIPPools_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.GET("/ippools", handler.ListIPPools)

	// Create test pools
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
	}

	req := httptest.NewRequest(http.MethodGet, "/ippools?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(3), response["total"])
	assert.NotNil(t, response["items"])
}

func TestListIPPools_WithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.GET("/ippools", handler.ListIPPools)

	// Create 10 test pools
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

	req := httptest.NewRequest(http.MethodGet, "/ippools?page=1&page_size=3", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), response["total"])
	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(3), response["page_size"])
}

func TestGetIPPool_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.GET("/ippools/:id", handler.GetIPPool)

	// Create test pool
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

	req := httptest.NewRequest(http.MethodGet, "/ippools/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.IPPool
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, pool.Name, response.Name)
}

func TestGetIPPool_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.GET("/ippools/:id", handler.GetIPPool)

	req := httptest.NewRequest(http.MethodGet, "/ippools/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetIPPool_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.GET("/ippools/:id", handler.GetIPPool)

	req := httptest.NewRequest(http.MethodGet, "/ippools/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateIPPool_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.PATCH("/ippools/:id", handler.UpdateIPPool)

	// Create test pool
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

	newName := "Updated Name"
	reqBody := service.UpdateIPPoolRequest{
		Name: &newName,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPatch, "/ippools/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.IPPool
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, newName, response.Name)
}

func TestUpdateIPPool_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.PATCH("/ippools/:id", handler.UpdateIPPool)

	newName := "Updated Name"
	reqBody := service.UpdateIPPoolRequest{Name: &newName}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPatch, "/ippools/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteIPPool_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.DELETE("/ippools/:id", handler.DeleteIPPool)

	// Create test pool
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

	req := httptest.NewRequest(http.MethodDelete, "/ippools/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteIPPool_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.DELETE("/ippools/:id", handler.DeleteIPPool)

	req := httptest.NewRequest(http.MethodDelete, "/ippools/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPoolStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.GET("/ippools/:id/stats", handler.GetPoolStats)

	// Create test pool
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

	// Create IP allocations
	db.Create(&models.IPAllocation{PoolID: pool.ID, IPAddress: "192.168.1.10", IsActive: false})
	db.Create(&models.IPAllocation{PoolID: pool.ID, IPAddress: "192.168.1.11", IsActive: true, VMID: "vm1"})

	req := httptest.NewRequest(http.MethodGet, "/ippools/1/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response service.IPPoolStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, pool.ID, response.PoolID)
	assert.Equal(t, int64(2), response.Total)
	assert.Equal(t, int64(1), response.Allocated)
}

func TestGetPoolStats_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.GET("/ippools/:id/stats", handler.GetPoolStats)

	req := httptest.NewRequest(http.MethodGet, "/ippools/999/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetAllPoolStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.GET("/ippools/stats", handler.GetAllPoolStats)

	// Create test pools
	for i := 1; i <= 2; i++ {
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
		db.Create(&models.IPAllocation{PoolID: pool.ID, IPAddress: "192.168.1.10", IsActive: false})
	}

	req := httptest.NewRequest(http.MethodGet, "/ippools/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["pools"])
}

func TestSuggestIPRange_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.POST("/ippools/suggest-range", handler.SuggestIPRange)

	reqBody := map[string]string{
		"cidr": "192.168.1.0/24",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/ippools/suggest-range", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "192.168.1.0/24", response["cidr"])
	assert.NotEmpty(t, response["suggested_start"])
	assert.NotEmpty(t, response["suggested_end"])
}

func TestSuggestIPRange_InvalidCIDR(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.POST("/ippools/suggest-range", handler.SuggestIPRange)

	reqBody := map[string]string{
		"cidr": "invalid-cidr",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/ippools/suggest-range", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSuggestIPRange_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.POST("/ippools/suggest-range", handler.SuggestIPRange)

	req := httptest.NewRequest(http.MethodPost, "/ippools/suggest-range", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Additional edge case tests

func TestListIPPools_InvalidPageNumber(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.GET("/ippools", handler.ListIPPools)

	req := httptest.NewRequest(http.MethodGet, "/ippools?page=invalid&page_size=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should still return OK with default pagination
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateIPPool_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.PATCH("/ippools/:id", handler.UpdateIPPool)

	newName := "Updated Name"
	reqBody := service.UpdateIPPoolRequest{Name: &newName}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPatch, "/ippools/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// May return 404 or 500 depending on implementation
	assert.Contains(t, []int{http.StatusNotFound, http.StatusInternalServerError}, w.Code)
}

func TestUpdateIPPool_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.PATCH("/ippools/:id", handler.UpdateIPPool)

	req := httptest.NewRequest(http.MethodPatch, "/ippools/1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteIPPool_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.DELETE("/ippools/:id", handler.DeleteIPPool)

	req := httptest.NewRequest(http.MethodDelete, "/ippools/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Delete may be idempotent and return 204 even if not found
	assert.Contains(t, []int{http.StatusNoContent, http.StatusNotFound, http.StatusInternalServerError}, w.Code)
}

func TestGetPoolStats_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := repository.SetupTestDB(t)
	defer repository.CleanupTestDB(t, db)

	ipPoolRepo := repository.NewIPPoolRepository(db)
	ipPoolService := service.NewIPPoolService(ipPoolRepo)
	handler := NewIPPoolHandler(ipPoolService)
	router := gin.New()
	router.GET("/ippools/:id/stats", handler.GetPoolStats)

	req := httptest.NewRequest(http.MethodGet, "/ippools/invalid/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
