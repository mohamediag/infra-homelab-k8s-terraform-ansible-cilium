# Homelab K8s Platform

> **Work in progress** — actively being built and tested.

Production-grade Kubernetes cluster on a single Hetzner EX44 bare-metal server, running 3 KVM virtual machines. Fully automated: Terraform provisions the VMs, Ansible bootstraps Kubernetes — all driven from your local machine over SSH.

```
[Your Laptop] ─SSH──► [Hetzner EX44: PUBLIC_IP]
                              │
                              │ bridge: br0 / k8s-net (10.0.0.0/24)
                              │ host gateway: 10.0.0.1
                              │
              ┌───────────────┼───────────────┐
              │               │               │
         [cp-01]        [worker-01]      [worker-02]
         10.0.0.10       10.0.0.11        10.0.0.12
         4 vCPU          6 vCPU           6 vCPU
         8 GB RAM        20 GB RAM        20 GB RAM
         80 GB disk      120 GB disk      120 GB disk
```

**Stack:**
- **KVM/QEMU** — virtualization on bare metal
- **Terraform** (`dmacvicar/libvirt`) — VM provisioning over SSH
- **Ansible** — OS configuration + Kubernetes bootstrap
- **kubeadm** — Kubernetes cluster initialization
- **Cilium** — CNI with kube-proxy replacement (eBPF mode)
- **Hubble** — Network observability UI
- **MetalLB** — LoadBalancer IPs for bare metal
- **cert-manager** — Automatic TLS certificates
- **Nginx Ingress** — HTTP/HTTPS routing

---

## Prerequisites

### Local Machine
- `terraform` >= 1.5
- `ansible` >= 2.14
- `ssh` client
- `make`
- `kubectl` (optional, for direct cluster access)

### Hetzner Host
- Ubuntu 22.04 LTS (fresh install)
- SSH access as `root`
- Hardware virtualization enabled (Intel VT-x / AMD-V in BIOS)

---

## Environment Setup

```bash
# Clone or navigate to this directory
cd infra-homelab-k8s-terraform-ansible-cilium/

# Copy the env example and fill in your values
cp .envrc.example .envrc
# Edit .envrc:
#   HETZNER_HOST_IP="your.server.ip"
#   HETZNER_SSH_KEY="$HOME/.ssh/id_ed25519"

# Source it (or use direnv: direnv allow)
source .envrc

# Copy the Terraform vars example
cp terraform/terraform.tfvars.example terraform/terraform.tfvars
# Edit terraform.tfvars with your server IP and SSH key path
```

---

## Quick Start

```bash
make check-local    # Verify all prerequisites
make all            # Full deployment: host prep → VMs → cluster → kubeconfig
make tunnel         # Open SSH tunnel for kubectl
make status         # Show cluster node + pod status
```

---

## Step-by-Step Deployment

### 1. Check local prerequisites
```bash
make check-local
```
This verifies terraform, ansible, ssh, make are installed, your env vars are set, and SSH to the Hetzner host works.

### 2. Prepare the Hetzner host
```bash
make host-prepare
```
Installs KVM/libvirt, configures iptables NAT for VM network, sets up UFW, and downloads the Ubuntu 22.04 cloud image. **This must run before terraform apply.**

### 3. Initialize and apply Terraform
```bash
make infra-init     # Download Terraform providers (~200 MB)
make infra-plan     # Preview what will be created
make infra          # Create the 3 KVM VMs
```
Terraform generates an ED25519 SSH key pair for VM access (`terraform/files/vm_key`) and writes the Ansible inventory to `ansible/inventory/hosts.yml`.

### 4. Bootstrap Kubernetes
```bash
make cluster
```
Runs all Ansible playbooks in order:
1. Common OS config (swap off, kernel modules, sysctl)
2. containerd installation and configuration
3. kubeadm/kubelet/kubectl installation
4. Control plane initialization (`kubeadm init`)
5. Worker nodes join the cluster
6. Cilium CNI installation
7. MetalLB, metrics-server, cert-manager, Nginx Ingress

### 5. Fetch kubeconfig
```bash
make fetch-kubeconfig
```
Downloads kubeconfig from cp-01, rewrites the server URL to `https://127.0.0.1:6443` for tunnel access.

---

## kubectl Access (via SSH Tunnel)

The Kubernetes API server is only accessible from inside the VM network. Use an SSH tunnel:

```bash
# Open the tunnel (runs in background)
make tunnel

# Now use kubectl
export KUBECONFIG=./kubeconfig-local.yml
kubectl get nodes
kubectl get pods -A
```

Or use the Makefile shortcut:
```bash
make status
```

The tunnel forwards `localhost:6443` → `cp-01:6443` through the Hetzner host.

---

## SSH Access to Nodes

```bash
make ssh-host    # SSH into Hetzner host as root
make ssh-cp      # SSH into cp-01 (control plane)
make ssh-w1      # SSH into worker-01
make ssh-w2      # SSH into worker-02
```

Manual SSH (with ProxyJump):
```bash
# To any VM via the Hetzner host as bastion
ssh -J root@$HETZNER_HOST_IP -i terraform/files/vm_key kubeadmin@10.0.0.10
```

---

## Tear Down and Rebuild

### Destroy all VMs (keeps KVM setup on host)
```bash
make nuke
```

### Full rebuild from VMs
```bash
make infra         # Recreate VMs
make cluster       # Re-bootstrap Kubernetes
make fetch-kubeconfig
```

### Reset Kubernetes (keep VMs)
```bash
make reset-cluster    # Runs kubeadm reset on all nodes
make cluster          # Re-bootstrap (skips common/containerd/kubeadm)
```

---

## Troubleshooting

### Terraform: libvirt SSH connection fails
```
Error: failed to connect to the hypervisor
```
- Verify `HETZNER_HOST_IP` and `HETZNER_SSH_KEY` env vars are set
- Test manually: `ssh -i $HETZNER_SSH_KEY root@$HETZNER_HOST_IP virsh list`
- Ensure `make host-prepare` has been run (libvirtd must be running)
- Check libvirt URI in `provider.tf` — the `keyfile=` path must be absolute or relative to where you run terraform

### Cloud-init not running / VM not getting IP
```bash
# Console access via the Hetzner host
ssh root@$HETZNER_HOST_IP
virsh console cp-01    # Ctrl+] to exit
```
Check `/var/log/cloud-init-output.log` inside the VM.

### kubeadm init fails
- Ensure containerd is running: `systemctl status containerd`
- Verify swap is off: `swapon --show` (should be empty)
- Check kernel modules: `lsmod | grep -E 'overlay|br_netfilter'`
- View kubelet logs: `journalctl -xu kubelet`

### Cilium not ready
```bash
# On cp-01
cilium status
cilium connectivity test
kubectl get pods -n kube-system
```
Common causes:
- kube-proxy still running (should have been skipped by kubeadm config)
- Wrong `k8sServiceHost` — must be cp-01's internal IP (`10.0.0.10`)

### MetalLB not assigning IPs
```bash
kubectl describe svc <service-name>
kubectl logs -n metallb-system deployment/controller
```
- Ensure the service type is `LoadBalancer`
- Verify the IP pool range doesn't conflict with existing hosts

### nodes stuck in NotReady
Usually means Cilium isn't fully up yet:
```bash
kubectl get pods -n kube-system
cilium status --wait
```

---

## Project Structure

```
.
├── Makefile                    # All automation targets
├── .envrc.example              # Environment variable template
├── .gitignore                  # Ignores secrets, state, generated files
├── scripts/
│   └── setup-local.sh          # Local prerequisite checker
├── terraform/
│   ├── provider.tf             # libvirt provider (over SSH)
│   ├── variables.tf            # Input variables
│   ├── terraform.tfvars.example
│   ├── outputs.tf
│   ├── ssh-keys.tf             # VM SSH key generation
│   ├── network.tf              # libvirt NAT network
│   ├── storage.tf              # Storage pool + base image
│   ├── vms.tf                  # VM instantiation
│   ├── ansible-inventory.tf    # Generates Ansible inventory
│   ├── files/                  # Generated keys (gitignored)
│   └── modules/vm/             # Reusable VM module
├── ansible/
│   ├── ansible.cfg
│   ├── group_vars/             # Variables per host group
│   ├── inventory/              # Generated by Terraform (gitignored)
│   ├── roles/
│   │   ├── host-prepare/       # KVM/libvirt setup on Hetzner host
│   │   ├── common/             # Base K8s node config
│   │   ├── containerd/         # Container runtime
│   │   ├── kubeadm/            # Kubernetes tooling
│   │   ├── control-plane/      # kubeadm init
│   │   ├── workers/            # kubeadm join
│   │   ├── cilium/             # CNI installation
│   │   └── cluster-essentials/ # MetalLB, cert-manager, etc.
│   └── playbooks/              # Ordered playbooks
└── docs/
    ├── SPEC-INFRA.md           # Infrastructure specification (read-only)
    └── PROJECT.md              # Architecture deep-dive
```
