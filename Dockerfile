FROM golang:1.18 as builder

WORKDIR /
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY plugin/ plugin/
COPY vendor/ vendor/

# Build
RUN go env -w GOPRIVATE=github.org/kubeslice && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -mod=vendor -a -o coredns main.go

FROM gcr.io/distroless/static:nonroot
USER nonroot:nonroot

EXPOSE 1053 1053/udp
ENTRYPOINT ["/coredns"]
