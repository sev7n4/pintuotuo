#!/bin/bash
# Update all seed users password hash with production JWT_SECRET
# This script should be run after deployment to fix legacy seed users

set -e

JWT_SECRET="pintuotuo-jwt-secret-key-2024-production-32"
DB_HOST="localhost"
DB_USER="pintuotuo"
DB_NAME="pintuotuo_db"

# Common passwords for seed users (password123 pattern)
# We need to update all users that were created with the old dev JWT_SECRET

# Update password for user123 (common password)
PASSWORD="user123"
HASH=$(echo -n "$PASSWORD$JWT_SECRET" | sha256sum | cut -d' ' -f1)
echo "Updating users with password 'user123'..."
docker exec pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db -c "UPDATE users SET password_hash = '$HASH' WHERE email IN ('user01@163.com', 'user02@163.com', 'user06@163.com', 'user10@163.com', 'user11@163.com', 'user12@163.com', 'user13@163.com', 'user14@163.com', 'user15@163.com', 'user16@163.com', 'user17@163.com', 'user18@163.com', 'user19@163.com', 'user20@163.com');"

# Update password for demo123456
PASSWORD="demo123456"
HASH=$(echo -n "$PASSWORD$JWT_SECRET" | sha256sum | cut -d' ' -f1)
echo "Updating users with password 'demo123456'..."
docker exec pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db -c "UPDATE users SET password_hash = '$HASH' WHERE email IN ('demo@example.com');"

# Update password for merchant123456
PASSWORD="merchant123456"
HASH=$(echo -n "$PASSWORD$JWT_SECRET" | sha256sum | cut -d' ' -f1)
echo "Updating users with password 'merchant123456'..."
docker exec pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db -c "UPDATE users SET password_hash = '$HASH' WHERE email LIKE 'merchant%';"

# Update password for admin123456
PASSWORD="admin123456"
HASH=$(echo -n "$PASSWORD$JWT_SECRET" | sha256sum | cut -d' ' -f1)
echo "Updating users with password 'admin123456'..."
docker exec pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db -c "UPDATE users SET password_hash = '$HASH' WHERE email IN ('admin@example.com', 'admin2@example.com');"

echo "Done! Verifying updates..."
docker exec pintuotuo-postgres psql -U pintuotuo -d pintuotuo_db -c "SELECT email, LEFT(password_hash, 10) as hash_preview FROM users WHERE email LIKE '%@%' ORDER BY id LIMIT 10;"
