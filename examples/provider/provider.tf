terraform {
  required_providers {
    ceph = {
      source  = "clouddicted/ceph"
      version = "~> 0.1"
    }
  }
}

# Configure the Ceph provider
provider "ceph" {
  url      = "https://ceph-dashboard.example.com:8443"
  username = "admin"
  password = "your-password"
  insecure = false  # Set to true to skip TLS verification
}

# Using variables (recommended for production)
variable "ceph_url" {
  type        = string
  description = "Ceph Dashboard URL"
}

variable "ceph_username" {
  type        = string
  description = "Ceph Dashboard username"
}

variable "ceph_password" {
  type        = string
  sensitive   = true
  description = "Ceph Dashboard password"
}

provider "ceph" {
  alias    = "with_vars"
  url      = var.ceph_url
  username = var.ceph_username
  password = var.ceph_password
  insecure = true
}