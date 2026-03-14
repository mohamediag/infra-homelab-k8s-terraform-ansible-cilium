# ==============================================================================
# outputs.tf — Terraform output values
# ==============================================================================
# These are displayed after `terraform apply` and can be queried with
# `terraform output`. Useful for verifying VM IPs and getting SSH hints.
# ==============================================================================

output "vm_ips" {
  description = "Map of VM name → IP address"
  value = {
    for name, def in var.vm_definitions : name => def.ip
  }
}

output "ssh_connect_hints" {
  description = "SSH commands to connect to each VM via the Hetzner host bastion"
  value = {
    for name, def in var.vm_definitions :
    name => "ssh -J root@${var.hetzner_host_ip} -i terraform/files/vm_key kubeadmin@${def.ip}"
  }
}

output "kubectl_tunnel_command" {
  description = "Command to open SSH tunnel for kubectl access to the API server"
  value       = "ssh -L 6443:10.0.0.10:6443 root@${var.hetzner_host_ip} -N -f -i ${var.ssh_key_path}"
}

output "vm_ssh_public_key" {
  description = "Public key injected into all VMs (for reference)"
  value       = tls_private_key.vm_key.public_key_openssh
  sensitive   = false
}
