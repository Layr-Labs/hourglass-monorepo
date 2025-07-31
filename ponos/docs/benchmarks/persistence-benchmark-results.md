# Persistence Layer Benchmark Results

## Overview

This document summarizes the performance characteristics of the in-memory storage implementation compared to a baseline sync.Map implementation.

## Benchmark Results

### Task Operations

| Operation | In-Memory Store | sync.Map | Performance Delta |
|-----------|-----------------|----------|-------------------|
| SaveTask | 620.9 ns/op (493 B/op, 10 allocs) | 679.7 ns/op (415 B/op, 13 allocs) | **+9.5% faster** |
| GetTask | 68.75 ns/op (13 B/op, 1 alloc) | 71.48 ns/op (13 B/op, 1 alloc) | **+3.8% faster** |
| ListPendingTasks | 12,094 ns/op (17,528 B/op, 11 allocs) | N/A | - |

### Concurrent Operations

| Operation | In-Memory Store | sync.Map | Performance Delta |
|-----------|-----------------|----------|-------------------|
| Concurrent Read/Write | 96,444 ns/op (198,814 B/op, 14 allocs) | 114,985 ns/op (105,072 B/op, 16 allocs) | **+16.1% faster** |

### Block Operations

| Operation | In-Memory Store | sync.Map | Performance Delta |
|-----------|-----------------|----------|-------------------|
| SetLastProcessedBlock | 9.077 ns/op (0 B/op, 0 allocs) | 117.0 ns/op (48 B/op, 4 allocs) | **+92.2% faster** |

### Config Operations

| Operation | In-Memory Store | sync.Map | Performance Delta |
|-----------|-----------------|----------|-------------------|
| SaveOperatorSetConfig | 235.8 ns/op (136 B/op, 7 allocs) | 277.9 ns/op (176 B/op, 9 allocs) | **+15.1% faster** |

### Executor Operations

| Operation | In-Memory Store |
|-----------|-----------------|
| SavePerformer | 479.7 ns/op (345 B/op, 5 allocs) |
| ListPerformers | 5,092 ns/op (20,096 B/op, 101 allocs) |

### Memory Usage

The large dataset test (10,000 tasks) shows:
- Operation time: 161,235 ns/op
- Memory allocation: 311,074 B/op
- Allocations: 37 allocs/op

## Key Findings

1. **Performance**: The in-memory store consistently outperforms sync.Map across all operations:
   - Simple operations (get/set) are 3-15% faster
   - Block operations are dramatically faster (92% improvement)
   - Concurrent operations show 16% improvement

2. **Memory Efficiency**: 
   - Similar memory footprint for most operations
   - Block operations use zero allocations vs 4 for sync.Map
   - Efficient handling of large datasets

3. **Additional Benefits**:
   - Type safety with structured interfaces
   - Built-in status validation and business logic
   - Defensive copying prevents data corruption
   - Support for complex queries (e.g., ListPendingTasks)

## Conclusion

The custom in-memory storage implementation provides better performance than sync.Map while adding essential features like:
- Status transition validation
- Defensive copying
- Complex query support
- Type-safe interfaces

These benchmarks establish a strong baseline for comparing future storage implementations (e.g., BadgerDB).