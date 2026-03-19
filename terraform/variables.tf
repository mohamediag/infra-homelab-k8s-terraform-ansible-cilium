# ==============================================================================
# variables.tf — Input variable declarations
# ==============================================================================
# Copy terraform.tfvars.example → terraform.tfvars and fill in your values.
# terraform.tfvars is gitignored (contains your server's public IP).
# ==============================================================================

variable "hetzner_host_ip" {
  description = "Public IP address of the Hetzner dedicated server (set via TF_VAR_hetzner_host_ip env var)"
  type        = string
}

variable "ssh_key_path" {
  description = "Path to the local SSH private key for the Hetzner host (set via TF_VAR_ssh_key_path env var)"
  type        = string
}

# Map of VM definitions. Each entry becomes a KVM VM.
# Keys are VM names (used as hostname and libvirt domain name).
variable "vm_definitions" {
  description = "Map of VM definitions: name → configuration"
  type = map(object({
    vcpus  = number # Number of virtual CPUs
    memory = number # RAM in MB (1024 = 1 GB)
    disk   = number # Disk size in bytes (use 1024^3 multiplication in tfvars)
    ip     = string # Static IP within the bridge subnet
    role   = string # "control_plane" or "worker" — used for inventory grouping
  }))

  default = {
    "cp-01" = {
      vcpus  = 4
      memory = 8192   # 8 GB
      disk   = 85899345920 # 80 GB
      ip     = "10.0.0.10"
      role   = "control_plane"
    }
    "worker-01" = {
      vcpus  = 6
      memory = 20480  # 20 GB
      disk   = 128849018880 # 120 GB
      ip     = "10.0.0.11"
      role   = "worker"
    }
    "worker-02" = {
      vcpus  = 6
      memory = 20480  # 20 GB
      disk   = 128849018880 # 120 GB
      ip     = "10.0.0.12"
      role   = "worker"
    }
  }
}

# ── Network configuration ─────────────────────────────────────────────────────

variable "bridge_subnet" {
  description = "CIDR for the internal VM bridge network"
  type        = string
  default     = "10.0.0.0/24"
}

variable "bridge_gateway" {
  description = "Gateway IP for the internal VM network (Hetzner host's bridge IP)"
  type        = string
  default     = "10.0.0.1"
}

variable "dns_servers" {
  description = "DNS servers for VMs to use"
  type        = list(string)
  default     = ["8.8.8.8", "1.1.1.1"]
}

# ── OS image ──────────────────────────────────────────────────────────────────

variable "ubuntu_image_url" {
  description = "URL to Ubuntu 22.04 LTS cloud image (downloaded to Hetzner host)"
  type        = string
  default     = "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img"
}

variable "ubuntu_image_path" {
  description = "Path on the Hetzner host where the Ubuntu cloud image is stored"
  type        = string
  default     = "/var/lib/libvirt/images/ubuntu-22.04-cloudimg-amd64.img"
}
