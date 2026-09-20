
output ?= go-pve-autosnap
version ?= devel

PHONY: all
all:

.PHONY: build
build:
	@go build \
		-trimpath \
		-ldflags="-s -w -X main.version=$(version)" \
		-o $(output)

.PHONY: coverage
coverage:
	@go test -covermode=atomic -coverprofile=_coverage.out $(UNIT_TEST_PATHS) \
		; go tool cover -html=_coverage.out -o _coverage.html

.PHONY: install
install: build
	mv go-pve-autosnap /usr/bin/go-pve-autosnap
	ln -s /usr/bin/go-pve-autosnap /usr/bin/pve-autosnap

.PHONY: lint
lint:
	@golint ./...

UNIT_TEST_PATHS=./cmd/... ./internal/...

.PHONY: test
test:
	@go test -race -vet=off $(UNIT_TEST_PATHS)

.PHONY: uninstall
uninstall:
	rm /usr/bin/go-pve-autosnap
	rm /usr/bin/pve-autosnap

.PHONY: vet
vet:
	@go vet ./...
