# ==============================================================================
# storage.tf — Libvirt storage pool and base OS image
# ==============================================================================
# Creates a storage pool on the Hetzner host at /var/lib/libvirt/images/.
# Downloads the Ubuntu 22.04 cloud image as the base volume.
# Each VM's disk is a qcow2 volume backed (copy-on-write) from this base.
#
# NOTE: The base image must already be downloaded to ubuntu_image_path on the
# Hetzner host before running terraform apply. The host-prepare Ansible role
# handles this download.
# ==============================================================================

# Storage pool — the directory where all VM disk images live on the host
resource "libvirt_pool" "images" {
  name = "k8s-images"
  type = "dir"

  target {
    path = "/var/lib/libvirt/images"
  }
}

# Base volume — the Ubuntu 22.04 cloud image, shared as a backing store.
# All VM volumes are copy-on-write clones of this base, saving disk space.
# The image is fetched from the host's local path (already downloaded by Ansible).
resource "libvirt_volume" "ubuntu_base" {
  name   = "ubuntu-22.04-base.qcow2"
  pool   = libvirt_pool.images.name
  source = var.ubuntu_image_path
  format = "qcow2"
}
