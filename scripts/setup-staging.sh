#!/bin/bash
# AWS EC2 Staging Setup - Local Machine Script
# This script helps prepare local environment for GitHub Actions secrets

set -e

echo "🚀 Pintuotuo Staging Deployment Setup"
echo "========================================"
echo ""

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Check prerequisites
echo -e "${BLUE}Step 1: Checking prerequisites...${NC}"
if ! command -v ssh-keygen &> /dev/null; then
    echo "❌ ssh-keygen not found. Please install OpenSSH."
    exit 1
fi
echo -e "${GREEN}✓ ssh-keygen available${NC}"

# Step 2: Generate SSH key for GitHub Actions
echo ""
echo -e "${BLUE}Step 2: Generate SSH key for GitHub Actions deployment${NC}"

KEY_PATH="$HOME/.ssh/github_staging_key"
if [ -f "$KEY_PATH" ]; then
    echo -e "${YELLOW}⚠️  Key already exists at $KEY_PATH${NC}"
    read -p "Overwrite? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Keeping existing key"
        KEY_EXISTS=true
    else
        rm "$KEY_PATH" "$KEY_PATH.pub"
        ssh-keygen -t rsa -b 4096 -f "$KEY_PATH" -N ""
        echo -e "${GREEN}✓ SSH key generated${NC}"
    fi
else
    ssh-keygen -t rsa -b 4096 -f "$KEY_PATH" -N ""
    echo -e "${GREEN}✓ SSH key generated${NC}"
fi

# Step 3: Display key information
echo ""
echo -e "${BLUE}Step 3: SSH Key Information${NC}"
echo -e "${YELLOW}Private Key (for GitHub secret STAGING_SSH_KEY):${NC}"
echo "Location: $KEY_PATH"
echo ""
echo "Public Key (to add to EC2 instance):"
cat "$KEY_PATH.pub"
echo ""

# Step 4: Prepare GitHub secrets template
echo ""
echo -e "${BLUE}Step 4: Prepare GitHub Secrets${NC}"
echo ""
echo "You need to add these secrets to GitHub:"
echo "https://github.com/sev7n4/pintuotuo/settings/secrets/actions"
echo ""

cat > /tmp/github_secrets_template.txt << 'EOF'
==========================================
GITHUB SECRETS TO ADD
==========================================

1. STAGING_SERVER
   Value: <Your EC2 public IP or DNS>
   Example: 52.1.2.3

2. STAGING_USER
   Value: ubuntu

3. STAGING_SSH_KEY
   Value: (Entire contents of ~/.ssh/github_staging_key)

4. STAGING_URL
   Value: http://<Your EC2 IP>
   Example: http://52.1.2.3

5. SLACK_WEBHOOK_URL (Optional)
   Value: https://hooks.slack.com/services/...

==========================================
EOF

cat /tmp/github_secrets_template.txt

echo ""
echo -e "${YELLOW}📋 Saved template to: /tmp/github_secrets_template.txt${NC}"

# Step 5: Next steps
echo ""
echo -e "${BLUE}Step 5: Next Steps${NC}"
echo ""
echo "1️⃣  Follow AWS_STAGING_SETUP.md to:"
echo "   - Create EC2 instance"
echo "   - Configure security group"
echo "   - Install Docker & Docker Compose"
echo "   - Add your public key to authorized_keys"
echo ""
echo "2️⃣  Gather these values:"
echo "   - STAGING_SERVER (EC2 public IP)"
echo "   - STAGING_USER (usually 'ubuntu')"
echo "   - STAGING_SSH_KEY (contents of $KEY_PATH)"
echo "   - STAGING_URL (your EC2 public address)"
echo ""
echo "3️⃣  Add secrets to GitHub:"
echo "   Go to: https://github.com/sev7n4/pintuotuo/settings/secrets/actions"
echo "   Add each secret from the template above"
echo ""
echo "4️⃣  Test deployment:"
echo "   Push to develop branch"
echo "   Watch GitHub Actions workflow"
echo ""

echo -e "${GREEN}✅ Setup script complete!${NC}"
echo ""
echo "📖 Full guide: docs/AWS_STAGING_SETUP.md"
echo ""
