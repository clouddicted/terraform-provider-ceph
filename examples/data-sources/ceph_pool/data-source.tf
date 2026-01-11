data "ceph_pool" "example" {
  name = "my-pool"
}

output "pool_pg_num" {
  value = data.ceph_pool.example.pg_num
}

output "pool_size" {
  value = data.ceph_pool.example.size
}
