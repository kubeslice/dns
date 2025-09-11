# KubeSlice CoreDNS Plugin Debug Setup

This guide helps you run CoreDNS with the modified KubeSlice plugin locally to debug DNS lookup latency issues.

## Prerequisites

1. **Go 1.18+** installed
2. **Docker** for containerized testing
3. **kubectl** configured to access your Kubernetes cluster
4. **kubeslice-worker-operator** CRDs installed in your cluster

## Building the Modified CoreDNS

### 1. Build the CoreDNS binary with debug logging

```bash
# Build the CoreDNS binary with the modified plugin
go build -o coredns-debug main.go

# Or build with additional debug flags
go build -ldflags="-s -w" -o coredns-debug main.go
```

### 2. Create a debug Corefile

Create `Corefile.debug`:

```bash
cat > Corefile.debug << 'EOF'
.:1053 {
    debug
    log
    kubeslice
    forward . 8.8.8.8 8.8.4.4
}

slice.local:1053 {
    debug
    log
    kubeslice
}
EOF
```

## Running CoreDNS Locally

### Option 1: Direct Binary Execution

```bash
# Set KUBECONFIG to your cluster
export KUBECONFIG=/path/to/your/kubeconfig

# Run CoreDNS with debug logging
./coredns-debug -conf Corefile.debug -dns.port 1053
```

### Option 2: Docker Container

```bash
# Build Docker image
cat > Dockerfile.debug << 'EOF'
FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o coredns-debug main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/coredns-debug .
COPY Corefile.debug .
CMD ["./coredns-debug", "-conf", "Corefile.debug", "-dns.port", "1053"]
EOF

# Build and run
docker build -f Dockerfile.debug -t coredns-kubeslice-debug .
docker run --rm -p 1053:1053/udp -v $KUBECONFIG:/root/.kube/config:ro coredns-kubeslice-debug
```

## Testing DNS Queries

### 1. Test with dig

```bash
# Test slice.local queries
dig @localhost -p 1053 test-service.slice.local

# Test with timing
dig @localhost -p 1053 test-service.slice.local +time=5 +tries=1

# Compare with cluster.local (if available)
dig @localhost -p 1053 test-service.cluster.local
```

### 2. Test with nslookup

```bash
# Test slice.local
nslookup test-service.slice.local localhost

# Test with verbose output
nslookup -debug test-service.slice.local localhost
```

### 3. Test with curl (if testing HTTP services)

```bash
# If your service exposes HTTP
curl -v http://test-service.slice.local:8080/health
```

## Monitoring and Debugging

### 1. Watch CoreDNS Logs

The modified plugin will output detailed timing information:

```bash
# Watch logs in real-time
./coredns-debug -conf Corefile.debug -dns.port 1053 2>&1 | grep -E "(DNS Query|Cache|Services|ServiceImport)"
```

### 2. Key Log Messages to Watch

Look for these log patterns:

- `DNS Query received` - Shows incoming query timing
- `A record lookup completed` - Shows lookup duration
- `Cache access completed` - Shows cache performance
- `Services lookup completed` - Shows service matching time
- `ServiceImport reconcile` - Shows controller sync timing

### 3. Performance Analysis

Monitor these metrics:

- `total_duration_ms` - Total query processing time
- `lookup_duration_ms` - A record lookup time
- `cache_duration_ms` - Cache access time
- `match_duration_ms` - Endpoint matching time
- `cache_hits` vs `cache_misses` - Cache efficiency

## Simulating ServiceImport Resources

### 1. Create a test ServiceImport

```bash
cat > test-serviceimport.yaml << 'EOF'
apiVersion: networking.kubeslice.io/v1beta1
kind: ServiceImport
metadata:
  name: test-service
  namespace: default
spec:
  dnsName: test-service.slice.local
  slice: test-slice
  aliases:
    - test-alias.slice.local
status:
  endpoints:
    - dnsName: test-service.slice.local
      ip: 10.0.0.1
    - dnsName: test-service.slice.local
      ip: 10.0.0.2
EOF

# Apply to your cluster
kubectl apply -f test-serviceimport.yaml
```

### 2. Watch ServiceImport events

```bash
# Watch for ServiceImport changes
kubectl get serviceimports -w

# Watch ServiceImport events
kubectl get events --field-selector involvedObject.kind=ServiceImport -w
```

## Troubleshooting Common Issues

### 1. Permission Issues

```bash
# Ensure proper RBAC
kubectl auth can-i get serviceimports
kubectl auth can-i list serviceimports
kubectl auth can-i watch serviceimports
```

### 2. Network Issues

```bash
# Test DNS connectivity
nslookup google.com localhost

# Test UDP port
nc -u localhost 1053
```

### 3. Cache Issues

```bash
# Check cache statistics in logs
grep "Cache access completed" logs.txt | tail -10

# Look for cache hit/miss ratios
grep "cache_hits\|cache_misses" logs.txt
```

## Performance Testing

### 1. Load Testing

```bash
# Install dnsperf if available
# dnsperf -s localhost -p 1053 -d queries.txt -c 10 -l 30

# Or use dig in a loop
for i in {1..100}; do
  time dig @localhost -p 1053 test-service.slice.local +short
done
```

### 2. Latency Measurement

```bash
# Measure average latency
time for i in {1..10}; do dig @localhost -p 1053 test-service.slice.local +short > /dev/null; done
```

## Expected Debug Output

With the modified plugin, you should see logs like:

```
[INFO] DNS Query received name=test-service.slice.local. type=1 timestamp=2024-01-01T10:00:00Z
[INFO] Services method called name=test-service.slice.local. exact=false timestamp=2024-01-01T10:00:00Z
[INFO] Cache access completed name=test-service.slice.local. cache_duration_ms=0 total_endpoints=2 cache_hits=1 cache_misses=0
[INFO] Services lookup completed name=test-service.slice.local. total_duration_ms=1 cache_duration_ms=0 match_duration_ms=0 matches_found=2
[INFO] A record lookup completed name=test-service.slice.local. duration_ms=2 records_found=2
[INFO] DNS response sent name=test-service.slice.local. total_duration_ms=3 lookup_duration_ms=2 response_duration_ms=0 records_count=2
```

This will help you identify exactly where the 2-3 second delay is occurring.
