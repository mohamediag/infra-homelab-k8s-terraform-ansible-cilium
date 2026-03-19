# ==============================================================================
# Homelab K8s Platform Makefile
# ==============================================================================
# Everything is driven from your local machine over SSH.
# Order of operations: check-local → host-prepare → infra-init → infra → cluster
#
# Required env vars: HETZNER_HOST_IP, HETZNER_SSH_KEY
# Copy .envrc.example → .envrc and source it (or use direnv).
# ==============================================================================

HETZNER_IP  ?= $(shell echo $$HETZNER_HOST_IP)
SSH_KEY     ?= $(shell echo $$HETZNER_SSH_KEY)
PUBLIC_IP   ?= $(shell echo $$MY_PUBLIC_IP)
ANSIBLE_DIR  = ansible
TERRAFORM_DIR = terraform

# Default target: show help
.DEFAULT_GOAL := help

# ── Local verification ─────────────────────────────────────────────────────────

.PHONY: check-local
check-local: ## Verify local prerequisites (tools, env vars, SSH connectivity)
	@bash scripts/setup-local.sh

# ── Host preparation (MUST run before terraform apply) ────────────────────────

.PHONY: host-prepare
host-prepare: ## Prepare Hetzner host: install KVM/libvirt, configure networking
	cd $(ANSIBLE_DIR) && ansible-playbook playbooks/00-host-prepare.yml \
		-i "$(HETZNER_IP)," \
		-e "ansible_user=root ansible_ssh_private_key_file=$(SSH_KEY) ansible_ssh_common_args='-o StrictHostKeyChecking=no' my_public_ip=$(PUBLIC_IP)"

# ── Terraform (VM provisioning) ───────────────────────────────────────────────

.PHONY: infra-init
infra-init: ## Terraform init (download providers)
	cd $(TERRAFORM_DIR) && terraform init

.PHONY: infra-plan
infra-plan: ## Terraform plan (preview changes)
	cd $(TERRAFORM_DIR) && terraform plan

.PHONY: infra
infra: ## Terraform apply — create KVM VMs on Hetzner host
	cd $(TERRAFORM_DIR) && terraform apply -auto-approve

.PHONY: infra-destroy
infra-destroy: ## Terraform destroy — remove all VMs
	cd $(TERRAFORM_DIR) && terraform destroy -auto-approve

# ── Kubernetes cluster bootstrap ──────────────────────────────────────────────

.PHONY: cluster
cluster: ## Bootstrap K8s cluster (all Ansible playbooks 01-07)
	cd $(ANSIBLE_DIR) && ansible-playbook playbooks/site.yml

.PHONY: fetch-kubeconfig
fetch-kubeconfig: ## Fetch kubeconfig from cp-01 to ./kubeconfig-local.yml
	cd $(ANSIBLE_DIR) && ansible-playbook playbooks/99-fetch-kubeconfig.yml

# ── kubectl access via SSH tunnel ────────────────────────────────────────────

.PHONY: tunnel
tunnel: ## Open SSH tunnel: localhost:6443 → cp-01:6443 (for kubectl)
	ssh -L 6443:10.0.0.10:6443 root@$(HETZNER_IP) -N -f -i $(SSH_KEY)
	@echo ""
	@echo "Tunnel is open. Use kubectl with:"
	@echo "  export KUBECONFIG=./kubeconfig-local.yml"
	@echo "  kubectl get nodes"

# ── Status ───────────────────────────────────────────────────────────────────

.PHONY: status
status: ## Show cluster node and pod status (requires open tunnel + kubeconfig)
	KUBECONFIG=./kubeconfig-local.yml kubectl get nodes -o wide
	KUBECONFIG=./kubeconfig-local.yml kubectl get pods -A

# ── SSH shortcuts ─────────────────────────────────────────────────────────────

.PHONY: ssh-host
ssh-host: ## SSH into Hetzner host as root
	ssh -i $(SSH_KEY) root@$(HETZNER_IP)

.PHONY: ssh-cp
ssh-cp: ## SSH into control plane (cp-01) via bastion
	ssh -J root@$(HETZNER_IP) -i terraform/files/vm_key kubeadmin@10.0.0.10

.PHONY: ssh-w1
ssh-w1: ## SSH into worker-01 via bastion
	ssh -J root@$(HETZNER_IP) -i terraform/files/vm_key kubeadmin@10.0.0.11

.PHONY: ssh-w2
ssh-w2: ## SSH into worker-02 via bastion
	ssh -J root@$(HETZNER_IP) -i terraform/files/vm_key kubeadmin@10.0.0.12

# ── Cluster management ────────────────────────────────────────────────────────

.PHONY: reset-cluster
reset-cluster: ## Reset kubeadm on all nodes (VMs remain, cluster wiped)
	cd $(ANSIBLE_DIR) && ansible k8s_nodes -m shell -a "kubeadm reset -f" --become

# ── Full lifecycle ────────────────────────────────────────────────────────────

.PHONY: all
all: host-prepare infra cluster fetch-kubeconfig ## Full deployment: host prep → VMs → cluster → kubeconfig
	@echo ""
	@echo "Deployment complete! Run 'make tunnel' then 'make status'."

.PHONY: nuke
nuke: infra-destroy ## Destroy all VMs (keeps host KVM setup)

# ── Help ──────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help message
	@echo ""
	@echo "Homelab K8s Platform"
	@echo "===================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Required env vars: HETZNER_HOST_IP, HETZNER_SSH_KEY"
	@echo "Quickstart: make check-local && make all && make tunnel && make status"
