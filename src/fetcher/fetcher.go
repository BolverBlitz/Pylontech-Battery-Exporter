package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// FetchConsoleOutput fetches lines of text from the device's console output.
// It takes a command (e.g., "bat", "pwr") as input.
func FetchConsoleOutput(command string) ([]string, error) {
	ip := os.Getenv("DEVICE_IP")
	if ip == "" {
		return nil, fmt.Errorf("DEVICE_IP not set in environment")
	}
	// Ensure DEVICE_PORT is also configurable, defaulting if not set
	port := os.Getenv("DEVICE_PORT")
	if port == "" {
		port = "80" // Default HTTP port
	}

	requestURL, err := buildRequestURL(ip, port, command)
	if err != nil {
		return nil, err
	}

	// Create an HTTP client with a timeout
	client := http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get data from %s: %w", requestURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code %d from %s", resp.StatusCode, requestURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return splitConsoleLines(string(body)), nil
}

func buildRequestURL(ip, port, command string) (string, error) {
	baseURL := fmt.Sprintf("http://%s:%s/req", ip, port)
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL %s: %w", baseURL, err)
	}

	encodedCommand := strings.ReplaceAll(url.QueryEscape(command), "+", "%20")
	parsedURL.RawQuery = "code=" + encodedCommand
	return parsedURL.String(), nil
}

func splitConsoleLines(body string) []string {
	parts := strings.FieldsFunc(body, func(r rune) bool {
		return r == '\r' || r == '\n'
	})

	lines := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmedLine := strings.TrimSpace(part)
		if trimmedLine != "" {
			lines = append(lines, trimmedLine)
		}
	}
	return lines
}
