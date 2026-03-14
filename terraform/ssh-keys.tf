# ==============================================================================
# ssh-keys.tf — VM SSH key generation
# ==============================================================================
# Generates a dedicated ED25519 SSH key pair for accessing the KVM VMs.
# The private key is written to terraform/files/vm_key (gitignored, chmod 600).
# The public key is injected into each VM via cloud-init.
#
# This key is SEPARATE from your Hetzner host key (HETZNER_SSH_KEY).
# Use: ssh -J root@<host> -i terraform/files/vm_key kubeadmin@<vm-ip>
# ==============================================================================

# Generate an ED25519 key pair specifically for VM access
resource "tls_private_key" "vm_key" {
  algorithm = "ED25519"
}

# Write the private key to a local file (gitignored)
# This file is used in SSH commands and Ansible inventory
resource "local_sensitive_file" "vm_private_key" {
  content         = tls_private_key.vm_key.private_key_openssh
  filename        = "${path.module}/files/vm_key"
  file_permission = "0600" # Strict permissions required by SSH
}

# Write the public key as well for reference
resource "local_file" "vm_public_key" {
  content         = tls_private_key.vm_key.public_key_openssh
  filename        = "${path.module}/files/vm_key.pub"
  file_permission = "0644"
}
