TEST?=$$(go list ./...)
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
GOOS?=$$(go env GOOS)
GOARCH?=$$(go env GOARCH)
VERSION?=$$(git describe --abbrev=0 --tags)

default: build

install: fmtcheck
	@echo "Installing the provider..."
	@go install terraform-provider-wallarm_$(VERSION) .

build: fmtcheck
	@echo "Building the sideloaded provider..."
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
	go test ./... -v -timeout=30s -parallel=4 -race -cover

testacc: fmtcheck
	TF_ACC=1 go test ./... -v -timeout 120m -race -cover

vet:
	@echo "go vet ."
	@go vet $$(go list ./...) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

apply: build init
	@echo "Running .tf files in the folder..."
	@terraform apply

destroy: build init
	@echo "Running .tf files in the folder..."
	@terraform destroy

.PHONY: install build init init-plugin test testacc vet fmt fmtcheck apply destroy
