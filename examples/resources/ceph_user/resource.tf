# Create a pool first
resource "ceph_pool" "kubernetes" {
  name                 = "kubernetes-rbd"
  pg_num               = 64
  type                 = "replicated"
  application_metadata = ["rbd"]
}

# User with access to single pool
resource "ceph_user" "csi_user" {
  name  = "client.kubernetes-csi"
  pools = [ceph_pool.kubernetes.name]
}

# Output the key for CSI configuration
output "csi_user_key" {
  value     = ceph_user.csi_user.key
  sensitive = true
}

# User with access to multiple pools
resource "ceph_pool" "backup" {
  name                 = "kubernetes-backup"
  pg_num               = 32
  type                 = "replicated"
  application_metadata = ["rbd"]
}

resource "ceph_user" "multi_pool_user" {
  name  = "client.multi-pool-access"
  pools = [ceph_pool.kubernetes.name, ceph_pool.backup.name]
}