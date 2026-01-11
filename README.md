# Terraform Provider for Ceph

A Terraform provider for managing Ceph clusters via the Ceph Dashboard API. Designed to provision storage resources and export configuration data for ceph-csi integration.

> **Important:** This provider communicates with Ceph exclusively through the Dashboard REST API. The Ceph Dashboard manager module must be enabled and accessible.

## Use Case

- Create and manage Ceph pools and users
- Export cluster FSID, monitor addresses, and user keys for ceph-csi configuration
- Automate Kubernetes storage provisioning with Ceph RBD

## Requirements

- Terraform >= 1.0
- Go >= 1.24 (for building)
- Ceph cluster with Dashboard manager module enabled

## Ceph Dashboard Setup

The Dashboard manager module must be enabled for this provider to work.

```shell
# Enable the dashboard module
ceph mgr module enable dashboard

# Create a self-signed certificate (or use your own)
ceph dashboard create-self-signed-cert

# Set the dashboard port (optional, default is 8443)
ceph config set mgr mgr/dashboard/server_port 8443

# Create an admin user
ceph dashboard ac-user-create admin -i <password-file> administrator

# Verify dashboard is running
ceph mgr services
```

The `ceph mgr services` command should show the dashboard URL:

```json
{
    "dashboard": "https://<host>:8443/"
}
```

## Installation

```hcl
terraform {
  required_providers {
    ceph = {
      source  = "clouddicted/ceph"
      version = "~> 0.1"
    }
  }
}
```

## Provider Configuration

```hcl
provider "ceph" {
  url      = "https://ceph-dashboard.example.com:8443"
  username = "admin"
  password = "your-password"
  insecure = true  # Skip TLS verification
}
```

## Implemented

### Resources

| Resource | Description |
|----------|-------------|
| `ceph_pool` | Create/update/delete pools (replicated). Supports pg_num, size, quotas, application_metadata.|
| `ceph_user` | Create/update/delete users with RBD access to specified pools. Exports the user key. |

### Data Sources

| Data Source | Description |
|-------------|-------------|
| `ceph_cluster` | Read cluster FSID |
| `ceph_monitors` | Read monitor addresses (name, addr, rank) |
| `ceph_pool` | Read pool configuration |
| `ceph_user` | Read user pools and key |

## Not (and probably never) Implemented

- Managmnent other resources than pools (RBD type only), users (keys)

## Example: ceph-csi Configuration

```hcl
resource "ceph_pool" "kubernetes" {
  name                 = "kubernetes-rbd"
  pg_num               = 64
  type                 = "replicated"
  application_metadata = ["rbd"]
}

resource "ceph_user" "csi" {
  name  = "client.kubernetes-csi"
  pools = [ceph_pool.kubernetes.name]
}

data "ceph_cluster" "main" {}
data "ceph_monitors" "main" {}

# Values for ceph-csi ConfigMap and Secret
output "cluster_id" {
  value = data.ceph_cluster.main.fsid
}

output "monitors" {
  value = join(",", [for m in data.ceph_monitors.main.monitors : m.addr])
}

output "user_key" {
  value     = ceph_user.csi.key
  sensitive = true
}
```

## Development

```shell
# Build
go build -o terraform-provider-ceph

# Test locally
cd examples/local-test
cp .env.example .env  # Configure credentials
./run_test.sh
```

## License

MPL-2.0
