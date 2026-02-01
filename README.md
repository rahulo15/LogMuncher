# LogMuncher 🪵🪓

A high-performance, concurrent log generator and analyzer built in Go. This project demonstrates advanced systems programming concepts, including parallel file chunking, memory optimization, and atomic operations.

**Performance Benchmark:** Processed **10GB** of log data in **8.29 seconds** (~1.2 GB/s) on a standard 12-core machine.

## 📂 Project Structure

```text
LogMuncher/
├── generator/
│   └── generator.go   # Multi-threaded log generator
├── reader/
│   └── reader.go      # High-performance parallel log analyzer
└── stress_test.log    # (Generated output file resides in the root)
```

## 🚀 How to Run

Open your terminal in the main `LogMuncher` folder.

### 1. Generate the Logs
Run the generator to create the `stress_test.log` file in your current folder.

```bash
go run generator/generator.go
```

### 2. Analyze the Logs
Run the reader to process the generated log file.

```bash
go run reader/reader.go
```

## 🧠 Technical Highlights

### The Optimization Journey
This project evolved through three stages of architectural refactoring to maximize hardware utilization:

1.  **Sequential Read (`bufio.Scanner`):** Achieved ~400 MB/s. Bottlenecked by a single CPU core.
2.  **Naive Concurrency (Channels):** Performance regression due to channel synchronization overhead and excessive memory allocation.
3.  **Parallel Chunking (Current Version):** Achieved ~1.2 GB/s. Successfully saturated the SSD bandwidth by eliminating synchronization bottlenecks.

### Key Features
* **Parallel File Chunking:** Implements a custom "Split-Line" algorithm to divide a single large file into `N` chunks (one per CPU core), allowing simultaneous reading without data corruption.
* **Zero-Copy Parsing:** Utilizes `bytes.Index` and `[]byte` manipulation instead of `string` conversions to drastically reduce Garbage Collection (GC) pressure.
* **Lock-Free Aggregation:** Each worker processes data in its own local map, merging results only once at completion to eliminate `sync.Mutex` contention.
* **Atomic Caching:** The generator leverages `sync/atomic` to cache timestamps, bypassing expensive `time.Format` calls during high-frequency write operations.

## Sample Output

```bash
Spawning 12 workers for 12 CPUs...
INFO 34242554
WARN 34243597
DEBUG 34252414
ERROR 34237782
Reading Finished in: 8.2895451s
```
