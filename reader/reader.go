package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func processChunk(filePath string, start int64, end int64, resultChan chan map[string]int) {
	file, _ := os.Open(filePath)
	defer file.Close()

	file.Seek(start, 0)
	reader := bufio.NewReaderSize(file, 1024*1024)
	currentPos := start
	if start > 0 {
		discardedLine, _ := reader.ReadBytes('\n')
		currentPos += int64(len(discardedLine))
	}

	myCounts := make(map[string]int)

	for {
		if currentPos >= end {
			break
		}
		line, err := reader.ReadBytes('\n')
		currentPos += int64(len(line))

		if len(line) == 0 {
			break
		}

		start := bytes.Index(line, []byte("[")) + 1
		end := bytes.Index(line, []byte("]"))
		myCounts[string(line[start:end])] = myCounts[string(line[start:end])] + 1

		if err != nil {
			break
		}
	}

	resultChan <- myCounts
}

func main() {
	startTime := time.Now()
	filePath := filepath.Join("..", "stress_test.log")
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()
	numWorkers := runtime.NumCPU()
	chunkSize := fileSize / int64(numWorkers)

	fmt.Printf("Spawning %d workers for %d CPUs...\n", numWorkers, numWorkers)

	var wg sync.WaitGroup
	results := make(chan map[string]int, 100)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		start := int64(i) * chunkSize
		end := int64(i+1) * chunkSize
		if i == numWorkers-1 {
			end = fileSize
		}
		go func(s, e int64) {
			defer wg.Done()
			processChunk(filePath, s, e, results)
		}(start, end)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	globalmp := make(map[string]int)
	for mps := range results {
		for a, b := range mps {
			globalmp[a] = globalmp[a] + b
		}
	}

	for a, b := range globalmp {
		fmt.Println(a, b)
	}

	duration := time.Since(startTime)
	fmt.Printf("Reading Finished in: %v\n", duration)
}
