# GitHub Secrets Setup Guide

**Last Updated**: 2026-03-18
**Purpose**: Configure GitHub Actions secrets for CI/CD pipeline

---

## Overview

GitHub Secrets are encrypted environment variables used by workflows. They're essential for:
- Docker Hub authentication (push images)
- Staging server deployment (SSH keys)
- Third-party integrations (Slack notifications)
- API tokens and credentials

### Security Notes

✅ **Best Practices**:
- Secrets are encrypted and never logged
- Only visible to workflow runs (not in logs)
- Rotated regularly (every 90 days recommended)
- Use minimal permissions (least privilege)
- One secret per credential

❌ **Never Do**:
- Commit secrets to code or git history
- Share secrets in chat/email
- Reuse old secrets
- Print secrets in workflow logs
- Use personal credentials

---

## Table of Contents

1. [Currently Configured Secrets](#currently-configured-secrets)
2. [Required Secrets for Deployment](#required-secrets-for-deployment)
3. [Step-by-Step Setup Guide](#step-by-step-setup-guide)
4. [Verification Steps](#verification-steps)
5. [Troubleshooting](#troubleshooting)
6. [Rotating Secrets](#rotating-secrets)

---

## Currently Configured Secrets

### ✅ Docker Hub Credentials (Complete)

**Status**: Fully configured and working

| Secret | Value | Used For |
|--------|-------|----------|
| `DOCKER_USERNAME` | (your Docker username) | Push backend image to Docker Hub |
| `DOCKER_PASSWORD` | (your Docker token) | Authenticate with Docker Hub |

**Where Used**:
- `.github/workflows/test.yml` - Build Docker image
- `.github/workflows/integration-tests.yml` - Build Docker image
- `.github/workflows/deploy-staging.yml` - Login to Docker Hub

**Status Check**:
```bash
# These workflows successfully build and push images
# ✅ Integration Tests job builds Docker image
# ✅ Images tagged as pintuotuo-backend:latest
```

---

## Required Secrets for Deployment

To enable automatic staging deployment, configure these secrets:

### 1. STAGING_SERVER ⚠️ Required

**What**: IP address or hostname of staging environment

**Example Values**:
```
staging.pintuotuo.com
45.142.182.91
staging-server.internal
```

**Where Used**:
- `.github/workflows/deploy-staging.yml` - SSH connection target

---

### 2. STAGING_USER ⚠️ Required

**What**: SSH username for staging server deployment

**Example Values**:
```
pintuotuo
deploy
ubuntu
ec2-user
```

**Where Used**:
- `.github/workflows/deploy-staging.yml` - SSH username

---

### 3. STAGING_SSH_KEY ⚠️ Required

**What**: Private SSH key for passwordless authentication

**Characteristics**:
- 4096-bit RSA key (recommended)
- Multiline secret (include all lines)
- Must correspond to public key on staging server
- Never share or commit to git

**Example Format**:
```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1y5p8kZ9x5j7k3m9n2o3p4q5r6s7t8u9v0w1x2y3z4a5b6c7d
8e9f0g1h2i3j4k5l6m7n8o9p0q1r2s3t4u5v6w7x8y9z0a1b2c3d4e5f6g7h8i9
...
-----END RSA PRIVATE KEY-----
```

**Where Used**:
- `.github/workflows/deploy-staging.yml` - SSH private key for authentication

---

### 4. STAGING_URL ⚠️ Required

**What**: Public URL of staging environment

**Example Values**:
```
https://staging.pintuotuo.com
https://staging-api.pintuotuo.com
https://staging.example.com:8080
```

**Where Used**:
- `.github/workflows/deploy-staging.yml` - Health checks, PR comments
- Helps developers verify deployment

---

### 5. SLACK_WEBHOOK_URL ⚠️ Required (for notifications)

**What**: Slack incoming webhook URL for notifications

**Where Used**:
- `.github/workflows/integration-tests.yml` - Send test results
- `.github/workflows/deploy-staging.yml` - Send deployment status

---

## Step-by-Step Setup Guide

### Step 1: Generate SSH Key Pair

Generate a new SSH key for staging deployment:

```bash
# Generate 4096-bit RSA key
ssh-keygen -t rsa -b 4096 -f ~/.ssh/staging_key -N ""

# Output:
# ~/.ssh/staging_key (private key)
# ~/.ssh/staging_key.pub (public key)
```

**View your keys**:
```bash
# Private key (for STAGING_SSH_KEY secret)
cat ~/.ssh/staging_key

# Public key (for server)
cat ~/.ssh/staging_key.pub
```

**Keep safe**:
- Private key: → GitHub secret
- Public key: → Staging server

---

### Step 2: Add Public Key to Staging Server

Copy public key to staging server:

```bash
# On your local machine:
cat ~/.ssh/staging_key.pub

# Copy the entire output
# -----BEGIN RSA PUBLIC KEY-----
# ... (copy all lines)
# -----END RSA PUBLIC KEY-----
```

Then on staging server:

```bash
# SSH into staging server
ssh pintuotuo@staging.pintuotuo.com

# Add your public key
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# Paste your public key
echo "paste_public_key_here" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys

# Test (should not require password)
# Exit and reconnect
exit
```

**Verify SSH works**:
```bash
ssh -i ~/.ssh/staging_key pintuotuo@staging.pintuotuo.com
# Should log in without password prompt
```

---

### Step 3: Add Secrets to GitHub

1. **Go to GitHub Settings**
   ```
   https://github.com/pintuotuo/pintuotuo/settings/secrets/actions
   ```

2. **Click "New repository secret"**

3. **Add STAGING_SERVER**
   - Name: `STAGING_SERVER`
   - Value: `staging.pintuotuo.com` (or your IP)
   - Click "Add secret"

4. **Add STAGING_USER**
   - Name: `STAGING_USER`
   - Value: `pintuotuo` (or your deployment user)
   - Click "Add secret"

5. **Add STAGING_SSH_KEY** (Multiline)
   - Name: `STAGING_SSH_KEY`
   - Value: Paste entire private key file contents
     ```
     -----BEGIN RSA PRIVATE KEY-----
     MIIEpAIBAAKCAQEA...
     ...
     -----END RSA PRIVATE KEY-----
     ```
   - Click "Add secret"

6. **Add STAGING_URL**
   - Name: `STAGING_URL`
   - Value: `https://staging.pintuotuo.com`
   - Click "Add secret"

7. **Add SLACK_WEBHOOK_URL** (Optional)
   - Name: `SLACK_WEBHOOK_URL`
   - Value: `https://hooks.slack.com/services/T.../B.../X...`
   - Click "Add secret"

---

### Step 4: Create Slack Webhook (Optional)

For Slack notifications:

1. **Go to Slack API**
   ```
   https://api.slack.com/apps
   ```

2. **Create New App**
   - From scratch
   - Name: "Pintuotuo CI/CD"
   - Workspace: Your Slack workspace

3. **Enable Incoming Webhooks**
   - Left sidebar: "Incoming Webhooks"
   - Toggle ON

4. **Add New Webhook to Workspace**
   - Click "Add New Webhook to Workspace"
   - Select channel: `#engineering` or similar
   - Click "Allow"

5. **Copy Webhook URL**
   - Copy the URL shown
   - Add to GitHub secret `SLACK_WEBHOOK_URL`

---

## Verification Steps

### ✅ Verify Docker Hub Credentials

```bash
# Test locally
docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD

# Should output:
# Login Succeeded
```

### ✅ Verify SSH Key

```bash
# Test SSH access
ssh -i ~/.ssh/staging_key pintuotuo@staging.pintuotuo.com "whoami"

# Should output:
# pintuotuo
```

### ✅ Verify Slack Webhook

```bash
# Test Slack webhook
curl -X POST $SLACK_WEBHOOK_URL \
  -H 'Content-Type: application/json' \
  -d '{"text":"Test message from CI/CD"}'

# Should appear in your Slack channel
```

### ✅ Test Full Workflow

1. **Push to develop branch**
   ```bash
   git push origin develop
   ```

2. **Check GitHub Actions**
   - Go to: https://github.com/pintuotuo/pintuotuo/actions
   - Find your workflow run
   - Watch it execute:
     - CI Pipeline ✅
     - Integration Tests ✅
     - Deploy to Staging ✅
     - Slack notification ✅

3. **Verify deployment**
   ```bash
   # Check staging environment
   curl https://staging.pintuotuo.com/api/v1/health

   # Should return:
   # {"status":"ok"}
   ```

4. **Check Slack**
   - Should see notification in #engineering
   - Contains deployment status and link

---

## Secret References in Workflows

### Example: Using Secrets in Workflow

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Login to Docker Hub
        run: docker login -u ${{ secrets.DOCKER_USERNAME }} -p ${{ secrets.DOCKER_PASSWORD }}

      - name: Deploy via SSH
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.STAGING_SSH_KEY }}" > ~/.ssh/staging_key
          chmod 600 ~/.ssh/staging_key

          ssh -i ~/.ssh/staging_key \
              ${{ secrets.STAGING_USER }}@${{ secrets.STAGING_SERVER }} \
              "cd /app && docker-compose up -d"
```

**Important**:
- Reference secrets as `${{ secrets.SECRET_NAME }}`
- Secrets are masked in logs (shown as `***`)
- Never echo or print secrets

---

## Troubleshooting

### "Permission denied (publickey)" Error

**Cause**: SSH authentication failed

**Fix**:
1. Verify public key is on server: `cat ~/.ssh/authorized_keys`
2. Check permissions: `chmod 600 ~/.ssh/authorized_keys`
3. Verify private key: `chmod 600 ~/.ssh/staging_key`
4. Test manually: `ssh -i ~/.ssh/staging_key user@server`

---

### "denied: requested access to the resource is denied"

**Cause**: Docker Hub credentials invalid

**Fix**:
1. Generate new Docker token (not password):
   - https://hub.docker.com/settings/security
   - Generate Access Token
   - Update `DOCKER_PASSWORD` secret
2. Test locally: `docker login`

---

### Slack Webhook Not Working

**Cause**: Invalid webhook URL or wrong channel

**Fix**:
1. Verify webhook URL: `https://hooks.slack.com/services/T.../B.../X...`
2. Check channel exists
3. Test with curl:
   ```bash
   curl -X POST "$SLACK_WEBHOOK_URL" \
     -H 'Content-Type: application/json' \
     -d '{"text":"test"}'
   ```

---

### "Secret not found" in Workflow

**Cause**: Secret name mismatch or not added

**Fix**:
1. Check secret name in workflow: `${{ secrets.STAGING_SERVER }}`
2. Verify in GitHub Settings: exactly matches name
3. Wait 5 minutes (GitHub sometimes caches)
4. Rerun workflow

---

### Deployment Hangs/Timeout

**Cause**: SSH key not working or server unreachable

**Fix**:
1. Check staging server is reachable:
   ```bash
   ssh -i ~/.ssh/staging_key pintuotuo@$STAGING_SERVER "echo works"
   ```
2. Increase timeout in workflow:
   ```yaml
   - name: Deploy
     timeout-minutes: 15
     run: ...
   ```

---

## Security Best Practices

### 1. Rotate Secrets Regularly

**Every 90 days**:
```bash
# Generate new SSH key
ssh-keygen -t rsa -b 4096 -f ~/.ssh/staging_key_new

# Update on server
ssh -i ~/.ssh/staging_key "echo new_key >> ~/.ssh/authorized_keys"

# Update GitHub secret
# Copy new ~/.ssh/staging_key_new to STAGING_SSH_KEY

# Remove old key from server
ssh -i ~/.ssh/staging_key "..."
```

### 2. Limit SSH Key Permissions

On server, restrict SSH key:

```bash
# Restrict to specific command
echo 'command="/usr/local/bin/deploy" ssh-rsa AAAA...' >> ~/.ssh/authorized_keys

# Or restrict to specific IP
echo 'from="1.2.3.4" ssh-rsa AAAA...' >> ~/.ssh/authorized_keys
```

### 3. Monitor Secret Usage

In GitHub:
1. Go to Settings → Security log
2. Review who accessed secrets
3. Check for unusual access patterns

### 4. Revoke Compromised Secrets

If secret leaked:
1. Delete from GitHub immediately
2. Revoke on external service (Docker Hub, etc.)
3. Update in all places
4. Review git history to ensure not committed

---

## Reference

### Secret Names Used in Workflows

| Workflow | Secrets Used |
|----------|--------------|
| ci-pipeline.yml | None |
| test.yml | `DOCKER_USERNAME`, `DOCKER_PASSWORD` |
| integration-tests.yml | `SLACK_WEBHOOK_URL` |
| deploy-staging.yml | `DOCKER_USERNAME`, `DOCKER_PASSWORD`, `STAGING_SERVER`, `STAGING_USER`, `STAGING_SSH_KEY`, `STAGING_URL`, `SLACK_WEBHOOK_URL` |

### GitHub UI Paths

| Task | URL |
|------|-----|
| View secrets | Settings → Secrets and variables → Actions |
| Add secret | Click "New repository secret" |
| Update secret | Delete and recreate (can't edit in place) |
| Delete secret | Click delete icon |
| Audit log | Settings → Security log |

---

## Checklist for Deployment

- [ ] STAGING_SERVER configured
- [ ] STAGING_USER configured
- [ ] STAGING_SSH_KEY configured (SSH key pair generated)
- [ ] Public key added to staging server
- [ ] STAGING_URL configured
- [ ] DOCKER_USERNAME configured
- [ ] DOCKER_PASSWORD configured (use token, not password)
- [ ] Verified SSH access works
- [ ] Verified Docker login works
- [ ] Test push to develop branch
- [ ] Verify deployment succeeds
- [ ] Verify Slack notification (if configured)

---

**Document Version**: 1.0
**Status**: Ready to Use ✅
**Maintained By**: DevOps / Security Team
