package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	mu          sync.Mutex
	file        *os.File
	writer      *bufio.Writer
	maxSize     int64
	currentSize int64
	timeStr     atomic.Value
}

func NewLogger(filename string, maxSizeGB int) (*Logger, error) {
	path := filepath.Join("..", filename)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	bw := bufio.NewWriter(f)
	fileInfo, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	limitBytes := int64(maxSizeGB) * 1024 * 1024 * 1024

	l := &Logger{
		file:        f,
		writer:      bw,
		maxSize:     limitBytes,
		currentSize: fileInfo.Size(),
	}

	l.timeStr.Store(time.Now().Format("2006-01-02 15:04:05"))

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			t := time.Now().Format("2006-01-02 15:04:05")
			l.timeStr.Store(t)
		}
	}()

	return l, nil
}

func (l *Logger) Log(level string, msg string) error {
	var sb strings.Builder
	sb.Grow(100)

	timestamp := l.timeStr.Load().(string)
	UserID := rand.IntN(100000)

	sb.WriteString(timestamp)
	sb.WriteByte(' ')

	sb.WriteByte('[')
	sb.WriteString(level)
	sb.WriteString("] ")

	sb.WriteString("UserID:")
	sb.WriteString(strconv.Itoa(UserID))
	sb.WriteByte(' ')

	sb.WriteString("Message:")
	sb.WriteString(msg)
	sb.WriteByte('\n')

	logLine := sb.String()

	lineSize := int64(len(logLine))

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentSize+lineSize > l.maxSize {
		return errors.New("log file limit reached, writing stopped")
	}

	n, err := l.writer.WriteString(logLine)
	if err != nil {
		return err
	}

	l.currentSize += int64(n)
	return nil
}

func (l *Logger) Close() error {
	if err := l.writer.Flush(); err != nil {
		return err
	}
	return l.file.Close()
}

func main() {
	logger, err := NewLogger("stress_test.log", 1)
	if err != nil {
		panic(err)
	}
	defer logger.Close()
	var wg sync.WaitGroup
	concurrency := runtime.NumCPU()

	typeofLog := [4]string{"INFO", "WARN", "ERROR", "DEBUG"}

	fmt.Println("Starting Stress Test (Target: 1GB)...")
	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			count := 0

			for {
				n := rand.IntN(4)
				err := logger.Log(typeofLog[n], "This is a log message content")
				if err != nil {
					fmt.Printf("Goroutine %d stopped: %v\n", id, err)
					return
				}
				count++
				if count%100000 == 0 {
					fmt.Printf("Goroutine %d wrote %d lines...\n", id, count)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	logger.Close()

	fi, err := os.Stat("stress_test.log")
	if err != nil {
		panic(err)
	}
	fmt.Println("------------------------------------------------")
	fmt.Printf("Test Finished in: %v\n", duration)
	fmt.Printf("Final File Size:  %.2f GB\n", float64(fi.Size())/(1024*1024*1024))
	fmt.Println("------------------------------------------------")
}
