# 🚀 Staging Deployment Configuration - Your Next Steps

**Status**: SSH Keys Generated ✅
**Date**: 2026-03-18
**Ready For**: AWS EC2 Setup

---

## Summary

Your SSH key pair has been generated successfully:

```
Private Key:  ~/.ssh/github_staging_key       (for GitHub secret)
Public Key:   ~/.ssh/github_staging_key.pub   (for EC2 instance)
```

---

## What You Need to Do (4 Steps - 30-45 minutes)

### Step 1: Create EC2 Instance on AWS (10 minutes)

1. **Go to AWS Console**: https://console.aws.amazon.com/ec2/
2. **Click "Launch instances"** in the top-right
3. **Configure with these settings**:

```
Instance Details:
  Name:                    pintuotuo-staging
  AMI:                     Ubuntu 22.04 LTS (Free Tier)
  Instance Type:           t3.micro (Free Tier) or t3.small ($9/month)
  Storage:                 20 GB SSD

Key Pair:
  Name:                    pintuotuo-deploy
  Action:                  Create new key pair
  Type:                    RSA
  (Download the .pem file and save to ~/.ssh/pintuotuo-deploy.pem)

Security Group:
  Name:                    pintuotuo-staging
  Inbound Rules:
    - SSH (22) from 0.0.0.0/0
    - HTTP (80) from 0.0.0.0/0
    - HTTPS (443) from 0.0.0.0/0

  (Later: restrict SSH to your IP for security)
```

4. **Click "Launch instance"**

5. **Wait for instance to start** (1-2 minutes)

6. **Note the following values** (you'll need these later):
   - Public IPv4 address: `52.1.2.3` (example - yours will be different)
   - Public DNS: `ec2-52-1-2-3.compute-1.amazonaws.com`

---

### Step 2: Setup EC2 Instance with Docker (10 minutes)

1. **SSH into your instance**:
```bash
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@YOUR_EC2_PUBLIC_IP
# Example: ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@52.1.2.3
```

2. **Once connected, run these commands** (copy-paste entire block):
```bash
# Update system
sudo apt-get update
sudo apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker ubuntu

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install PostgreSQL client
sudo apt-get install -y postgresql-client-14

# Create app directory
mkdir -p /app/pintuotuo
cd /app/pintuotuo

# Verify installations (should show version numbers)
echo "Docker version:"
docker --version
echo "Docker Compose version:"
docker-compose --version

# Create SSH directory for GitHub Actions
mkdir -p ~/.ssh
chmod 700 ~/.ssh
```

3. **Add GitHub Actions SSH Public Key**:
```bash
# Still on EC2, add the public key
cat >> ~/.ssh/authorized_keys << 'EOF'
PASTE_YOUR_PUBLIC_KEY_HERE
EOF

# Replace PASTE_YOUR_PUBLIC_KEY_HERE with contents of:
# ~/.ssh/github_staging_key.pub (the full ssh-rsa line)

# Set permissions
chmod 600 ~/.ssh/authorized_keys
```

**Your public key is**:
```
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCrXfMe5zaS3FDNt50cRzlaRdflDws/QUZJwAVOagXdZwdsMMAvJvoCtHqkf0UviRa+LBM2a+//A7M+5z6Vo8/89xrUdQQe/9sc2BPHF4XLUvkqp3WxvpuI7JaLAijpDS4JQz0G1xCpqg+31zukPlEM6nwzwlwDTYoSysvBEcHvmJ0F93CUSquliE1t3aa3yw/qYkXnSSlYTS3UPA82NWiv+jJ6DNaaOGrc3gkuxWiDvrQ3fyXF42js6nf/Ncq+RcB9ftVaqRru4Rm+jXxgxwDGCZRwuZ8WR9ITbMYh9dmLQZ8F9vmYkWUiwB9f2uDfvV8kxwjImLdmpfxc1pigOmhsZpelkx/Xqgmf7DoR7WAy8ipAhX6f83G0aGraY3kWjBgUMB2D18hFMxmnZAy34ES3qAqBiTYzB/mZTTpVifAtthkIIbFbOY1wgSlr1+LrEIvvLZp7X4O9saSdC4AqLrTaZYdBeVczxJ9V77a9q6eSD+DovS8Zshat/cBp/GXrdjYcDyLkro4Jx/a+pHIME6V1CX200hDZ2Z/1vVuPjRXIUUbevMR0F+PPz+4ci417h2yj0REgrBXqJm7RTDEmjLcmSQt0CPv3pszI05tTJS4XtDOXkbizlF8eVzmq+Z/aStScsvnnFRBGWHJ2XEEmwQu1bWJDGrmiq+2pDkmOpupQsw== 4seven@4sevendeMacBook-Pro.local
```

4. **Test GitHub Actions SSH Key** (from EC2):
```bash
# This should show "ubuntu"
whoami
# Exit
exit
```

5. **Test SSH key from your local machine**:
```bash
ssh -i ~/.ssh/github_staging_key ubuntu@YOUR_EC2_PUBLIC_IP whoami
# Should output: ubuntu
```

---

### Step 3: Add Secrets to GitHub (5-10 minutes)

1. **Go to GitHub Secrets**: https://github.com/sev7n4/pintuotuo/settings/secrets/actions

2. **Click "New repository secret"** and add these 4 secrets:

#### Secret 1: STAGING_SERVER
```
Name:  STAGING_SERVER
Value: 52.1.2.3
       (Replace with your EC2 public IPv4 address)
```

#### Secret 2: STAGING_USER
```
Name:  STAGING_USER
Value: ubuntu
```

#### Secret 3: STAGING_SSH_KEY
```
Name:  STAGING_SSH_KEY
Value: -----BEGIN OPENSSH PRIVATE KEY-----
       b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAACFwAAAAdzc2gtcn
       NhAAAAAwEAAQAAAgEAq13zHuc2ktxQzbedHEc5WkXX5Q8LP0FGScAFTmoF3WcHbDDALyb6
       ArR6pH9FL4kWviwTNmvv/wOzPuc+laPP/Pca1HUEHv/bHNgTxxeFy1L5Kqd1sb6biOyWiw
       Io6Q0uCUM9BtcQqaoPt9c7pD5RDOp8M8JcA02KEsrLwRHB75idBfdwlEqrpYhNbd2mt8sP
       6mJF50kpWE0t1DwPNjVor/oyegzWmjhq3N4JLsVog760N38lxeNo7Op3/zXKvkXAfX7VWq
       ka7uEZvo18YMcAxgmUcLmfFkfSE2zGIfXZi0GfBfb5mJFlIsAfX9rg371fJMcIyJi3ZqX8
       XNaYoDpobGaXpZMf16oJn+w6Ee1gMvIqQIV+n/NxtGhq2mN5FowYFDAdg9fIRTMZp2QMt+
       BEt6gKgYk2Mwf5mU06VYnwLbYZCCGxWzmNcIEpa9fi6xCL7y2ae1+DvbGknQuAKi602mWH
       QXlXM8SfVe+2vaunkg/g6L0vGbIWrf3Aafxl63Y2HA8i5K6OCcf2vqRyDBOldQl9tNIQ2d
       mf9b1bj40VyFFG3rzEdBfjz8/uHIuNe4dso9ERIKwV6iZu0UwxJoy3JkkLdAj796bMyNOb
       UyUuF7Qzl5G4s5RfHlc5qvmf2krUnLL55xUQRlhydlxBJsELtW1iQxq5oqvtqQ5JjqbqUL
       MAAAdYnqp79Z6qe/UAAAAHc3NoLXJzYQAAAgEAq13zHuc2ktxQzbedHEc5WkXX5Q8LP0FG
       ... (paste entire file from ~/.ssh/github_staging_key)
       -----END OPENSSH PRIVATE KEY-----
```

**OR faster**: Copy entire file:
```bash
# On your local machine
cat ~/.ssh/github_staging_key | pbcopy  # macOS
# Then paste into GitHub secret
```

#### Secret 4: STAGING_URL
```
Name:  STAGING_URL
Value: http://52.1.2.3
       (Replace with your EC2 public IPv4 address)
       Or use a domain if you have one
```

---

### Step 4: Test Deployment (5 minutes)

1. **Create docker-compose.yml on EC2**:
```bash
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
      JWT_SECRET: dev-secret-key
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
```

2. **Push to develop branch**:
```bash
git push origin develop
```

3. **Watch GitHub Actions**:
   - Go to: https://github.com/sev7n4/pintuotuo/actions
   - Find your workflow run
   - Wait for: CI Pipeline → Integration Tests → Deploy to Staging
   - Check for any errors

4. **Verify deployment**:
```bash
# Check containers running
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@YOUR_EC2_IP docker ps

# Test API
curl http://YOUR_EC2_IP:8080/api/v1/health

# Should return:
# {"status":"ok"}
```

---

## Your SSH Keys Location

**Private Key** (Keep Secret!):
```
~/.ssh/github_staging_key
```

**Public Key** (Share with servers):
```
~/.ssh/github_staging_key.pub
```

---

## Checklist

- [ ] AWS EC2 instance created and running
- [ ] SSH'd into EC2 successfully
- [ ] Docker installed on EC2
- [ ] Docker Compose installed on EC2
- [ ] Public key added to EC2 authorized_keys
- [ ] Tested SSH with new key from local machine
- [ ] STAGING_SERVER secret added to GitHub
- [ ] STAGING_USER secret added to GitHub
- [ ] STAGING_SSH_KEY secret added to GitHub
- [ ] STAGING_URL secret added to GitHub
- [ ] docker-compose.yml created on EC2
- [ ] Pushed code to develop branch
- [ ] GitHub Actions workflow started
- [ ] Deployment succeeded (no SSH errors)
- [ ] Docker container running on EC2
- [ ] API endpoint responding (curl test)

---

## Cost

| Service | Cost | Notes |
|---------|------|-------|
| EC2 t3.micro | Free (12 months) | Free tier for new AWS accounts |
| EC2 t3.small | ~$9/month | If need more resources |
| Data transfer | ~$0/month | Low during testing |

---

## Reference Files

All created/updated files in your project:

```
docs/
├── STAGING_QUICK_START.md      ← Quick reference
├── AWS_STAGING_SETUP.md        ← Detailed guide (this file)
├── SECRETS_SETUP_GUIDE.md      ← Secret configuration
└── README.md                   ← Overview of all docs

scripts/
└── setup-staging.sh            ← Setup script (already ran)

.github/workflows/
└── deploy-staging.yml          ← Already configured
```

---

## Need Help?

**SSH connection fails?**
- Check EC2 security group allows SSH (port 22)
- Verify AWS .pem key file has right permissions: `chmod 600 ~/.ssh/pintuotuo-deploy.pem`
- Try with verbose: `ssh -v -i ~/.ssh/pintuotuo-deploy.pem ubuntu@IP`

**Docker not found on EC2?**
- Might need to logout/login after `usermod -aG docker ubuntu`
- Run `docker --version` to verify

**GitHub Actions deployment fails?**
- Check workflow logs: https://github.com/sev7n4/pintuotuo/actions
- Look at "Deploy via SSH" step
- Common issue: STAGING_SSH_KEY not copied correctly (missing BEGIN/END lines)

**Can't curl the API?**
- Verify Docker container is running: `docker ps` on EC2
- Check backend logs: `docker-compose logs backend`
- Verify port 8080 is exposed in docker-compose.yml

---

## Next: Production (After Staging Works)

Once staging is working perfectly:

1. Create production EC2 instance
2. Generate new SSH key: `ssh-keygen -t rsa -b 4096 -f ~/.ssh/github_production_key`
3. Follow same setup steps for production
4. Add PRODUCTION_* secrets to GitHub
5. Create `.github/workflows/deploy-production.yml`
6. Trigger from main branch (manual approval)

---

**You're all set! Follow the 4 steps above to get staging deployment working.** 🚀

**Estimated Time**: 30-45 minutes
**Estimated Cost**: Free (first 12 months)
