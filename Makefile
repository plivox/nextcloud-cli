BINARY := nextcloud

GOLINES := go run github.com/golangci/golines@v0.16.0

GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0

.PHONY: build run clean fmt lint vet tidy

build:
	go build -o $(BINARY) .

run:
	go run . $(ARGS)

clean:
	rm -f $(BINARY)

fmt:
	-$(GOLANGCI_LINT) run --fix
	$(GOLINES) -m 80 --no-chain-split-dots -w .

lint:
	$(GOLANGCI_LINT) run

vet:
	go vet ./...

tidy:
	go mod tidy
