package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Custom errors
var (
	ErrInvalidURL     = errors.New("invalid URL provided")
	ErrRequestTimeout = errors.New("request timed out")
)

// getTitleFromURL fetches the title from a given URL
func getTitleFromURL(url string) (string, error) {
	// Input validation
	if url == "" {
		return "", ErrInvalidURL
	}

	// Create a client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression: false,
		},
	}

	// Make the request
	res, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer res.Body.Close()

	// Check status code
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d - %s", res.StatusCode, res.Status)
	}

	// Limit the response body size to prevent memory issues
	bodyReader := io.LimitReader(res.Body, 1024*1024) // 1MB limit

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract title
	title := doc.Find("title").First().Text()
	if title == "" {
		return "", errors.New("no title found")
	}

	return title, nil
}
