VERSION ?= 0.0.1
IMG ?= nexus.dev.aveshalabs.io/kubeslice-dns:$(VERSION)

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: build
build: fmt vet ## Build coredns binary
	go build -o bin/coredns main.go

.PHONY: run
run:
	go run main.go

.PHONY: docker-build
docker-build:
	docker build -t ${IMG} .
