-- ============================================
-- E2E Test Data Seed Script
-- This script creates test accounts for E2E testing
-- ============================================

-- Password hash algorithm: SHA256(password + JWT_SECRET)
-- Default JWT_SECRET: "pintuotuo-secret-key-dev"

-- Regular user: demo@example.com / demo123456
INSERT INTO users (email, name, password_hash, role, status)
VALUES (
    'demo@example.com',
    'Demo User',
    'd5e40f73f3eb863d24cff64aa15877de65a52fcea56d29db3585b90988a51311',
    'user',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    status = EXCLUDED.status;

-- Merchant user: merchant@example.com / merchant123456
INSERT INTO users (email, name, password_hash, role, status)
VALUES (
    'merchant@example.com',
    'Test Merchant',
    'f43af1330d72420b14e89d803373cc3f2db79b3d293aec9257f227b10588e5c9',
    'merchant',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    status = EXCLUDED.status;

-- Admin user: admin@example.com / admin123456
INSERT INTO users (email, name, password_hash, role, status)
VALUES (
    'admin@example.com',
    'Test Admin',
    'e1e80b6b77cc437cdfd1d183a7831b0a6b7c35c4b4d6a1c1d4b6932de9e75cc6',
    'admin',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    status = EXCLUDED.status;

-- Create token records for test users
INSERT INTO tokens (user_id, balance, total_used, total_earned)
SELECT id, 100.00, 0, 100.00 FROM users WHERE email = 'demo@example.com'
ON CONFLICT (user_id) DO UPDATE SET balance = 100.00;

INSERT INTO tokens (user_id, balance, total_used, total_earned)
SELECT id, 500.00, 0, 500.00 FROM users WHERE email = 'merchant@example.com'
ON CONFLICT (user_id) DO UPDATE SET balance = 500.00;

INSERT INTO tokens (user_id, balance, total_used, total_earned)
SELECT id, 1000.00, 0, 1000.00 FROM users WHERE email = 'admin@example.com'
ON CONFLICT (user_id) DO UPDATE SET balance = 1000.00;

-- Verify test data
SELECT id, email, name, role, status FROM users 
WHERE email IN ('demo@example.com', 'merchant@example.com', 'admin@example.com');
