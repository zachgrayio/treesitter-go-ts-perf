package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

const MaxWorkers = 12

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalf("Usage: %s [--cpuprofile file] <directory> [<directory>...]", os.Args[0])
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var files []string
	for _, dir := range flag.Args() {
		dirFiles, err := getTSFiles(dir)
		if err != nil {
			log.Fatalf("Failed to get TypeScript files from %s: %v", dir, err)
		}
		files = append(files, dirFiles...)
	}

	var wg sync.WaitGroup
	fileChan := make(chan string, len(files))
	resultChan := make(chan string, len(files))

	for i := 0; i < MaxWorkers; i++ {
		wg.Add(1)
		go worker(fileChan, resultChan, &wg)
	}

	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		fmt.Println(result)
	}
}

func worker(fileChan <-chan string, resultChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	for file := range fileChan {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Failed to read file %s: %v", file, err)
			continue
		}

		start := time.Now()
		ctx := context.Background()
		tree, err := parser.ParseCtx(ctx, nil, content)
		duration := time.Since(start)

		if err != nil {
			log.Printf("Failed to parse file %s: %v", file, err)
			continue
		}

		elementCount := countElements(tree.RootNode())
		resultChan <- fmt.Sprintf("Parsed %s in %v with %d elements", file, duration, elementCount)

		tree.Close()
	}
}

func countElements(node *sitter.Node) int {
	count := 1
	for i := 0; i < int(node.ChildCount()); i++ {
		count += countElements(node.Child(i))
	}
	return count
}

func getTSFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".ts" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
