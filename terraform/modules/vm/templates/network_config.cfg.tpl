# ==============================================================================
# network_config.cfg.tpl — Cloud-init network configuration template
# ==============================================================================
# Configures static networking for the VM.
# DHCP is disabled on the libvirt network, so static config is required.
#
# Template variables (injected by Terraform):
#   ip_address   — Static IP for this VM (e.g. 10.0.0.10)
#   gateway      — Default gateway (Hetzner host bridge: 10.0.0.1)
#   dns_servers  — Comma-separated DNS server IPs
# ==============================================================================

version: 2
ethernets:
  # The first ethernet interface (virtio NIC attached to k8s-net bridge)
  # i440fx machine type + PCI slot 0x03 → Ubuntu names this "ens3"
  ens3:
    dhcp4: false
    addresses:
      - ${ip_address}/24
    gateway4: ${gateway}
    nameservers:
      addresses: [${dns_servers}]
