#!/usr/bin/env bash
# ==============================================================================
# setup-local.sh — Local prerequisites check for homelab K8s platform
# ==============================================================================
# Verifies that all required tools are installed, env vars are set,
# and SSH connectivity to the Hetzner host works.
# Run via: make check-local
# ==============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS=0
FAIL=0

check_tool() {
  local tool="$1"
  if command -v "$tool" &>/dev/null; then
    echo -e "  ${GREEN}✓${NC} $tool ($(command -v "$tool"))"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}✗${NC} $tool — NOT FOUND (install it first)"
    FAIL=$((FAIL + 1))
  fi
}

check_env() {
  local var="$1"
  local val="${!var:-}"
  if [[ -n "$val" ]]; then
    echo -e "  ${GREEN}✓${NC} $var = $val"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}✗${NC} $var — NOT SET"
    FAIL=$((FAIL + 1))
  fi
}

echo ""
echo "=================================================="
echo " Homelab K8s Platform — Local Prerequisites Check"
echo "=================================================="
echo ""

# ── Required tools ────────────────────────────────────────────────────────────
echo "Required tools:"
check_tool terraform
check_tool ansible
check_tool ansible-playbook
check_tool ssh
check_tool make
echo ""

# ── Optional but useful ───────────────────────────────────────────────────────
echo "Optional tools:"
check_tool kubectl  || true
check_tool direnv   || true
echo ""

# ── Environment variables ─────────────────────────────────────────────────────
echo "Environment variables:"
check_env HETZNER_HOST_IP
check_env HETZNER_SSH_KEY
echo ""

# ── SSH key exists ────────────────────────────────────────────────────────────
SSH_KEY="${HETZNER_SSH_KEY:-}"
if [[ -n "$SSH_KEY" ]]; then
  if [[ -f "$SSH_KEY" ]]; then
    echo -e "  ${GREEN}✓${NC} SSH key file exists: $SSH_KEY"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}✗${NC} SSH key file NOT found: $SSH_KEY"
    FAIL=$((FAIL + 1))
  fi
fi
echo ""

# ── SSH connectivity test ─────────────────────────────────────────────────────
HOST_IP="${HETZNER_HOST_IP:-}"
if [[ -n "$HOST_IP" && -n "$SSH_KEY" && -f "$SSH_KEY" ]]; then
  echo "SSH connectivity test (root@$HOST_IP):"
  if ssh -i "$SSH_KEY" \
       -o ConnectTimeout=5 \
       -o StrictHostKeyChecking=accept-new \
       -o BatchMode=yes \
       "root@$HOST_IP" "echo 'SSH OK'" &>/dev/null; then
    echo -e "  ${GREEN}✓${NC} SSH to root@$HOST_IP — OK"
    PASS=$((PASS + 1))
  else
    echo -e "  ${RED}✗${NC} SSH to root@$HOST_IP — FAILED"
    echo -e "     Check that your key is authorized on the server."
    FAIL=$((FAIL + 1))
  fi
else
  echo -e "  ${YELLOW}⚠${NC}  Skipping SSH test (HETZNER_HOST_IP or HETZNER_SSH_KEY not set)"
fi
echo ""

# ── Summary ───────────────────────────────────────────────────────────────────
echo "=================================================="
echo " Summary"
echo "=================================================="
echo -e "  ${GREEN}Passed${NC}: $PASS"
echo -e "  ${RED}Failed${NC}: $FAIL"
echo ""

if [[ $FAIL -gt 0 ]]; then
  echo -e "${RED}Fix the issues above before proceeding.${NC}"
  echo ""
  echo "Next steps:"
  echo "  1. Install missing tools"
  echo "  2. Copy .envrc.example → .envrc and set your values"
  echo "  3. Source .envrc (or use direnv)"
  echo "  4. Re-run: make check-local"
  exit 1
else
  echo -e "${GREEN}All checks passed! You're ready to deploy.${NC}"
  echo ""
  echo "Next steps:"
  echo "  make host-prepare  # Prepare Hetzner host (KVM, libvirt)"
  echo "  make infra-init    # Init Terraform"
  echo "  make infra         # Create VMs"
  echo "  make cluster       # Bootstrap Kubernetes"
  echo "  make fetch-kubeconfig"
  echo "  make tunnel"
  echo "  make status"
fi
