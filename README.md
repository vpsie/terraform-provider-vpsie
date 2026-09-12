# Terraform Provider for VPSie

The VPSie Terraform provider lets you manage resources on the
[VPSie](https://vpsie.com) cloud platform — servers, storage, snapshots,
networking (VPC, floating IPs, firewalls, gateways), DNS, load balancers,
Kubernetes, container registries, managed databases, certificates, tags and
more — using [Terraform](https://www.terraform.io).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://go.dev/doc/install) >= 1.25 (to build the provider from source)
- A VPSie API access token

## Using the provider

```hcl
terraform {
  required_providers {
    vpsie = {
      source = "vpsie/vpsie"
    }
  }
}

provider "vpsie" {
  # Recommended: export VPSIE_ACCESS_TOKEN instead of hard-coding a token.
  access_token = var.vpsie_access_token
}

resource "vpsie_tag" "example" {
  name  = "production"
  color = "#ff0000"
}
```

Set your token via the environment so it never lands in configuration or state
files:

```sh
export VPSIE_ACCESS_TOKEN="your-api-token"
```

Full documentation for every resource and data source lives in [`docs/`](./docs)
and on the [Terraform Registry](https://registry.terraform.io/providers/vpsie/vpsie).

## Developing the provider

```sh
# Build
go build -v .

# Lint
golangci-lint run

# Regenerate documentation from schemas and examples
go generate ./...

# Acceptance tests (creates real resources; requires a token)
TF_ACC=1 VPSIE_ACCESS_TOKEN="your-api-token" go test ./... -v -timeout 120m
```

## License

See [LICENSE](./LICENSE) if present, or the repository for licensing details.
