GOLANGCI_LINT ?= go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: testacc
testacc: clearacc
	TF_ACC=1 go test ./... -p 1 -v -cover $(TESTARGS) --timeout 10m --count 1

.PHONY: clearacc
clearacc:
	TF_ACC=1 go test ./internal/provider -v --sweep all

.PHONY: generate
generate:
	go generate

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run

.PHONY: fix
fix:
	$(GOLANGCI_LINT) run --fix
