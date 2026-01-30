# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Build the provider binary
go build -v .

# Run acceptance tests (requires VPSIE_ACCESS_TOKEN env var)
TF_ACC=1 go test ./... -v -timeout 120m

# Run a single test
TF_ACC=1 go test ./internal/services/storage -run TestAccStorageResource -v -timeout 120m

# Regenerate docs and format examples
go generate ./...

# Lint
golangci-lint run
```

## Architecture

This is a **Terraform provider for the VPSie cloud platform**, built on the HashiCorp Terraform Plugin Framework v1 (`terraform-plugin-framework`).

### Entry Point

`main.go` starts a plugin server with debug support. The provider address is `registry.terraform.local/hashicorp/vpsie`.

### Provider Configuration

`internal/provider/provider.go` defines the provider. It accepts a single `access_token` attribute (or reads `VPSIE_ACCESS_TOKEN` from env) and creates an authenticated `govpsie.Client` used by all resources and data sources.

### Service Modules

Each cloud service lives in `internal/services/<service>/` with two files:
- `<service>_resource.go` — implements CRUD via the `resource.Resource` interface
- `<service>_datasource.go` — implements read-only data fetching via `datasource.DataSource`

Services: backup, domain, firewall, gateway, image, kubernetes, loadbalancer, project, script, server, snapshot, sshkey, storage, vpc.

The kubernetes service also has a `kubernetes_group_resource.go` for node group management.

### Resource Implementation Pattern

Every resource follows this structure:
1. A struct holding `*govpsie.Client`
2. A model struct with `types.*` fields mapping Terraform schema to Go
3. Interface methods: `Metadata()`, `Schema()`, `Configure()`, `Create()`, `Read()`, `Update()`, `Delete()`, `ImportState()`
4. `Configure()` extracts the `govpsie.Client` from provider data
5. CRUD methods call the `govpsie` SDK, then map API responses to the model

Data sources are similar but only implement `Metadata()`, `Schema()`, `Configure()`, and `Read()`.

### API Client

All API calls go through `github.com/vpsie/govpsie` — an external SDK. The provider does not make HTTP calls directly.

### Testing

Tests use `terraform-plugin-testing` with `resource.TestCase` and `ProtoV5ProviderFactories` from `internal/acctest/acctest.go`. Test files follow the pattern `*_resource_test.go` alongside their resource.

### Code Generation & Docs

`go generate ./...` runs two commands (declared in `main.go`):
- `terraform fmt -recursive ./examples/` — formats example HCL files
- `tfplugindocs` — generates `docs/` from schemas and `examples/`

### CI/CD

- `.github/workflows/test.yml` — builds, lints, validates code generation, runs acceptance tests across Terraform 1.0–1.4
- `.github/workflows/release.yml` — goreleaser builds multi-platform binaries on version tags (`v*`)

## Working Guidelines

1. First think through the problem, read the codebase for relevant files.
2. Before you make any major changes, check in with me and I will verify the plan.
3. Please every step of the way just give me a high level explanation of what changes you made.
4. Make every task and code change you do as simple as possible. We want to avoid making any massive or complex changes. Every change should impact as little code as possible. Everything is about simplicity.
5. Maintain a documentation file that describes how the architecture of the app works inside and out.
6. Never speculate about code you have not opened. If the user references a specific file, you MUST read the file before answering. Make sure to investigate and read relevant files BEFORE answering questions about the codebase.
