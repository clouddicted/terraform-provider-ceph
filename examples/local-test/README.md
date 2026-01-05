# Local Test Environment

This directory contains a Terraform configuration to test the `ceph` provider locally against an external Ceph cluster.

## Prerequisites

- Terraform
- Go
- Access to a Ceph Cluster with Dashboard enabled

## Setup

1.  **Configure Environment:**
    Copy `.env.example` to `.env` and update the values with your Ceph cluster details:
    ```bash
    cp .env.example .env
    vim .env
    ```

2.  **Run the Test:**
    Execute the test script, which will build the provider and run `terraform plan`:
    ```bash
    ./run_test.sh
    ```

    This script will:
    - Build the provider binary from source.
    - Configure Terraform to use the local binary (via `dev_overrides`).
    - Load credentials from `.env`.
    - Run `terraform plan`.

## Cleanup

To remove the Terraform state and lock files:
```bash
rm -rf .terraform .terraform.lock.hcl terraform.tfstate*
```
