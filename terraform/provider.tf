# ==============================================================================
# provider.tf — Terraform provider configuration
# ==============================================================================
# The libvirt provider connects to the remote Hetzner host over SSH,
# so Terraform runs locally but manages KVM VMs on the remote server.
#
# Provider docs: https://registry.terraform.io/providers/dmacvicar/libvirt
# ==============================================================================

terraform {
  required_version = ">= 1.5"

  required_providers {
    # Manages KVM/QEMU VMs on the remote Hetzner host via SSH
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "~> 0.8.0"
    }

    # Generates ED25519 SSH key pair for VM access
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }

    # Writes the generated private key to a local file
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
  }
}

# Connect to libvirt daemon on the Hetzner host over SSH.
# The URI format:
#   qemu+ssh://<user>@<host>/system  — connects to system-level libvirt
#   sshauth=privkey                   — use private key authentication
#   keyfile=<path>                    — path to your local SSH private key
#
# NOTE: The Hetzner host must have libvirtd running (make host-prepare).
provider "libvirt" {
  uri = "qemu+ssh://root@${var.hetzner_host_ip}/system?sshauth=privkey&keyfile=${var.ssh_key_path}"
}

provider "tls" {}

provider "local" {}
