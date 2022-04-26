FROM golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY plugin/ plugin/
COPY vendor/ vencor/

# Build
RUN go env -w GOPRIVATE=bitbucket.org/realtimeai && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -mod=vendor -a -o coredns main.go

FROM scratch

WORKDIR /

COPY --from=builder /workspace/coredns .
COPY Corefile Corefile

EXPOSE 53 53/udp
ENTRYPOINT ["/coredns"]
