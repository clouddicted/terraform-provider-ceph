# Basic replicated pool
resource "ceph_pool" "basic" {
  name   = "my-replicated-pool"
  pg_num = 32
  type   = "replicated"
}

# Pool with quotas and replication settings
resource "ceph_pool" "with_quota" {
  name            = "my-quota-pool"
  pg_num          = 64
  type            = "replicated"
  size            = 3              # 3-way replication
  quota_max_bytes = 10737418240    # 10GB quota
}

# Pool for RBD images
resource "ceph_pool" "rbd_pool" {
  name                 = "kubernetes-rbd"
  pg_num               = 128
  type                 = "replicated"
  application_metadata = ["rbd"]
}