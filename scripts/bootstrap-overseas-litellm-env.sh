#!/usr/bin/env bash
# 将大陆生产的 LITELLM_MASTER_KEY 同步到海外 /opt/pintuotuo-litellm/.env
#
# 用法（在本机，已能 ssh 大陆机）：
#   ./scripts/bootstrap-overseas-litellm-env.sh
#
# 用法（指定海外目标，需本机可 ssh ubuntu@海外IP）：
#   OVERSEAS_SSH=ubuntu@43.160.204.9 ./scripts/bootstrap-overseas-litellm-env.sh --install
#
# 仅在大陆机生成待拷贝包（海外暂不能 ssh 时）：
#   ./scripts/bootstrap-overseas-litellm-env.sh --prepare-only

set -euo pipefail

DOMESTIC_SSH="${DOMESTIC_SSH:-root@119.29.173.89}"
DOMESTIC_KEY="${DOMESTIC_KEY:-$HOME/.ssh/tencent_cloud_deploy}"
OVERSEAS_SSH="${OVERSEAS_SSH:-ubuntu@43.160.204.9}"
OVERSEAS_DIR="${OVERSEAS_DIR:-/opt/pintuotuo-litellm}"
BOOT_DIR="/root/pintuotuo-overseas-bootstrap"

ssh_domestic() {
  ssh -o StrictHostKeyChecking=no -i "$DOMESTIC_KEY" "$DOMESTIC_SSH" "$@"
}

prepare_on_domestic() {
  ssh_domestic 'bash -s' <<SCRIPT
set -euo pipefail
set -a
# shellcheck source=/dev/null
source /opt/pintuotuo/.env
set +a
mkdir -p ${BOOT_DIR}
umask 077
cat > ${BOOT_DIR}/.env <<ENVEOF
LITELLM_MASTER_KEY=\${LITELLM_MASTER_KEY}
LITELLM_IMAGE_TAG=v1.83.3-stable
ENVEOF
chmod 600 ${BOOT_DIR}/.env
cat > ${BOOT_DIR}/install-on-overseas.sh <<'INST'
#!/usr/bin/env bash
set -euo pipefail
TARGET_DIR="${OVERSEAS_DIR:-/opt/pintuotuo-litellm}"
sudo mkdir -p "\$TARGET_DIR"
sudo chown "\$USER:\$USER" "\$TARGET_DIR"
install -m 600 .env "\$TARGET_DIR/.env"
echo "Installed \$TARGET_DIR/.env"
grep -E '^LITELLM_MASTER_KEY=' "\$TARGET_DIR/.env" | sed 's/=.*/=***configured***/'
INST
chmod 755 ${BOOT_DIR}/install-on-overseas.sh
echo "Prepared ${BOOT_DIR} on domestic host."
SCRIPT
}

install_on_overseas() {
  local tmp
  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT
  scp -o StrictHostKeyChecking=no -i "$DOMESTIC_KEY" \
    "${DOMESTIC_SSH}:${BOOT_DIR}/.env" \
    "${DOMESTIC_SSH}:${BOOT_DIR}/install-on-overseas.sh" \
    "$tmp/"
  scp -o StrictHostKeyChecking=no "$tmp/.env" "$tmp/install-on-overseas.sh" "${OVERSEAS_SSH}:/tmp/"
  ssh -o StrictHostKeyChecking=no "$OVERSEAS_SSH" "cd /tmp && OVERSEAS_DIR=${OVERSEAS_DIR} bash ./install-on-overseas.sh"
}

case "${1:-}" in
  --prepare-only)
    prepare_on_domestic
    echo "Next: scp bundle to overseas or run with --install when OVERSEAS_SSH works."
    ;;
  --install)
    prepare_on_domestic
    install_on_overseas
    ;;
  *)
    prepare_on_domestic
    if ssh -o StrictHostKeyChecking=no -o BatchMode=yes "$OVERSEAS_SSH" true 2>/dev/null; then
      install_on_overseas
    else
      echo "Overseas SSH not available in BatchMode. Prepared on domestic: ${BOOT_DIR}"
      echo "Manual:"
      echo "  scp -i ${DOMESTIC_KEY} ${DOMESTIC_SSH}:${BOOT_DIR}/.env ${OVERSEAS_SSH}:/tmp/pintuotuo-litellm.env"
      echo "  ssh ${OVERSEAS_SSH} 'sudo mkdir -p ${OVERSEAS_DIR} && sudo chown \$USER:\$USER ${OVERSEAS_DIR} && install -m 600 /tmp/pintuotuo-litellm.env ${OVERSEAS_DIR}/.env'"
    fi
    ;;
esac
