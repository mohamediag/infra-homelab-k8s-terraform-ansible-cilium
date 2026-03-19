#cloud-config
# ==============================================================================
# cloud_init.cfg.tpl — Cloud-init user-data template
# ==============================================================================
# This configures each VM on first boot:
# - Sets the hostname
# - Creates the kubeadmin user with sudo access and SSH key
# - Installs basic packages
# - Disables password authentication
# - Sets timezone to UTC
#
# Template variables (injected by Terraform):
#   hostname        — VM name (e.g. cp-01)
#   ssh_public_key  — ED25519 public key for kubeadmin access
# ==============================================================================

hostname: ${hostname}
fqdn: ${hostname}.k8s.local
manage_etc_hosts: true

# Create the kubeadmin user
# This user is used by Ansible and for manual access to nodes
users:
  - name: kubeadmin
    gecos: "Kubernetes Admin"
    shell: /bin/bash
    sudo: ALL=(ALL) NOPASSWD:ALL
    lock_passwd: true  # No password login — SSH key only
    ssh_authorized_keys:
      - ${ssh_public_key}

# Disable root SSH password login (key-based only)
disable_root: false
ssh_pwauth: false

# Install useful packages on first boot
packages:
  - curl
  - wget
  - vim
  - git
  - htop
  - net-tools
  - jq
  - apt-transport-https
  - ca-certificates
  - gnupg
  - lsb-release

package_update: true
package_upgrade: true

# Set timezone to UTC for consistent log timestamps across cluster
timezone: UTC

# Final message logged to /var/log/cloud-init-output.log
final_message: |
  Cloud-init complete for ${hostname}.
  kubeadmin user created with SSH key access.
  Ready for Ansible configuration.
