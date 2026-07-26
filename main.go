package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

type match struct {
	path string
	line int
	text string
}

func main() {
	args := rebuildArgs(os.Args[1:])
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	dir := flag.String("dir", ".", "root directory to search")
	word := flag.String("word", "", "word to search for (case-insensitive)")
	workers := flag.Int("workers", 4, "number of worker goroutines")
	timeout := flag.Duration("timeout", 10*time.Second, "maximum execution time")
	flag.CommandLine.Parse(args) //nolint

	*dir = filepath.Clean(*dir)

	if *word == "" {
		fmt.Fprintln(os.Stderr, "error: -word is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		cancel()
	}()

	ctx, timeoutCancel := context.WithTimeout(ctx, *timeout)
	defer timeoutCancel()

	start := time.Now()

	fileCh := make(chan string, 100)
	resultCh := make(chan match, 100)

	var walkWg sync.WaitGroup
	var workerWg sync.WaitGroup

	var scannedFiles atomic.Int64
	var matchCount atomic.Int64
	fileWithMatchMap := make(map[string]bool)
	var fileWithMatchMu sync.Mutex

	printerDone := make(chan struct{})
	go func() {
		for m := range resultCh {
			fmt.Printf("%s:%d: %s\n", m.path, m.line, m.text)
			matchCount.Add(1)
			fileWithMatchMu.Lock()
			fileWithMatchMap[m.path] = true
			fileWithMatchMu.Unlock()
		}
		close(printerDone)
	}()

	walkWg.Add(1)
	go func() {
		defer walkWg.Done()
		defer close(fileCh)
		walkDir(ctx, *dir, fileCh)
	}()

	for i := 0; i < *workers; i++ {
		workerWg.Add(1)
		go worker(ctx, fileCh, resultCh, &workerWg, &scannedFiles, *word)
	}

	walkWg.Wait()
	close(resultCh)
	workerWg.Wait()
	<-printerDone

	elapsed := time.Since(start)
	fileWithMatchMu.Lock()
	fileWithMatchCount := len(fileWithMatchMap)
	fileWithMatchMu.Unlock()

	fmt.Fprintf(os.Stderr, "Scanned %d files, found %d matches in %d files, elapsed %dms\n",
		scannedFiles.Load(), matchCount.Load(), fileWithMatchCount, elapsed.Milliseconds())

	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "cancelled: search interrupted")
		os.Exit(1)
	}
}

// here it combines args and glue the next arg onto it.
func rebuildArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		if strings.HasSuffix(arg, "=") && i+1 < len(args) {
			out = append(out, arg+strings.TrimSpace(args[i+1]))
			i++
			continue
		}
		out = append(out, arg)
	}
	return out
}

func walkDir(ctx context.Context, root string, fileCh chan<- string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return
		default:
		}
		path := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			walkDir(ctx, path, fileCh)
			continue
		}
		if !entry.Type().IsRegular() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Size() > 10<<20 { // skip files larger than 10 MB
			continue
		}
		select {
		case fileCh <- path:
		case <-ctx.Done():
			return
		}
	}
}

// Worker intialized func
func worker(ctx context.Context, fileCh <-chan string, resultCh chan<- match, wg *sync.WaitGroup, scanned *atomic.Int64, word string) {
	defer wg.Done()
	for path := range fileCh {
		select {
		case <-ctx.Done():
			return
		default:
		}
		scanned.Add(1)
		processFile(ctx, path, resultCh, word)
	}
}

// Process file and search for the word
func processFile(ctx context.Context, path string, resultCh chan<- match, word string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if !utf8.Valid(data) {
		return
	}
	lowerWord := strings.ToLower(word)
	scanner := bufio.NewScanner(bytes.NewReader(data))

	// Increase buffer to 1 MB to handle long lines
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		select {
		case <-ctx.Done():
			return
		default:
		}
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), lowerWord) {
			select {
			case resultCh <- match{path: path, line: lineNum, text: line}:
			case <-ctx.Done():
				return
			}
		}
	}
}
