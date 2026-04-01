# 腾讯云服务器部署配置指南

## 目标

让腾讯云服务器能够直接从 GitHub 拉取代码，实现自动化部署（CI/CD）。

---

## 前置条件

- 腾讯云服务器已安装 Docker 和 Docker Compose
- 本地已配置 SSH 访问服务器的密钥 (`~/.ssh/tencent_cloud_deploy`)
- GitHub 仓库管理员权限

---

## 第一步：生成 SSH 密钥对

在腾讯云服务器上生成用于访问 GitHub 的 SSH 密钥：

```bash
# SSH 连接到服务器
ssh -i ~/.ssh/tencent_cloud_deploy root@119.29.173.89

# 生成 SSH 密钥对（ED25519 算法，更安全）
ssh-keygen -t ed25519 -C 'tencent-cloud-deploy@pintuotuo' -f /root/.ssh/github_deploy -N ''

# 查看公钥（下一步需要用到）
cat /root/.ssh/github_deploy.pub
```

**生成的公钥**：
```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGMjIrkUCK+DUpeX46PsP20Lb/niTrySmnqyA2Nv/QM+ tencent-cloud-deploy@pintuotuo
```

---

## 第二步：在 GitHub 添加 SSH 公钥

### 2.1 复制公钥

从上一步获取公钥，格式类似：
```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGMjIrkUCK+DUpeX46PsP20Lb/niTrySmnqyA2Nv/QM+ tencent-cloud-deploy@pintuotuo
```

### 2.2 添加到 GitHub

1. 打开 GitHub → 点击右上角头像 → **Settings**
2. 左侧菜单找到 **SSH and GPG keys**
3. 点击 **New SSH key**
4. 填写：
   - **Title**: `Tencent Cloud Deploy`
   - **Key type**: `Authentication Key`
   - **Key**: 粘贴上面的公钥内容
5. 点击 **Add SSH key**

---

## 第三步：验证 SSH 连接

在本地终端执行验证：

```bash
ssh -i ~/.ssh/tencent_cloud_deploy root@119.29.173.89 "ssh -T git@github.com"
```

**成功输出**：
```
Hi sev7n4! You've successfully authenticated, but GitHub does not provide shell access.
```

如果看到 `Permission denied`，请检查：
- 公钥是否正确添加到 GitHub
- GitHub 是否是同一个账户

---

## 第四步：配置 Git 仓库

在服务器上配置 Git 克隆仓库：

```bash
ssh -i ~/.ssh/tencent_cloud_deploy root@119.29.173.89 << 'EOF'
# 配置 Git 用户信息
git config --global user.name "CI/CD Bot"
git config --global user.email "ci-bot@pintuotuo"

# 如果仓库已存在，配置远程仓库 URL
cd /opt/pintuotuo
git remote set-url origin git@github.com:sev7n4/pintuotuo.git

# 验证配置
git remote -v
EOF
```

---

## 第五步：配置 SSH 配置文件（可选但推荐）

在服务器上创建 SSH 配置文件，优化 GitHub 连接：

```bash
ssh -i ~/.ssh/tencent_cloud_deploy root@119.29.173.89 << 'EOF'
cat > /root/.ssh/config << 'CONFIG'
Host github.com
    HostName github.com
    User git
    IdentityFile /root/.ssh/github_deploy
    StrictHostKeyChecking no
    ServerAliveInterval 60
CONFIG

chmod 600 /root/.ssh/config
cat /root/.ssh/config
EOF
```

---

## 部署流程说明

### 新的 CI/CD 流程

```
代码提交到 main 分支
        │
        ▼
GitHub Actions 触发
        │
        ▼
SSH 连接到服务器
        │
        ▼
服务器执行: git pull origin main
        │
        ▼
服务器执行: docker-compose up -d --build
        │
        ▼
完成！
```

### 优势

1. **更稳定**: 不依赖 GitHub Actions runner 传输文件
2. **更快**: 服务器直接拉取增量更新
3. **更简单**: 流程更清晰，易于调试
4. **节省资源**: 不占用 runner 带宽

---

## 测试充值按钮开关（仅测试环境）

若需要在「我的 Token」页面显示“模拟支付完成”按钮，请同时开启后端与前端两个开关，然后重建部署。

### 1) 服务器 `.env` 增加（或更新）

```bash
ALLOW_TEST_RECHARGE=true
VITE_ALLOW_MOCK_RECHARGE=true
```

说明：
- `ALLOW_TEST_RECHARGE`：后端允许 `POST /api/v1/tokens/recharge/orders/:id/mock-pay`
- `VITE_ALLOW_MOCK_RECHARGE`：前端构建时决定是否显示按钮（编译期变量）

### 2) 重建并重启

```bash
cd /opt/pintuotuo
docker-compose -f docker-compose.prod.yml up -d --build --force-recreate
```

### 3) 关闭按钮（恢复默认）

将 `.env` 中上述变量改为 `false`（或删除），然后再次执行重建命令。

---

## 故障排查

### 问题 1: `Permission denied (publickey)`

**原因**: SSH 公钥未添加到 GitHub

**解决**:
1. 检查公钥是否正确复制（没有多余空格或换行）
2. 确认添加到的是同一个 GitHub 账户
3. 验证 GitHub 设置 → SSH and GPG keys 中公钥存在

### 问题 2: `Could not resolve hostname`

**原因**: 服务器 DNS 配置问题

**解决**:
```bash
echo "8.8.8.8 github.com" >> /etc/hosts
```

### 问题 3: `Repository not found`

**原因**: 远程仓库 URL 配置错误

**解决**:
```bash
git remote set-url origin git@github.com:sev7n4/pintuotuo.git
```

---

## 安全注意事项

1. **密钥权限**: 确保私钥文件权限为 600
2. **只读访问**: 该密钥仅用于读取代码，不需要写入权限
3. **定期轮换**: 建议定期更换 SSH 密钥
4. **监控日志**: 定期检查服务器 SSH 登录日志

---

## 相关文件

- CI/CD 配置: `.github/workflows/deploy-tencent.yml`
- Docker Compose: `docker-compose.prod.yml`
- 环境变量: `.env` (服务器上维护，不提交到 Git)

---

## 维护记录

| 日期 | 操作 | 原因 |
|------|------|------|
| 2026-03-26 | 初始配置 | 实现服务器直接从 GitHub 拉取代码 |

---

## 联系方式

如有问题，请检查：
1. GitHub Actions 日志
2. 服务器 Docker 容器状态
3. Git 远程仓库配置
