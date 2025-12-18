# syntax=docker/dockerfile:1.4
FROM --platform=$BUILDPLATFORM golang:1.24.2 AS builder

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETOS=linux

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Copy the go source
COPY main.go main.go
COPY plugin/ plugin/
COPY vendor/ vendor/

# Build with cross-compilation
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GO111MODULE=on \
    go build -mod=vendor -ldflags="-w -s" -trimpath -o coredns main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /workspace/coredns .
COPY Corefile Corefile

USER nonroot:nonroot

EXPOSE 1053 1053/udp
ENTRYPOINT ["/coredns"]
