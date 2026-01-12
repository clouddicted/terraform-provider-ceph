data "ceph_crush_rule" "default" {
  name = "replicated_rule"
}

output "rule_id" {
  value = data.ceph_crush_rule.default.rule_id
}
