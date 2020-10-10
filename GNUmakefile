TEST?=$$(go list ./...)
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
GOOS?=$$(go env GOOS)
GOARCH?=$$(go env GOARCH)
VERSION?=$$(git describe --abbrev=0 --tags)
TESTTIMEOUT=120m
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=wallarm

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
	go test $(TEST) -v -timeout=30s -parallel=4 -race -cover

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout $(TESTTIMEOUT) -race -cover -ldflags="-X=github.com/416e64726579/terraform-provider-wallarm/version.ProviderVersion=acc"

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

apply: build init
	@echo "Running .tf files in the folder..."
	@terraform apply

destroy: build init
	@echo "Running .tf files in the folder..."
	@terraform destroy

website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	ln -sf ../../../../ext/providers/wallarm/website/docs $(GOPATH)/src/github.com/hashicorp/terraform-website/content/source/docs/providers/wallarm
	ln -sf ../../../ext/providers/wallarm/website/wallarm.erb $(GOPATH)/src/github.com/hashicorp/terraform-website/content/source/layouts/wallarm.erb
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	ln -sf ../../../../ext/providers/wallarm/website/docs $(GOPATH)/src/github.com/hashicorp/terraform-website/content/source/docs/providers/wallarm
	ln -sf ../../../ext/providers/wallarm/website/wallarm.erb $(GOPATH)/src/github.com/hashicorp/terraform-website/content/source/layouts/wallarm.erb
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

.PHONY: install build init init-plugin test testacc vet fmt fmtcheck apply destroy website website-test
