#!/bin/bash

set -eo pipefail

version="0.0.0-local"
platform="$(go env GOOS)_$(go env GOARCH)"

mkdir -p ~/.terraform.d/plugins/registry.terraform.io/epilayer-public/epilayer/$version/$platform
cp terraform-provider-epilayer ~/.terraform.d/plugins/registry.terraform.io/epilayer-public/epilayer/$version/$platform/terraform-provider-epilayer

# Cleanup:
# rm -fr ~/.terraform.d/plugins/registry.terraform.io/epilayer-public/
