#!/bin/bash
set -e  # Exit on any error

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up test environment...${NC}"

# Function to cleanup on exit
cleanup() {
    echo -e "\n${GREEN}Cleaning up test environment...${NC}"
    docker-compose down -v
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}Cleanup failed with exit code $exit_code${NC}"
        exit $exit_code
    fi
}

# Register the cleanup function to run on script exit
trap cleanup EXIT

# Start PostgreSQL container
echo "Starting PostgreSQL container..."
docker-compose up -d postgres

# Create .env.test if it doesn't exist
if [ ! -f .env.test ]; then
    echo "Creating .env.test file..."
    cat > .env.test << EOL
# Server Configuration
PORT=8081

# Database Configuration
DATABASE_URL=postgres://postgres:postgres@localhost:5432/golinks_test?sslmode=disable

# Authentication
JWT_SECRET=test-secret-key-do-not-use-in-production
ENABLE_OKTA_SSO=false

# Optional: Okta Configuration (disabled for tests)
# OKTA_ORG_URL=
# OKTA_CLIENT_ID=
# OKTA_CLIENT_SECRET=
EOL
fi

# Function to check if postgres is ready
wait_for_postgres() {
    local retries=60  # Increase timeout to 60 seconds
    local counter=0
    echo "Waiting for PostgreSQL to be ready..."
    until docker exec golinks-db pg_isready -U postgres -h localhost > /dev/null 2>&1; do
        counter=$((counter + 1))
        if [ $counter -ge $retries ]; then
            echo -e "${RED}Timed out waiting for PostgreSQL to be ready${NC}"
            return 1
        fi
        echo "Waiting... ($counter/$retries)"
        sleep 1
    done
    # Additional wait to ensure PostgreSQL is fully ready
    sleep 5
    return 0
}

# Wait for PostgreSQL to be ready
if ! wait_for_postgres; then
    echo -e "${RED}Failed to connect to PostgreSQL${NC}"
    exit 1
fi

# Create test database
echo "Creating test database..."
if ! docker exec -i golinks-db psql -U postgres -c "DROP DATABASE IF EXISTS golinks_test;"; then
    echo -e "${RED}Failed to drop test database${NC}"
    exit 1
fi

if ! docker exec -i golinks-db psql -U postgres -c "CREATE DATABASE golinks_test;"; then
    echo -e "${RED}Failed to create test database${NC}"
    exit 1
fi

# Load schema into test database
echo "Loading schema..."
if ! docker exec -i golinks-db psql -U postgres -d golinks_test < scripts/schema.sql; then
    echo -e "${RED}Failed to load schema${NC}"
    exit 1
fi

# Run the tests
echo -e "\n${GREEN}Running tests...${NC}"
go test ./tests/integration -v

# Exit code will trigger cleanup function 