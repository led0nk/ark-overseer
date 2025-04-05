GO_VERSION=1.23
GOLINT_VERSION=v2.0.2
TEMPL_VERSION=v0.2.680
OVERSEER_VERSION ?= 0.1.0
PROJECT ?= ark-overseer



ROOT_DIR=$(shell git rev-parse --show-toplevel)
TOOLS_DIR=$(ROOT_DIR)/.tools
PROJECT_DIR=$(PROJECT)-$(OVERSEER_VERSION)
overseer := $(ROOT_DIR)/bin/ark-overseer

ALL_GO_FILES=$(shell find $(ROOT_DIR) -type f -name "*.go")

LINT := $(TOOLS_DIR)/golangci-lint
TEMPL := $(TOOLS_DIR)/templ

GOCMD ?= go
GO_ENV=$(shell CGO_ENABLED=0)


$(TOOLS_DIR):
	mkdir -p $@

.PHONY: generate
generate: 
	$(TEMPL) generate

.PHONY: check-fmt
check-fmt: fmt
	@git diff -s --exit-code *.go || (echo "Build failed: a go file is not formated correctly. Run 'make fmt' and update your PR." && exit 1)

.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...
	$(TEMPL) fmt .

.PHONY: govet
govet:
	$(GOCMD) vet ./...

.PHONY: test
test: govet 
	$(GO_ENV) $(GOCMD) test -v ./... -failfast

.PHONY: gomoddownload
gomoddownload:
	$(GOCMD) mod download -x

.PHONY: tools
tools: $(TOOLS_DIR)
	GOBIN=$(TOOLS_DIR) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLINT_VERSION)
	GOBIN=$(TOOLS_DIR) go install github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION)

.PHONY: lint
lint:
	$(LINT) run --verbose --allow-parallel-runners --timeout=10m 

.PHONY: tidy
tidy:
	$(GOCMD) mod tidy -compat=$(GO_VERSION)

.PHONY: vendor
vendor: 
	$(GOCMD) mod vendor

.PHONY: build
build: tools generate
	$(GOCMD) build -o bin/overseer cmd/api/main.go

.PHONY: exec
exec: gofmt build 
	./bin/overseer 

.PHONY: run
run: generate
	$(GOCMD) run cmd/api/main.go

.PHONY: archive
archive: vendor
	mkdir -p src/
	mkdir -p _build/$(PROJECT_DIR)
	cp go.mod go.sum $(PROJECT).service _build/$(PROJECT_DIR)
	GO_DIRS=$$(find . -type f -name '*.go' -exec dirname {} \; | awk -F'/' '{print $$2}' | sort -u) && \
	for DIR in $$GO_DIRS; do \
		echo "copying files from $$DIR to _build/$$DIR"; \
		cp -r $$DIR _build/$(PROJECT_DIR); \
	done  

	@echo "create a tarball..." 
	tar -cz \
	--file ./src/$(PROJECT_DIR).tar.gz \
	-C ./_build $(PROJECT_DIR)
	@echo "output:" 
	@find src/*.tar.gz 

.PHONY: clean
clean:
	rm -rf *.tar.gz *.rpm
	rm -rf ./SRPMS
	rm -rf ark-overseer-0.1.0*
	rm -rf ./src
	rm -rf ./_build

