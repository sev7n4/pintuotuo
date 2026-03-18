# AWS EC2 Staging Deployment Setup Guide

**Purpose**: Set up a staging environment on AWS EC2 and configure GitHub Actions for automated deployment
**Time Required**: 20-30 minutes
**Cost**: ~$5-15/month for micro/small instance

---

## Step 1: Create AWS EC2 Instance

### 1.1 Login to AWS Console
Go to: https://console.aws.amazon.com/ec2/

### 1.2 Launch New Instance

Click **"Launch instances"** and configure:

```
Name:                      pintuotuo-staging
AMI:                       Ubuntu 22.04 LTS (Free Tier eligible)
Instance Type:             t3.micro or t3.small
                          (Free Tier: t2.micro available)
Storage:                   20 GB SSD (Free Tier: 30 GB)
Key Pair:                  Create new "pintuotuo-deploy"
Security Group:            Create new "pintuotuo-staging"
```

### 1.3 Configure Security Group

**Inbound Rules**:
```
Protocol    Port        Source          Name
SSH         22          0.0.0.0/0       SSH (restrict to your IP for security)
HTTP        80          0.0.0.0/0       HTTP
HTTPS       443         0.0.0.0/0       HTTPS
```

**Outbound Rules**: Allow all (default)

### 1.4 Get Instance Details

Once instance is running, note down:
```
Public IPv4:      52.1.2.3 (example)
Public DNS:       ec2-52-1-2-3.compute-1.amazonaws.com
Instance ID:      i-0123456789abcdef0
Region:           us-east-1
```

---

## Step 2: Configure EC2 Instance for Deployment

### 2.1 Create SSH Key File

From AWS Console, download the `.pem` key file you created:
- Name: `pintuotuo-deploy.pem`
- Save to: `~/.ssh/pintuotuo-deploy.pem`

```bash
# Set proper permissions
chmod 600 ~/.ssh/pintuotuo-deploy.pem

# Verify
ls -la ~/.ssh/pintuotuo-deploy.pem
# Should show: -rw------- 1 user user ... pintuotuo-deploy.pem
```

### 2.2 SSH into Instance

```bash
# Test SSH connection
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@52.1.2.3

# Or use the DNS name
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@ec2-52-1-2-3.compute-1.amazonaws.com

# Should see:
# Welcome to Ubuntu 22.04.1 LTS ...
# ubuntu@ip-172-31-XX-XX:~$
```

### 2.3 Setup Instance Software (Run on EC2)

```bash
# SSH into instance first
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@<your-ip>

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

# Verify installations
docker --version
docker-compose --version

# Install PostgreSQL client (for database scripts)
sudo apt-get install -y postgresql-client-14

# Create app directory
mkdir -p /app/pintuotuo
cd /app/pintuotuo

# Logout
exit
```

---

## Step 3: Create SSH Key for GitHub Actions

Generate a dedicated SSH key for GitHub Actions (different from AWS key pair):

```bash
# On your local machine
ssh-keygen -t rsa -b 4096 -f ~/.ssh/github_staging_key -N ""

# Output:
# Your identification has been saved in /home/user/.ssh/github_staging_key
# Your public key has been saved in /home/user/.ssh/github_staging_key.pub

# View your keys
cat ~/.ssh/github_staging_key      # PRIVATE KEY (for GitHub secret)
cat ~/.ssh/github_staging_key.pub  # PUBLIC KEY (for server)
```

### Add Public Key to EC2 Instance

```bash
# On EC2 (SSH first with AWS key)
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@52.1.2.3

# Add GitHub Actions public key
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# Create authorized_keys file
cat >> ~/.ssh/authorized_keys << 'EOF'
# Paste the contents of ~/.ssh/github_staging_key.pub here
EOF

# Set proper permissions
chmod 600 ~/.ssh/authorized_keys

# Logout
exit
```

### Test GitHub Actions SSH Key

```bash
# On your local machine, test the new key
ssh -i ~/.ssh/github_staging_key ubuntu@52.1.2.3 "whoami"

# Should output:
# ubuntu
```

---

## Step 4: Prepare Values for GitHub Secrets

Gather all required values:

```bash
# 1. STAGING_SERVER (from AWS)
STAGING_SERVER=52.1.2.3
# or use DNS name:
STAGING_SERVER=ec2-52-1-2-3.compute-1.amazonaws.com

# 2. STAGING_USER (Ubuntu username)
STAGING_USER=ubuntu

# 3. STAGING_SSH_KEY (entire contents)
cat ~/.ssh/github_staging_key

# 4. STAGING_URL (public URL for app)
STAGING_URL=https://staging.yourdomain.com
# or use EC2 IP:
STAGING_URL=http://52.1.2.3

# 5. SLACK_WEBHOOK_URL (optional, from Slack)
# https://api.slack.com/messaging/webhooks
```

---

## Step 5: Add Secrets to GitHub

Go to: https://github.com/sev7n4/pintuotuo/settings/secrets/actions

### 5.1 Add STAGING_SERVER

```
Name:  STAGING_SERVER
Value: 52.1.2.3
       (or your EC2 public IP/DNS)
```

Click "Add secret"

### 5.2 Add STAGING_USER

```
Name:  STAGING_USER
Value: ubuntu
```

Click "Add secret"

### 5.3 Add STAGING_SSH_KEY

This is important! Include the entire key:

```
Name:  STAGING_SSH_KEY
Value: -----BEGIN RSA PRIVATE KEY-----
       MIIEpAIBAAKCAQEA1y5p8kZ9x5j7k3m9n2o3p4q5r...
       ...
       ...
       -----END RSA PRIVATE KEY-----
```

Click "Add secret"

### 5.4 Add STAGING_URL

```
Name:  STAGING_URL
Value: http://52.1.2.3
       (or https://staging.yourdomain.com if using domain)
```

Click "Add secret"

### 5.5 Add SLACK_WEBHOOK_URL (Optional)

1. Go to Slack API: https://api.slack.com/apps
2. Create "New App" → "From scratch"
3. Name: "Pintuotuo CI/CD"
4. Workspace: Your Slack workspace
5. Click "Incoming Webhooks" → Toggle ON
6. Click "Add New Webhook to Workspace"
7. Select channel: `#engineering` or `#deployments`
8. Copy webhook URL

```
Name:  SLACK_WEBHOOK_URL
Value: https://hooks.slack.com/services/T.../B.../X...
```

Click "Add secret"

---

## Step 6: Test the Setup

### 6.1 Create docker-compose.yml on EC2

SSH into staging server:

```bash
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@52.1.2.3

# Create docker-compose.yml
cd /app/pintuotuo
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  backend:
    image: pintuotuo/backend:staging
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgresql://pintuotuo:dev_password@postgres:5432/pintuotuo_db
      REDIS_URL: redis://redis:6379
      JWT_SECRET: ${JWT_SECRET:-dev-key}
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
    ports:
      - "6379:6379"
    restart: always

volumes:
  postgres_data:
EOF

# Exit
exit
```

### 6.2 Test Deployment

Push code to develop branch:

```bash
# Make a small change
echo "# Staging Deployment Test" >> README.md
git add README.md
git commit -m "test: trigger staging deployment"
git push origin develop
```

Watch GitHub Actions:
1. Go to: https://github.com/sev7n4/pintuotuo/actions
2. Find workflow run for develop branch
3. Wait for CI Pipeline → Integration Tests → Deploy to Staging
4. Check logs for any errors

### 6.3 Verify Deployment

```bash
# SSH into EC2
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@52.1.2.3

# Check running containers
docker ps

# Expected output:
# CONTAINER ID  IMAGE                    STATUS
# xxxxx         pintuotuo/backend:staging  Up 2 minutes

# Check logs
docker-compose logs -f backend

# Test API endpoint
curl http://localhost:8080/api/v1/health

# Should return:
# {"status":"ok"}
```

### 6.4 Check Slack Notification (if configured)

If you added SLACK_WEBHOOK_URL, you should see deployment notification in Slack channel.

---

## Step 7: Setup Domain (Optional)

To use a custom domain instead of IP address:

### Option A: Route53 (AWS)

1. In AWS Console, go to Route 53
2. Create hosted zone for your domain
3. Create A record:
   ```
   Name:    staging
   Type:    A
   Value:   52.1.2.3 (your EC2 public IP)
   ```
4. Update STAGING_URL secret to: `http://staging.yourdomain.com`

### Option B: Update hosts file (local testing)

```bash
# Edit /etc/hosts (Mac/Linux) or C:\Windows\System32\drivers\etc\hosts (Windows)
52.1.2.3  staging.yourdomain.com
```

### Option C: Update DNS provider

If you use Cloudflare, GoDaddy, etc.:
1. Add A record pointing to your EC2 IP
2. Wait for DNS propagation (5-30 minutes)

---

## Step 8: Production Deployment (Future)

Once staging works well, setup production:

```bash
# Create new EC2 instance for production
# Configure similar to staging
# Create new SSH key: github_production_key
# Add secrets: PRODUCTION_SERVER, PRODUCTION_USER, PRODUCTION_SSH_KEY, PRODUCTION_URL

# Create production workflow: .github/workflows/deploy-production.yml
# Trigger: Manual approval on main branch
# Similar to deploy-staging.yml but with production secrets
```

---

## Troubleshooting

### SSH Connection Refused

```bash
# Check security group allows SSH (port 22)
# AWS Console → EC2 → Security Groups → Check inbound rules

# Verify key file permissions
chmod 600 ~/.ssh/github_staging_key
chmod 600 ~/.ssh/pintuotuo-deploy.pem

# Try with verbose output
ssh -v -i ~/.ssh/pintuotuo-deploy.pem ubuntu@52.1.2.3
```

### Docker Container Won't Start

```bash
# SSH into EC2
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@52.1.2.3

# Check Docker status
docker ps -a

# Check logs
docker-compose logs backend

# Common issues:
# - Port 8080 already in use
# - Database connection failed
# - Image pull failed
```

### GitHub Actions Deploy Failed

Check workflow logs:
1. Go to Actions → Deploy to Staging workflow run
2. Click job → Click step "Deploy via SSH"
3. Look for error message
4. Common causes:
   - SSH key not configured correctly
   - Server unreachable
   - Docker command not found

### Slow or No Response from API

```bash
# SSH into EC2
ssh -i ~/.ssh/pintuotuo-deploy.pem ubuntu@52.1.2.3

# Check available resources
free -h          # Memory
df -h            # Disk
top -b -n 1      # CPU

# For t3.micro: limited resources, may need t3.small
```

---

## Cleanup/Termination

To stop paying for staging:

```bash
# AWS Console → EC2 → Instances
# Right-click instance → Terminate
# Warning: This deletes everything on the instance
```

---

## Checklist

- [ ] EC2 instance created and running
- [ ] Instance can SSH with AWS key
- [ ] Docker and Docker Compose installed
- [ ] GitHub Actions SSH key generated
- [ ] Public key added to EC2 authorized_keys
- [ ] STAGING_SERVER secret added to GitHub
- [ ] STAGING_USER secret added to GitHub
- [ ] STAGING_SSH_KEY secret added to GitHub
- [ ] STAGING_URL secret added to GitHub
- [ ] SLACK_WEBHOOK_URL secret added (optional)
- [ ] docker-compose.yml created on EC2
- [ ] Test deployment by pushing to develop
- [ ] Verified container running on EC2
- [ ] Tested API endpoint works
- [ ] Received Slack notification (if configured)

---

**Document Version**: 1.0
**Last Updated**: 2026-03-18
**Status**: Ready for Implementation
