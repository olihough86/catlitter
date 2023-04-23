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
	"time"
)

var client = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        48,
		MaxIdleConnsPerHost: 48,
		IdleConnTimeout:     30 * time.Second,
	},
}

func checkURL(baseURL, urlPath, ext string, wg *sync.WaitGroup, sem chan struct{}, validURLs chan<- string) {
	defer wg.Done()

	sem <- struct{}{} // Acquire semaphore
	defer func() { <-sem }()

	fullURL := baseURL + urlPath + ext
	fmt.Println("Trying:", fullURL) // Print the URL being tried

	req, err := http.NewRequest("HEAD", fullURL, nil)
	if err != nil {
		fmt.Printf("Error creating request for %s: %v\n", fullURL, err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error while accessing %s: %v\n", fullURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("Valid:", fullURL) // Print the valid URL
		validURLs <- fullURL
	}
}


func main() {
	dirPath := flag.String("d", "", "Directory path to select a random file from")
	baseURL := flag.String("url", "", "Base URL to prepend to each input URL")
	ext := flag.String("ext", "", "File extension to append to each input URL")
	flag.Parse()

	if *dirPath == "" || *baseURL == "" || *ext == "" {
		fmt.Println("Usage: <program-name> -d <directory-path> -url <base-url> -ext <file-extension>")
		return
	}

	rand.Seed(time.Now().UnixNano())
	files, err := ioutil.ReadDir(*dirPath)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("The specified directory is empty.")
		return
	}

	randomFile := files[rand.Intn(len(files))]
	filePath := filepath.Join(*dirPath, randomFile.Name())
	numWorkers := 48 // Set the number of concurrent Goroutines (2x number of cores)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	outputPath := "valid.txt" // Set the output file name
	outputFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
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

	for scanner.Scan() {
		urlPath := scanner.Text()
		wg.Add(1)
		go checkURL(*baseURL, urlPath, *ext, &wg, sem, validURLs)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error while reading file: %v\n", err)
		return
	}

	wg.Wait()
	close(validURLs)
	fmt.Println("All URLs checked.")

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
