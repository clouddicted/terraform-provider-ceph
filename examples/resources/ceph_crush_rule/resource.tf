# CRUSH rule for HDD devices with host-level failure domain
resource "ceph_crush_rule" "hdd_hosts" {
  name           = "hdd_hosts_rule"
  root           = "default"
  failure_domain = "host"
  device_class   = "hdd"
}

# CRUSH rule for SSD devices
resource "ceph_crush_rule" "ssd_hosts" {
  name           = "ssd_hosts_rule"
  root           = "default"
  failure_domain = "host"
  device_class   = "ssd"
}

# CRUSH rule for single-node cluster (OSD failure domain)
resource "ceph_crush_rule" "single_node" {
  name           = "single_node_rule"
  root           = "default"
  failure_domain = "osd"
}

# Use CRUSH rule with a pool
resource "ceph_pool" "fast_pool" {
  name      = "fast-storage"
  pg_num    = 64
  type      = "replicated"
  rule_name = ceph_crush_rule.ssd_hosts.name
}
