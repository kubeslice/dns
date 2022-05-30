VERSION ?= latest
IMG ?= docker.io/aveshasystems/dns:$(VERSION)

.PHONY: build
build: fmt vet ## Build coredns binary
	go build -o bin/coredns main.go

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: run
run:
	go run main.go

.PHONY: docker-build
docker-build:
	docker build -t ${IMG} .

.PHONY: docker-run
docker-run:
	docker run -ti ${IMG}

.PHONY: docker-push
docker-push: ## Push docker image
	docker push ${IMG}

.PHONY: test
test: # Run UTs
	go test ./...

.PHONY: test-docker
test-docker:
	docker build -t test -f test.Dockerfile . && docker run test
