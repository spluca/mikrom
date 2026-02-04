package main

import (
	"fmt"
	"log"

	"github.com/spluca/mikrom/config"
	"github.com/spluca/mikrom/internal/models"
	"github.com/spluca/mikrom/pkg/database"
)

func main() {
	fmt.Println("🚀 Initializing IP Pools...")

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db, err := database.NewDatabase(cfg.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check if default pool already exists
	var count int64
	if err := db.DB.Model(&models.IPPool{}).Where("name = ?", "default").Count(&count).Error; err != nil {
		log.Fatalf("Failed to check existing pools: %v", err)
	}

	if count > 0 {
		fmt.Println("⚠️  Default IP pool already exists. Skipping initialization.")
		return
	}

	fmt.Println("📦 Creating default IP pool (10.100.0.0/24)...")

	// Create default IP pool
	pool := &models.IPPool{
		Name:     "default",
		Network:  "10.100.0.0",
		CIDR:     "10.100.0.0/24",
		Gateway:  "10.100.0.1",
		StartIP:  "10.100.0.10",
		EndIP:    "10.100.0.254",
		IsActive: true,
	}

	if err := db.DB.Create(pool).Error; err != nil {
		log.Fatalf("Failed to create IP pool: %v", err)
	}

	fmt.Printf("✅ IP pool created with ID: %d\n", pool.ID)

	fmt.Println("📝 Generating IP allocations (10.100.0.10 - 10.100.0.254)...")

	// Generate IP allocations
	allocations := make([]models.IPAllocation, 0, 245)
	for i := 10; i <= 254; i++ {
		allocations = append(allocations, models.IPAllocation{
			PoolID:    pool.ID,
			IPAddress: fmt.Sprintf("10.100.0.%d", i),
			VMID:      "",
			IsActive:  false,
		})
	}

	// Batch insert allocations
	if err := db.DB.CreateInBatches(allocations, 100).Error; err != nil {
		log.Fatalf("Failed to create IP allocations: %v", err)
	}

	fmt.Printf("✅ Created %d IP allocations\n", len(allocations))

	fmt.Println("")
	fmt.Println("🎉 IP Pool initialization complete!")
	fmt.Println("")
	fmt.Println("Pool Details:")
	fmt.Println("  Name:     default")
	fmt.Println("  Network:  10.100.0.0/24")
	fmt.Println("  Gateway:  10.100.0.1")
	fmt.Println("  Range:    10.100.0.10 - 10.100.0.254")
	fmt.Printf("  Total IPs: %d\n", len(allocations))
	fmt.Println("")
	fmt.Println("You can now create VMs and they will automatically get IPs from this pool.")
}
