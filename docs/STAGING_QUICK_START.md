# 🚀 Quick Staging Deployment Setup - Summary

**Time Required**: 30-45 minutes
**Cost**: ~$5-15/month AWS EC2
**Status**: Ready to implement

---

## Overview

This guide will help you:
1. ✅ Generate SSH keys for GitHub Actions
2. ✅ Create EC2 instance on AWS
3. ✅ Configure instance for Docker deployment
4. ✅ Add secrets to GitHub
5. ✅ Test automated deployment

---

## Quick Start - 4 Commands

### Command 1: Generate SSH Keys & Show Setup Info

```bash
cd /Users/4seven/pintuotuo
bash scripts/setup-staging.sh
```

This will:
- ✅ Create SSH key pair for GitHub Actions
- ✅ Show private key (copy for GitHub secret)
- ✅ Show public key (copy to EC2)
- ✅ Create checklist template

### Command 2: Create AWS EC2 Instance

1. Go to AWS Console: https://console.aws.amazon.com/ec2/
2. Click **"Launch instances"**
3. Configure as shown in `docs/AWS_STAGING_SETUP.md` Step 1
4. Note the public IP (example: `52.1.2.3`)

**Quick Config**:
```
Name:              pintuotuo-staging
AMI:               Ubuntu 22.04 LTS
Instance Type:     t3.micro (free tier) or t3.small
Storage:           20GB SSD
Key Pair:          pintuotuo-deploy (download the .pem file)
Security Group:    Allow SSH (22), HTTP (80), HTTPS (443)
```

### Command 3: Setup Instance (SSH into EC2)

```bash
# SSH using AWS key pair
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@YOUR_EC2_IP

# Then copy-paste this entire block:
sudo apt-get update && sudo apt-get upgrade -y
curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh
sudo usermod -aG docker ubuntu
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo apt-get install -y postgresql-client-14
mkdir -p /app/pintuotuo

# Verify installations
docker --version
docker-compose --version

# Create authorized_keys for GitHub Actions
mkdir -p ~/.ssh
chmod 700 ~/.ssh
cat >> ~/.ssh/authorized_keys << 'PASTE_PUBLIC_KEY_HERE'
# Paste contents of ~/.ssh/github_staging_key.pub from Step 1
PASTE_PUBLIC_KEY_HERE
chmod 600 ~/.ssh/authorized_keys

exit
```

### Command 4: Add Secrets to GitHub

1. Go to: https://github.com/sev7n4/pintuotuo/settings/secrets/actions
2. Click **"New repository secret"** for each:

```bash
# Get these values ready:
STAGING_SERVER=YOUR_EC2_IP        # Example: 52.1.2.3
STAGING_USER=ubuntu
STAGING_SSH_KEY=$(cat ~/.ssh/github_staging_key)
STAGING_URL=http://YOUR_EC2_IP    # Example: http://52.1.2.3
```

**Add to GitHub** (paste each):

| Secret Name | Value |
|---|---|
| `STAGING_SERVER` | `52.1.2.3` |
| `STAGING_USER` | `ubuntu` |
| `STAGING_SSH_KEY` | (entire file contents) |
| `STAGING_URL` | `http://52.1.2.3` |
| `SLACK_WEBHOOK_URL` | (optional) |

---

## Test Deployment (5 minutes)

```bash
# Create docker-compose.yml on EC2
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@YOUR_EC2_IP

cat > /app/pintuotuo/docker-compose.yml << 'EOF'
version: '3.8'
services:
  backend:
    image: pintuotuo/backend:staging
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgresql://pintuotuo:dev_password@postgres:5432/pintuotuo_db
      REDIS_URL: redis://redis:6379
      JWT_SECRET: dev-key
      GIN_MODE: release
    depends_on:
      - postgres
      - redis
    restart: always

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: pintuotuo
      POSTGRES_PASSWORD: dev_password
      POSTGRES_DB: pintuotuo_db
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: always

  redis:
    image: redis:7-alpine
    restart: always

volumes:
  postgres_data:
EOF

exit

# Push to develop to trigger deployment
git push origin develop

# Watch deployment in GitHub Actions
# https://github.com/sev7n4/pintuotuo/actions
```

---

## Reference Documents

| Document | Purpose |
|----------|---------|
| `docs/AWS_STAGING_SETUP.md` | Complete detailed setup guide |
| `docs/SECRETS_SETUP_GUIDE.md` | Secret configuration details |
| `docs/LOCAL_TESTING_GUIDE.md` | Testing before deployment |
| `scripts/setup-staging.sh` | Automated setup script |

---

## Troubleshooting

### SSH Connection Failed
```bash
# Check security group allows SSH (port 22)
# Check instance is running
# Verify AWS key file: chmod 600 ~/.ssh/pintuotuo-deploy.pem
```

### Docker Not Found After SSH
```bash
# Add user to docker group (requires logout/login)
sudo usermod -aG docker ubuntu
# Logout and SSH back in
```

### Deployment Failed in GitHub Actions
1. Check GitHub Actions logs: https://github.com/sev7n4/pintuotuo/actions
2. Look for "Deploy via SSH" step
3. Most common: SSH key not formatted correctly in secret
4. See `docs/SECRETS_SETUP_GUIDE.md` Troubleshooting section

---

## Checklist

- [ ] Run `bash scripts/setup-staging.sh`
- [ ] Create EC2 instance on AWS
- [ ] SSH into EC2 and run setup commands
- [ ] Add public key to EC2 authorized_keys
- [ ] Test SSH with new key: `ssh -i ~/.ssh/github_staging_key ubuntu@YOUR_IP`
- [ ] Add STAGING_SERVER secret to GitHub
- [ ] Add STAGING_USER secret to GitHub
- [ ] Add STAGING_SSH_KEY secret to GitHub
- [ ] Add STAGING_URL secret to GitHub
- [ ] Create docker-compose.yml on EC2
- [ ] Push to develop branch
- [ ] Monitor GitHub Actions workflow
- [ ] Verify container running: `docker ps`
- [ ] Test API: `curl http://YOUR_IP:8080/api/v1/health`
- [ ] Check Slack notification (if configured)

---

## Cost Estimate

| Item | Cost | Notes |
|------|------|-------|
| EC2 t3.micro | Free (12 months) | Free tier if new account |
| EC2 t3.small | $9-15/month | If need more resources |
| Data transfer | ~$0.01/month | Low during testing |
| **Total** | **Free - $15/month** | |

---

## Next: Production Deployment

Once staging is working:
1. Create separate production EC2 instance
2. Create new SSH key: `github_production_key`
3. Add production secrets to GitHub
4. Create `.github/workflows/deploy-production.yml`
5. Trigger manually from main branch

---

## Getting Help

1. **Setup stuck?** → Read `docs/AWS_STAGING_SETUP.md` in detail
2. **Secrets not working?** → Check `docs/SECRETS_SETUP_GUIDE.md`
3. **Testing locally first?** → Use `docs/LOCAL_TESTING_GUIDE.md`
4. **Understanding workflows?** → Review `docs/GITHUB_WORKFLOWS.md`

---

**Ready to start?** Begin with: `bash scripts/setup-staging.sh`

**Estimated Time**: 30-45 minutes from start to working deployment ⏱️
