TEST?=$$(go list ./...)
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
GOOS?=$$(go env GOOS)
GOARCH?=$$(go env GOARCH)
VERSION?=$$(git describe --abbrev=0 --tags)
TESTTIMEOUT=120m

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

init-plugin: build
	@echo "Initializing and copying the Wallarm provider..."
	./scripts/plugindircheck.sh $(GOOS) $(GOARCH) $(VERSION)
	@terraform init

test:
	go test $(TEST) -v -timeout=30s -parallel=4 -race -cover

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout $(TESTTIMEOUT) -race -cover -ldflags="-X=github.com/wallarm/terraform-provider-wallarm/version.ProviderVersion=acc"

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

.PHONY: install build init init-plugin test testacc vet fmt fmtcheck
