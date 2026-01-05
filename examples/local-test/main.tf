terraform {
  required_providers {
    ceph = {
      source = "clouddicted/ceph"
    }
  }
}

provider "ceph" {
  # Credentials will be loaded from environment variables:
  # CEPH_DASHBOARD_URL
  # CEPH_DASHBOARD_USERNAME
  # CEPH_DASHBOARD_PASSWORD
  # CEPH_DASHBOARD_INSECURE
}

resource "ceph_pool" "test" {
  name   = "terraform_test_pool"
  pg_num = 16
  type   = "replicated"
}

resource "ceph_user" "test_user" {
  name         = "client.terraform_test_user"
  capabilities = "mon 'allow r' osd 'allow *'"
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
