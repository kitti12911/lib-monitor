# lib-monitor

observability helpers for homelab Go services. provides OpenTelemetry tracing setup and continuous profiling with Pyroscope.

## install

```bash
go get github.com/kitti12911/lib-monitor
```

## packages

### tracing

opentelemetry tracing setup with OTLP gRPC exporter.

```go
import "github.com/kitti12911/lib-monitor/tracing"

tp, err := tracing.New(ctx, "my-service", "localhost:4317")
if err != nil {
    log.Fatal(err)
}
defer tracing.Shutdown(ctx, tp)
```

- exports traces via OTLP gRPC (e.g. to alloy, otel collector)
- sets global tracer provider
- supports TraceContext and Baggage propagation

### profiling

continuous profiling setup with Pyroscope.

```go
import "github.com/kitti12911/lib-monitor/profiling"

profiler, err := profiling.New(
    "my-service",
    "http://pyroscope.observability.svc.cluster.local:4040",
    profiling.WithNamespace("demo"),
)
if err != nil {
    log.Fatal(err)
}
defer profiling.Shutdown(profiler)
```

- exports profiles to pyroscope
- enables cpu, allocation, in-use heap, and goroutine profiles by default
- supports custom tags, profile types, logger, basic auth, and tenant id options

options:

| function                     | description                                |
| ---------------------------- | ------------------------------------------ |
| `WithNamespace(namespace)`   | add namespace tag to profiling data        |
| `WithTags(tags)`             | add or override custom profiling tags      |
| `WithProfileTypes(types...)` | choose which pyroscope profiles to collect |
| `WithLogger(logger)`         | set pyroscope client logger                |
| `WithBasicAuth(user, pass)`  | set basic auth for pyroscope               |
| `WithTenantID(tenantID)`     | set pyroscope tenant id                    |

## requirements

- go 1.26 or higher

## available commands

```bash
go mod tidy
go fmt ./...
go test ./...
```
