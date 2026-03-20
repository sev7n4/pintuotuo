-- ============================================
-- 创建管理员账户 SQL 脚本
-- 数据库: pintuotuo_db
-- 用户: pintuotuo
-- ============================================

-- 注意：密码哈希算法为 SHA256(password + JWT_SECRET)
-- 默认 JWT_SECRET 为 "pintuotuo-secret-key-dev"
-- 如果您修改了 JWT_SECRET，需要重新计算密码哈希

-- 方式1：创建新的管理员账户
-- 邮箱: admin@pintuotuo.com
-- 密码: Admin@123 (请登录后立即修改)
-- 密码哈希: SHA256("Admin@123" + "pintuotuo-secret-key-dev")

INSERT INTO users (email, name, password_hash, role, status)
VALUES (
    'admin@pintuotuo.com',
    '系统管理员',
    'c007705a6b4d725b90b1564975be5bd9ac2cfa6408114eace6279f03511f2bba',
    'admin',
    'active'
) ON CONFLICT (email) DO UPDATE SET
    role = 'admin',
    status = 'active';

-- 为管理员创建 token 余额记录
INSERT INTO tokens (user_id, balance, total_used, total_earned)
SELECT id, 0, 0, 0 FROM users WHERE email = 'admin@pintuotuo.com'
ON CONFLICT (user_id) DO NOTHING;

-- 查询确认
SELECT id, email, name, role, status, created_at FROM users WHERE email = 'admin@pintuotuo.com';

-- ============================================
-- 方式2：将现有用户升级为管理员
-- ============================================

-- 取消下面的注释并修改邮箱为您自己的账户
-- UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';

-- ============================================
-- 生成密码哈希的方法
-- ============================================

-- 如果您的 JWT_SECRET 不是默认值，需要在服务器上运行以下命令生成密码哈希：
-- echo -n "您的密码您的JWT_SECRET" | shasum -a 256

-- 或者在 Go 代码中：
-- hash := sha256.Sum256([]byte("您的密码" + "您的JWT_SECRET"))
-- fmt.Sprintf("%x", hash)
