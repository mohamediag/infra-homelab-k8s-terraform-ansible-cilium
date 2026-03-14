# ==============================================================================
# modules/vm/outputs.tf — VM module outputs
# ==============================================================================

output "domain_id" {
  description = "Libvirt domain ID for this VM"
  value       = libvirt_domain.vm.id
}

output "ip_address" {
  description = "Static IP address of this VM"
  value       = var.ip_address
}

output "name" {
  description = "Name of this VM"
  value       = var.name
}
