data "ceph_user" "example" {
  name = "client.myuser"
}

output "user_pools" {
  value = data.ceph_user.example.pools
}

output "user_key" {
  value     = data.ceph_user.example.key
  sensitive = true
}
