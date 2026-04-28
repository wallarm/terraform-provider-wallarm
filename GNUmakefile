TEST?=$$(go list ./...)
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
GOOS?=$$(go env GOOS)
GOARCH?=$$(go env GOARCH)
VERSION?=$$(git describe --abbrev=0 --tags)
DEV_VERSION=2.99.0
TESTTIMEOUT=120m

WALLARM_API_HOST?=https://audit.api.wallarm.com

default: build

install: fmtcheck
	@echo "Installing the provider..."
	@go install terraform-provider-wallarm_$(VERSION) .

build: fmtcheck
	@echo "Building a sideloaded provider..."
	@go build -o terraform-provider-wallarm_$(VERSION) .
	@echo "Binary has been built to $(CURDIR)"

init:
	@echo "Initializing the Wallarm provider..."
	@terraform init

dev: fmtcheck
	@echo "Building dev provider (v$(DEV_VERSION))..."
	@go build -o terraform-provider-wallarm .
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/wallarm/wallarm/$(DEV_VERSION)/$(GOOS)_$(GOARCH)
	@cp terraform-provider-wallarm ~/.terraform.d/plugins/registry.terraform.io/wallarm/wallarm/$(DEV_VERSION)/$(GOOS)_$(GOARCH)/terraform-provider-wallarm_v$(DEV_VERSION)
	@rm terraform-provider-wallarm
	@echo "Installed to ~/.terraform.d/plugins/ (v$(DEV_VERSION)). Run 'terraform init' in your HCL project."

dev-clean:
	@rm -rf ~/.terraform.d/plugins/registry.terraform.io/wallarm/wallarm/$(DEV_VERSION)
	@echo "Dev provider (v$(DEV_VERSION)) removed. Run 'terraform init -upgrade' to switch to released version."

init-plugin: build
	@echo "Initializing and copying the Wallarm provider..."
	./scripts/plugindircheck.sh $(GOOS) $(GOARCH) $(VERSION)
	@terraform init

lint:
	@echo "Running golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run

test:
	go test $(TEST) -v -timeout=30s -parallel=4 -race -cover

testacc: fmtcheck
	TF_ACC=1 TF_ACC_TERRAFORM_PATH=$${TF_ACC_TERRAFORM_PATH:-$$(command -v terraform)} go test $(TEST) -v $(TESTARGS) -timeout $(TESTTIMEOUT) -cover -ldflags="-X=github.com/wallarm/terraform-provider-wallarm/version.ProviderVersion=acc"

vet:
	@echo "go vet ."
	@go vet $(TEST) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

.PHONY: install build init dev dev-clean init-plugin test testacc vet fmt fmtcheck
