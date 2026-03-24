#!/bin/bash
# Update seed users password hash with production JWT_SECRET
# Only the 3 seed users from seed_test_data.sql need to be fixed

set -e

JWT_SECRET="pintuotuo-jwt-secret-key-2024-production-32"

# Function to generate password hash
generate_hash() {
    local password=$1
    echo -n "$password$JWT_SECRET" | sha256sum | cut -d' ' -f1
}

# Update seed users (these are the only users injected by migrations)

# admin@example.com / admin123456
HASH=$(generate_hash "admin123456")
echo "Updating admin@example.com..."
docker exec pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db -c "UPDATE users SET password_hash = '$HASH' WHERE email = 'admin@example.com';"

# demo@example.com / demo123456
HASH=$(generate_hash "demo123456")
echo "Updating demo@example.com..."
docker exec pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db -c "UPDATE users SET password_hash = '$HASH' WHERE email = 'demo@example.com';"

# merchant@example.com / merchant123456
HASH=$(generate_hash "merchant123456")
echo "Updating merchant@example.com..."
docker exec pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db -c "UPDATE users SET password_hash = '$HASH' WHERE email = 'merchant@example.com';"

echo "Done! Verifying updates..."
docker exec pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db -c "SELECT email, LEFT(password_hash, 10) as hash_preview FROM users WHERE email IN ('admin@example.com', 'demo@example.com', 'merchant@example.com');"
