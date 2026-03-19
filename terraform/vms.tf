# ==============================================================================
# vms.tf — Instantiate VM module for each defined VM
# ==============================================================================

module "vms" {
  source = "./modules/vm"

  for_each = var.vm_definitions

  name = each.key

  vcpus  = each.value.vcpus
  memory = each.value.memory

  disk_size      = each.value.disk
  base_volume_id = libvirt_volume.ubuntu_base.id
  storage_pool   = libvirt_pool.images.name

  network_id  = libvirt_network.k8s_net.id
  ip_address  = each.value.ip
  gateway     = var.bridge_gateway
  dns_servers = var.dns_servers

  ssh_public_key = tls_private_key.vm_key.public_key_openssh
}
