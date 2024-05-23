GO_VERSION=1.22
GOLINT_VERSION=v1.57.2
TEMPL_VERSION=v0.2.680
CLUSTERINFO_VERSION=0.1.0
PROJECT ?= ark-clusterinfo

ROOT_DIR=$(shell git rev-parse --show-toplevel)
TOOLS_DIR=$(ROOT_DIR)/.tools
clusterinfo := $(ROOT_DIR)/bin/ark-clusterinfo

ALL_GO_FILES=$(shell find $(ROOT_DIR) -type f -name "*.go")

LINT := $(TOOLS_DIR)/golangci-lint

GOCMD ?= go
TEMPLCMD ?= templ
GO_ENV=$(shell CGO_ENABLED=0)


$(TOOLS_DIR):
	mkdir -p $@

.PHONY: generate
generate: 
	$(TEMPLCMD) generate

.PHONY: check-fmt
check-fmt: fmt
	@git diff -s --exit-code *.go || (echo "Build failed: a go file is not formated correctly. Run 'make fmt' and update your PR." && exit 1)

.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...
	$(TEMPLCMD) fmt .

.PHONY: govet
govet:
	$(GOCMD) vet ./...

.PHONY: gotest
test: gofmt govet ensure-fmt
	$(GO_ENV) $(GOCMD) test -v ./... -failfast

.PHONY: gomoddownload
gomoddownload:
	$(GOCMD) mod download -x

.PHONY: tools
tools: $(TOOLS_DIR)
	GOBIN=$(TOOLS_DIR) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLINT_VERSION)
	GOBIN=$(TOOLS_DIR) go install github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION)

.PHONY: golint
golint:
	$(LINT) run --verbose --allow-parallel-runners --timeout=10m 

.PHONY: gotidy
gotidy:
	$(GOCMD) mod tidy -compat=$(GO_VERSION)

.PHONY: build
build: generate
	$(GOCMD) build -o bin/clusterinfo cmd/server/main.go

.PHONY: exec
exec: gofmt build 
	./bin/clusterinfo 

.PHONY: run
run: generate
	$(GOCMD) run cmd/server/main.go

.PHONY: archive
archive: build
	mkdir -p src/
	@echo "create a tarball..."
	tar -cz \
	--file ./src/$(PROJECT)-$(CLUSTERINFO_VERSION).tar.gz \
	-C ./bin .
	@echo "output:"
	@find src/*.tar.gz
