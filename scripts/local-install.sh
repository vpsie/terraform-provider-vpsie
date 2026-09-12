#!/usr/bin/env bash
# Build the provider and install it into the local Terraform plugin cache so it
# can be used via a dev_override, without publishing to a registry.
#
# Usage: ./scripts/local-install.sh
set -euo pipefail

NAMESPACE="vpsie"
TYPE="vpsie"
# Version is only used for the on-disk path; dev_overrides ignore it.
VERSION="0.0.0-dev"

OS="$(go env GOOS)"
ARCH="$(go env GOARCH)"
BIN="terraform-provider-${TYPE}"

# ~/.terraform.d/plugins is the standard local filesystem mirror location.
DEST="${HOME}/.terraform.d/plugins/registry.terraform.io/${NAMESPACE}/${TYPE}/${VERSION}/${OS}_${ARCH}"

echo "==> Building ${BIN}"
go build -o "${BIN}" .

echo "==> Installing to ${DEST}"
mkdir -p "${DEST}"
mv "${BIN}" "${DEST}/"

GOBIN_DIR="$(go env GOBIN)"
[ -n "${GOBIN_DIR}" ] || GOBIN_DIR="$(go env GOPATH)/bin"

cat <<HINT

Done. For active development, prefer a dev override in ~/.terraformrc:

provider_installation {
  dev_overrides {
    "vpsie/vpsie" = "${GOBIN_DIR}"
  }
  direct {}
}

Then run 'go install .' to update the binary GOBIN uses. With a dev override,
'terraform init' is not required (and will warn that overrides are in effect).
HINT
