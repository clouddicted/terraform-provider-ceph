#!/bin/bash
set -e

# Get the absolute path of the project root
# Assuming the script is run from the project root or examples/local-test
if [ -f "go.mod" ]; then
    PROJECT_ROOT=$(pwd)
    TEST_DIR="$PROJECT_ROOT/examples/local-test"
elif [ -f "../../go.mod" ]; then
    PROJECT_ROOT=$(cd ../.. && pwd)
    TEST_DIR=$(pwd)
else
    echo "Error: Could not find project root (go.mod not found)"
    exit 1
fi

echo "Building provider..."
cd "$PROJECT_ROOT"
go build -o terraform-provider-ceph

echo "Generating dev_overrides.tfrc..."
cat <<EOF > "$TEST_DIR/dev_overrides.tfrc"
provider_installation {
  dev_overrides {
    "clouddicted/ceph" = "$PROJECT_ROOT"
  }
  direct {}
}
EOF

echo "Loading environment variables from .env..."
if [ -f "$TEST_DIR/.env" ]; then
  export $(cat "$TEST_DIR/.env" | xargs)
else
  echo "Error: .env file not found in $TEST_DIR. Please copy .env.example to .env and configure it."
  exit 1
fi

echo "Running Terraform..."
cd "$TEST_DIR"
export TF_CLI_CONFIG_FILE="$TEST_DIR/dev_overrides.tfrc"

# Clean up previous runs
rm -rf .terraform .terraform.lock.hcl

echo "Skipping terraform init (not needed with dev_overrides for local backend)..."
terraform plan
