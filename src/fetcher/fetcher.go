package fetcher

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
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

	url := fmt.Sprintf("http://%s:%s/req?code=%s", ip, port, command)

	// Create an HTTP client with a timeout
	client := http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get data from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code %d from %s", resp.StatusCode, url)
	}

	var lines []string
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			lines = append(lines, trimmedLine)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return lines, fmt.Errorf("error reading response body: %w", err)
		}
	}
	return lines, nil
}
