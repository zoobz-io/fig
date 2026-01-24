# Benchmarks

Performance benchmarks for fig.

## Running

```bash
make test-bench
```

Or directly:

```bash
go test -tags testing -bench=. -benchmem ./testing/benchmarks/...
```

## Writing Benchmarks

```go
func BenchmarkResolve(b *testing.B) {
    // Setup
    for b.Loop() {
        // Code to benchmark
    }
}
```

## Baseline Results

Captured on AMD Ryzen 5 3600X, Go 1.25:

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| Resolve_Small (3 fields) | 1244 | 176 | 6 |
| Resolve_Medium (10 fields) | 3321 | 248 | 12 |
| Resolve_Large (20 fields) | 7378 | 344 | 24 |
| Resolve_WithProvider | 1237 | 176 | 5 |
| ProviderGet_Hit | 10.85 | 0 | 0 |
| ProviderGet_Miss | 3.78 | 0 | 0 |
