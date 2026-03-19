# ==============================================================================
# network.tf — Libvirt NAT network for VMs
# ==============================================================================
# Creates a private NAT network (10.0.0.0/24) for the VMs.
# DHCP is disabled — VMs get static IPs via cloud-init network config.
# The Hetzner host acts as the gateway (10.0.0.1) and NATs VM traffic
# through its public IP for internet access.
# ==============================================================================

resource "libvirt_network" "k8s_net" {
  name      = "k8s-net"
  mode      = "nat"
  domain    = "k8s.local"
  addresses = [var.bridge_subnet]
  autostart = true

  dhcp {
    enabled = false
  }
}
