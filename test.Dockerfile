FROM golang:1.23rc2 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
ADD vendor vendor
COPY Makefile Makefile

# Copy the go source
COPY main.go main.go
COPY plugin/ plugin/
COPY vendor/ vendor/

CMD ["make", "test"]
