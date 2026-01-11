data "ceph_monitors" "example" {}

output "monitors" {
  value = data.ceph_monitors.example.monitors
}

output "monitor_addresses" {
  value = [for m in data.ceph_monitors.example.monitors : m.addr]
}
