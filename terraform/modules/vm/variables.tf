# ==============================================================================
# modules/vm/variables.tf — VM module input variables
# ==============================================================================

variable "name" {
  description = "VM hostname and libvirt domain name (e.g. cp-01, worker-01)"
  type        = string
}

variable "vcpus" {
  description = "Number of virtual CPUs"
  type        = number
}

variable "memory" {
  description = "RAM in MB (e.g. 8192 for 8 GB)"
  type        = number
}

variable "disk_size" {
  description = "Disk size in bytes (e.g. 85899345920 for 80 GB)"
  type        = number
}

variable "base_volume" {
  description = "ID of the libvirt base volume (Ubuntu cloud image) to clone from"
  type        = string
}

variable "storage_pool" {
  description = "Name of the libvirt storage pool to use"
  type        = string
}

variable "network_id" {
  description = "ID of the libvirt network to attach the VM to"
  type        = string
}

variable "ip_address" {
  description = "Static IP address for this VM (e.g. 10.0.0.10)"
  type        = string
}

variable "gateway" {
  description = "Default gateway IP (the Hetzner host bridge IP, 10.0.0.1)"
  type        = string
  default     = "10.0.0.1"
}

variable "dns_servers" {
  description = "List of DNS servers for the VM"
  type        = list(string)
  default     = ["8.8.8.8", "1.1.1.1"]
}

variable "ssh_public_key" {
  description = "SSH public key to inject into the kubeadmin user via cloud-init"
  type        = string
}
