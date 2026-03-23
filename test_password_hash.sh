#!/bin/bash

# 测试密码哈希生成逻辑
echo "Testing password hash generation logic..."

# 模拟从docker-compose配置中获取JWT_SECRET
JWT_SECRET="test-secret-production"
echo "JWT_SECRET: $JWT_SECRET"

# 测试密码
PASSWORD="admin123456"
echo "Password: $PASSWORD"

# 生成密码哈希
if command -v sha256sum >/dev/null 2>&1; then
  PASSWORD_HASH=$(echo -n "$PASSWORD$JWT_SECRET" | sha256sum | awk '{print $1}')
  echo "Using sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  PASSWORD_HASH=$(echo -n "$PASSWORD$JWT_SECRET" | shasum -a 256 | awk '{print $1}')
  echo "Using shasum"
else
  echo "Error: No sha256sum or shasum command found"
  exit 1
fi

echo "Generated password hash: $PASSWORD_HASH"

echo "\nTesting backend password verification..."

# 模拟后端验证逻辑
verify_password() {
  local password="$1"
  local hash="$2"
  local secret="$3"
  
  if command -v sha256sum >/dev/null 2>&1; then
    local calculated_hash=$(echo -n "${password}${secret}" | sha256sum | awk '{print $1}')
  else
    local calculated_hash=$(echo -n "${password}${secret}" | shasum -a 256 | awk '{print $1}')
  fi
  
  if [ "$calculated_hash" = "$hash" ]; then
    echo "✓ Password verification successful"
    return 0
  else
    echo "✗ Password verification failed"
    return 1
  fi
}

# 测试验证逻辑
verify_password "$PASSWORD" "$PASSWORD_HASH" "$JWT_SECRET"
verify_password "wrongpassword" "$PASSWORD_HASH" "$JWT_SECRET"

# 测试不同JWT_SECRET的情况
echo "\nTesting with different JWT_SECRET..."
NEW_JWT_SECRET="different-secret"
if command -v sha256sum >/dev/null 2>&1; then
  NEW_PASSWORD_HASH=$(echo -n "$PASSWORD$NEW_JWT_SECRET" | sha256sum | awk "{print $1}")
else
  NEW_PASSWORD_HASH=$(echo -n "$PASSWORD$NEW_JWT_SECRET" | shasum -a 256 | awk "{print $1}")
fi
echo "New JWT_SECRET: $NEW_JWT_SECRET"
echo "New password hash: $NEW_PASSWORD_HASH"

# 验证不同JWT_SECRET生成的哈希值不同
if [ "$PASSWORD_HASH" != "$NEW_PASSWORD_HASH" ]; then
  echo "✓ Different JWT_SECRET generates different hash"
else
  echo "✗ Different JWT_SECRET generates same hash (ERROR)"
fi

echo "\nTest completed successfully!"
