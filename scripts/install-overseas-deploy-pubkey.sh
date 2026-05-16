#!/usr/bin/env bash
# 将 ~/.ssh/pintuotuo_overseas_deploy.pub 写入海外机 authorized_keys
# 与 GitHub production Secret TENCENT_CLOUD_OVERSEAS_SSH_KEY（私钥）成对使用。
#
# 用法：
#   ./scripts/install-overseas-deploy-pubkey.sh
#   OVERSEAS_SSH=ubuntu@43.160.204.9 ./scripts/install-overseas-deploy-pubkey.sh

set -euo pipefail

OVERSEAS_SSH="${OVERSEAS_SSH:-ubuntu@43.160.204.9}"
KEY_PRIV="${KEY_PRIV:-$HOME/.ssh/pintuotuo_overseas_deploy}"
KEY_PUB="${KEY_PUB:-${KEY_PRIV}.pub}"

if [ ! -f "$KEY_PRIV" ] || [ ! -f "$KEY_PUB" ]; then
  echo "Generating ${KEY_PRIV} ..."
  ssh-keygen -t ed25519 -f "$KEY_PRIV" -N "" -C "github-actions-litellm-overseas"
fi

PUB_LINE=$(cat "$KEY_PUB")
echo "Public key to install:"
echo "  ${PUB_LINE}"

if ssh -o BatchMode=yes -o ConnectTimeout=8 -i "$KEY_PRIV" "$OVERSEAS_SSH" 'echo ok' 2>/dev/null; then
  echo "Already authorized. Nothing to do."
  exit 0
fi

if [ -z "${OVERSEAS_SSH_PASSWORD:-}" ]; then
  if command -v osascript >/dev/null 2>&1; then
    OVERSEAS_SSH_PASSWORD=$(osascript -e 'Tell application "System Events" to display dialog "海外机 SSH 密码 (ubuntu@43.160.204.9):" default answer "" with hidden answer' -e 'text returned of result' 2>/dev/null || true)
  fi
fi

if [ -z "${OVERSEAS_SSH_PASSWORD:-}" ]; then
  echo "Set OVERSEAS_SSH_PASSWORD or run on macOS to use dialog prompt."
  echo "Or manually: ssh-copy-id -i ${KEY_PUB} ${OVERSEAS_SSH}"
  exit 1
fi

if ! command -v sshpass >/dev/null 2>&1; then
  echo "sshpass required. Install: brew install sshpass"
  exit 1
fi

export SSHPASS="$OVERSEAS_SSH_PASSWORD"
REMOTE_USER="${OVERSEAS_SSH%@*}"
REMOTE_HOST="${OVERSEAS_SSH#*@}"

sshpass -e ssh -o StrictHostKeyChecking=no "${OVERSEAS_SSH}" "mkdir -p ~/.ssh && chmod 700 ~/.ssh && touch ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"

if sshpass -e ssh -o StrictHostKeyChecking=no "${OVERSEAS_SSH}" "grep -qxF '${PUB_LINE}' ~/.ssh/authorized_keys"; then
  echo "Public key already present in authorized_keys."
else
  sshpass -e ssh -o StrictHostKeyChecking=no "${OVERSEAS_SSH}" "echo '${PUB_LINE}' >> ~/.ssh/authorized_keys"
  echo "Public key appended."
fi

unset SSHPASS OVERSEAS_SSH_PASSWORD

if ssh -o BatchMode=yes -o ConnectTimeout=8 -i "$KEY_PRIV" "$OVERSEAS_SSH" 'echo SSH key auth OK'; then
  echo "Verified: key-based login works."
else
  echo "ERROR: key login still failing after install."
  exit 1
fi

echo ""
echo "Next: add private key to GitHub → Settings → Environments → production →"
echo "  Secret name: TENCENT_CLOUD_OVERSEAS_SSH_KEY"
echo "  Value: contents of ${KEY_PRIV}"
echo "(Do not paste the private key in chat.)"
