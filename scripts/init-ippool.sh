#!/usr/bin/env bash
set -e

# IP Pool Initialization Script
# This script initializes the IP pools in the database

echo "🚀 Initializing IP Pools..."

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-mikrom}"

export PGPASSWORD="$DB_PASSWORD"

# Function to execute SQL
execute_sql() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$1"
}

# Check if default pool already exists
POOL_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM ip_pools WHERE name='default';")

if [ "$POOL_EXISTS" -gt 0 ]; then
    echo "⚠️  Default IP pool already exists. Skipping initialization."
    exit 0
fi

echo "📦 Creating default IP pool (10.100.0.0/24)..."

# Create default IP pool
execute_sql "
INSERT INTO ip_pools (name, network, cidr, gateway, start_ip, end_ip, is_active, created_at, updated_at)
VALUES (
    'default',
    '10.100.0.0',
    '10.100.0.0/24',
    '10.100.0.1',
    '10.100.0.10',
    '10.100.0.254',
    true,
    NOW(),
    NOW()
);
"

# Get the pool ID
POOL_ID=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT id FROM ip_pools WHERE name='default';")

echo "✅ IP pool created with ID: $POOL_ID"

echo "📝 Generating IP allocations (10.100.0.10 - 10.100.0.254)..."

# Generate IP allocations
execute_sql "
INSERT INTO ip_allocations (pool_id, ip_address, vm_id, is_active, allocated_at)
SELECT 
    $POOL_ID,
    '10.100.0.' || i,
    '',
    false,
    NOW()
FROM generate_series(10, 254) AS i;
"

# Count allocations
ALLOCATION_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM ip_allocations WHERE pool_id=$POOL_ID;")

echo "✅ Created $ALLOCATION_COUNT IP allocations"

echo ""
echo "🎉 IP Pool initialization complete!"
echo ""
echo "Pool Details:"
echo "  Name:     default"
echo "  Network:  10.100.0.0/24"
echo "  Gateway:  10.100.0.1"
echo "  Range:    10.100.0.10 - 10.100.0.254"
echo "  Total IPs: $ALLOCATION_COUNT"
echo ""
echo "You can now create VMs and they will automatically get IPs from this pool."
