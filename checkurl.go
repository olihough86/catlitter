package main

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
)

func checkURL(baseURL, urlPath, ext string, wg *sync.WaitGroup, sem chan struct{}, validURLs chan<- string, totalRequests *int64, treat301AsValid, noExt bool, client *http.Client) {
    defer wg.Done()

    sem <- struct{}{} // Acquire semaphore
    defer func() { <-sem }()

    fullURL := baseURL + urlPath
    if !noExt {
        fullURL += ext
    }

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

    // Check for 200 status code and not redirected to removed.png
    if resp.StatusCode == 200 && resp.Request.URL.String() != "https://i.imgur.com/removed.png" {
        fmt.Println("Valid:", fullURL) // Print the valid URL
        validURLs <- fullURL
    }

    atomic.AddInt64(totalRequests, 1)
}