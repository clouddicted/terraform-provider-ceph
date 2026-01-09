resource "ceph_pool" "pool1" {
  name = "terraform_pool_1"

  quota_max_bytes = 10737418240 # Quota 10GB
  size            = 2           # Replicate into two destinations
}