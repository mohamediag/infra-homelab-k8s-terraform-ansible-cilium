# ==============================================================================
# vms.tf — Instantiate VM module for each defined VM
# ==============================================================================
# Loops over var.vm_definitions and creates one KVM VM per entry.
# All VM details (sizing, IP, role) come from the vm_definitions variable.
# ==============================================================================

module "vms" {
  source = "./modules/vm"

  for_each = var.vm_definitions

  # Identity
  name = each.key

  # Compute resources
  vcpus  = each.value.vcpus
  memory = each.value.memory

  # Storage
  disk_size    = each.value.disk
  base_volume  = libvirt_volume.ubuntu_base.id
  storage_pool = libvirt_pool.images.name

  # Networking
  network_id  = libvirt_network.k8s_net.id
  ip_address  = each.value.ip
  gateway     = var.bridge_gateway
  dns_servers = var.dns_servers

  # SSH access — public key injected via cloud-init
  ssh_public_key = tls_private_key.vm_key.public_key_openssh
}
