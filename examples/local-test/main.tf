terraform {
  required_providers {
    ceph = {
      source = "clouddicted/ceph"
    }
  }
}

variable "ceph_url" {
  type = string
}

variable "ceph_username" {
  type = string
}

variable "ceph_password" {
  type      = string
  sensitive = true
}

variable "ceph_insecure" {
  type    = bool
  default = false
}

provider "ceph" {
  url      = var.ceph_url
  username = var.ceph_username
  password = var.ceph_password
  insecure = var.ceph_insecure
}

resource "ceph_pool" "test" {
  name   = "terraform_test_pool"
  pg_num = 16
  type   = "replicated"
}

resource "ceph_user" "test_user" {
  name  = "client.terraform_test_user"
  pools = [ceph_pool.test.name]
}

output "user_key" {
  value     = ceph_user.test_user.key
  sensitive = true
}

data "ceph_pool" "test_ds" {
  name = ceph_pool.test.name
}

data "ceph_user" "test_user_ds" {
  name = ceph_user.test_user.name
}

output "pool_ds_pg_num" {
  value = data.ceph_pool.test_ds.pg_num
}

output "user_ds_key" {
  value     = data.ceph_user.test_user_ds.key
  sensitive = true
}

data "ceph_cluster" "main" {}

output "cluster_fsid" {
  value = data.ceph_cluster.main.fsid
}

data "ceph_monitors" "main" {}

output "monitors" {
  value = data.ceph_monitors.main.monitors
}
