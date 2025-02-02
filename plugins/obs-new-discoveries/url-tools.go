package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Custom errors
var (
	ErrInvalidURL     = errors.New("invalid or malformed URL provided")
	ErrRequestTimeout = errors.New("request timed out")
	ErrNoTitle        = errors.New("no title found in webpage")
	ErrHTTPStatus     = errors.New("unexpected HTTP status")
)

const (
	defaultTimeout = 10 * time.Second
	maxBodySize    = 1 << 20 // 1MB
	maxRedirects   = 10
	userAgent      = "Mozilla/5.0 (compatible; URLTitleBot/1.0)"
)

var defaultTransport = &http.Transport{
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	DisableCompression:    false,
	DisableKeepAlives:     false,
	MaxIdleConnsPerHost:   10,
	ResponseHeaderTimeout: 5 * time.Second,
}

// getTitleFromURL fetches the title from a given URL
func getTitleFromURL(u string) (string, error) {
	if u == "" {
		return "", ErrInvalidURL
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrInvalidURL, err)
	}

	// Validate scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("%w: scheme must be http or https", ErrInvalidURL)
	}

	// Validate host
	if parsedURL.Host == "" {
		return "", fmt.Errorf("%w: missing host", ErrInvalidURL)
	}

	// Create a client with optimized settings
	client := &http.Client{
		Timeout:   defaultTimeout,
		Transport: defaultTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			return nil
		},
	}

	// Prepare the request with custom headers
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	// Make the request
	res, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", ErrRequestTimeout
		}
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %d - %s", ErrHTTPStatus, res.StatusCode, res.Status)
	}

	// Limit the response body size
	bodyReader := io.LimitReader(res.Body, maxBodySize)
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract and clean title
	title := strings.TrimSpace(doc.Find("title").First().Text())
	if title == "" {
		return "", ErrNoTitle
	}

	return title, nil
}
