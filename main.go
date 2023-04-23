// main.go

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	dirPath := flag.String("d", "", "Directory path to select a random file from")
	baseURL := flag.String("url", "", "Base URL to prepend to each input URL")
	ext := flag.String("ext", "", "File extension to append to each input URL")
	noExt := flag.Bool("no-ext", false, "Do not append any file extension to input URL")
	treat301AsValid := flag.Bool("301-valid", false, "Treat 301 status code as valid")
	geoNode := flag.Bool("geonode", false, "Use GeoNode rotating proxy")
	flag.Parse()

	if *dirPath == "" || *baseURL == "" || (*ext == "" && !*noExt) {
		fmt.Println("Usage: catlitter -d <directory-path> -url <base-url> -ext <file-extension> [-no-ext]")
		return
	}

	if *ext != "" && *noExt {
		fmt.Println("Error: The -ext and -no-ext flags are exclusive and cannot be used together.")
		return
	}

	var client *http.Client

	if *geoNode {
        config, err := loadProxyConfig("proxy_config.json") // Replace "proxy_config.json" with the path to your config file
        if err != nil {
            fmt.Printf("Failed to load proxy config: %v\n", err)
            os.Exit(1)
        }

		client, err = createGeoNodeClient(config)
		if err != nil {
			fmt.Printf("Failed to create GeoNode client: %v\n", err)
			os.Exit(1)
		}
		} else {
			client = &http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        48,
					MaxIdleConnsPerHost: 48,
					IdleConnTimeout:     30 * time.Second,
				},
			}
		}

	rand.Seed(time.Now().UnixNano())
	files, err := ioutil.ReadDir(*dirPath)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	filteredFiles := make([]os.FileInfo, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			filteredFiles = append(filteredFiles, file)
		}
	}

	if len(filteredFiles) == 0 {
		fmt.Println("The specified directory is empty or contains only subdirectories.")
		return
	}

	randomFile := filteredFiles[rand.Intn(len(filteredFiles))]

	filePath := filepath.Join(*dirPath, randomFile.Name())
	numWorkers := 48 // Set the number of concurrent Goroutines (2x number of cores)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	outputPath := "valid.txt" // Set the output file name
	outputFile, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(file)
	var wg sync.WaitGroup
	sem := make(chan struct{}, numWorkers) // Buffered channel as semaphore
	validURLs := make(chan string)

	go func() {
		for url := range validURLs {
			outputFile.WriteString(url + "\n")
		}
	}()

	startTime := time.Now()
	var totalRequests int64

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			elapsed := time.Since(startTime)
			requests := atomic.LoadInt64(&totalRequests)
			reqPerSec := float64(requests) / elapsed.Seconds()
			fmt.Printf("\rElapsed time: %v, Total requests: %d, Requests/sec: %.2f", elapsed, requests, reqPerSec)
		}
	}()

	for scanner.Scan() {
		urlPath := scanner.Text()
		wg.Add(1)
		go checkURL(*baseURL, urlPath, *ext, &wg, sem, validURLs, &totalRequests, *treat301AsValid, *noExt, client)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error while reading file: %v\n", err)
		return
	}

	wg.Wait()
	close(validURLs)
	fmt.Println("\nAll URLs checked.")

	doneDir := filepath.Join(*dirPath, "done")
	err = os.MkdirAll(doneDir, 0755)
	if err != nil {
		fmt.Printf("Error creating 'done' subdirectory: %v\n", err)
		return
	}

	doneFilePath := filepath.Join(doneDir, randomFile.Name())
	err = os.Rename(filePath, doneFilePath)
	if err != nil {
		fmt.Printf("Error moving processed file to 'done' subdirectory: %v\n", err)
		return
	}

	fmt.Printf("Processed file moved to: %s\n", doneFilePath)
}

