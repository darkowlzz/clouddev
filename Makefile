PROJECT = github.com/darkowlzz/clouddev
GOARCH ?= amd64
GO_VERSION = 1.15.0
CACHE_DIR = ${CURDIR}/.cache

BIN_DIR = ${CURDIR}/bin
GOLANGCI_LINT = ${BIN_DIR}/golangci-lint
GOLANGCI_LINT_VERSION = "v1.30.0"


.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""
	@echo "To run any of the above in docker, suffix the command with '-docker':"
	@echo ""
	@echo "  make clouddev-docker"
	@echo ""


##############################
# Development                #
##############################

##@ Development

.PHONY: tidy clouddev install cobra lint test

clouddev: ## Build clouddev binary.
	go build -mod=vendor -o bin/clouddev $(shell ./build/print-ldflags.sh) .

install: clouddev ## Build and install clouddev.
	cp ./bin/clouddev $(shell go env GOPATH)/bin/clouddev

tidy: ## Prune, add and vendor go dependencies.
	go mod tidy -v
	go mod vendor -v

cobra: ## Run cobra. Pass args with ARGS="arg1 arg2..."
	go run -mod=vendor ./vendor/github.com/spf13/cobra/cobra $(ARGS)

lint: golangci-lint ## Run code lint.
	$(GOLANGCI_LINT) run -v

test: ## Run all the tests.
	go test -mod=vendor -v -race ./... -count=1

##############################
# Third-party tools          #
##############################

.PHONY: golangci-lint

golangci-lint:
	@if [ ! -f $(GOLANGCI_LINT) ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN_DIR) $(GOLANGCI_LINT_VERSION); \
	fi


# This target matches any target ending in '-docker' eg. 'tidy-docker'. This
# allows running makefile targets inside a container by appending '-docker' to
# it.
%-docker:
	# Create cache dirs.
	mkdir -p $(CACHE_DIR)/go $(CACHE_DIR)/cache
	# golangci-lint build cache.
	mkdir -p $(CACHE_DIR)/golangci-lint
	# Run the make target in docker.
	docker run -it --rm \
		-v $(CACHE_DIR)/go:/go \
		-v $(CACHE_DIR)/cache:/.cache/go-build \
		-v $(CACHE_DIR)/golangci-lint:/.cache/golangci-lint \
		-v $(shell pwd):/go/src/${PROJECT} \
		-w /go/src/${PROJECT} \
		-u $(shell id -u):$(shell id -g) \
		-e GOARCH=$(GOARCH) \
		--entrypoint "make" \
		golang:$(GO_VERSION) \
		"$(patsubst %-docker,%,$@)"
