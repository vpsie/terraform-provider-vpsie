# terraform-provider-vpsie — common developer tasks.
NAME    := terraform-provider-vpsie
default: check

.PHONY: build install lint fmt docs test testacc check clean

build:        ## Compile the provider binary
	go build -v .

install:       ## Build and install the provider for local dev overrides
	./scripts/local-install.sh

lint:         ## Run golangci-lint (v2)
	golangci-lint run

fmt:          ## Format Go code and example HCL
	gofmt -w .
	terraform fmt -recursive ./examples/

docs:         ## Regenerate docs from schemas and examples
	go generate ./...

test:         ## Build/compile check and unit tests (no credentials)
	go test ./... -count=1

testacc:       ## Run acceptance tests (creates real resources; needs a token)
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

check: fmt lint build test  ## Format, lint, build and test

clean:        ## Remove the built binary
	rm -f $(NAME)
