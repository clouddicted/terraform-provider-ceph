data "ceph_cluster" "example" {}

output "cluster_fsid" {
  value = data.ceph_cluster.example.fsid
}
