# ==============================================================================
# modules/vm/main.tf — KVM VM module
# ==============================================================================
# Creates a single KVM virtual machine with:
#   1. A qcow2 disk volume (copy-on-write clone of the Ubuntu base image)
#   2. A cloud-init disk (user-data + network-config)
#   3. A libvirt domain (the actual VM)
#
# The VM gets a static IP, kubeadmin user, and SSH key via cloud-init.
# ==============================================================================

terraform {
  required_providers {
    libvirt = {
      source = "dmacvicar/libvirt"
    }
  }
}

# ── Disk volume ───────────────────────────────────────────────────────────────

# Each VM gets its own qcow2 volume, backed by the Ubuntu base image.
# copy-on-write means the base image is shared; only changes are stored per-VM.
resource "libvirt_volume" "vm_disk" {
  name           = "${var.name}.qcow2"
  pool           = var.storage_pool
  base_volume_id = var.base_volume
  format         = "qcow2"
  size           = var.disk_size # Resize to desired disk size
}

# ── Cloud-init configuration ──────────────────────────────────────────────────

# Create the cloud-init disk (a small ISO attached to the VM)
# cloud-init reads this on first boot to configure the VM.
# We use the built-in templatefile() function (no external provider needed).
resource "libvirt_cloudinit_disk" "cloudinit" {
  name = "${var.name}-cloudinit.iso"
  pool = var.storage_pool

  user_data = templatefile("${path.module}/templates/cloud_init.cfg.tpl", {
    hostname       = var.name
    ssh_public_key = var.ssh_public_key
  })

  network_config = templatefile("${path.module}/templates/network_config.cfg.tpl", {
    ip_address  = var.ip_address
    gateway     = var.gateway
    dns_servers = join(", ", var.dns_servers)
  })
}

# ── Libvirt domain (the VM itself) ────────────────────────────────────────────

resource "libvirt_domain" "vm" {
  name   = var.name
  vcpu   = var.vcpus
  memory = var.memory # In MB

  # CPU settings — host-passthrough exposes all host CPU features to the VM.
  # This is required for Cilium's eBPF operations which need certain CPU instructions.
  cpu {
    mode = "host-passthrough"
  }

  # Attach the cloud-init disk (provides first-boot configuration)
  cloudinit = libvirt_cloudinit_disk.cloudinit.id

  # Main disk — the qcow2 volume backed by the Ubuntu cloud image
  disk {
    volume_id = libvirt_volume.vm_disk.id
  }

  # Network interface — connect to the k8s-net NAT bridge
  network_interface {
    network_id     = var.network_id
    wait_for_lease = false # Static IPs via cloud-init; don't wait for DHCP
  }

  # Serial console — accessible via `virsh console <name>` on the Hetzner host.
  # Useful for debugging cloud-init issues when SSH isn't up yet.
  console {
    type        = "pty"
    target_type = "serial"
    target_port = "0"
  }

  # Ensure VM starts automatically when libvirtd restarts (e.g. after host reboot)
  autostart = true
}
