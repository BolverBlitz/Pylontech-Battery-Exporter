package main

import (
	"log"
	"net/http"
	"time"

	"bat-monitor/src/fetcher"
	"bat-monitor/src/metrics"
	"bat-monitor/src/parser"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Initialize Prometheus metrics
	metrics.InitMetrics()

	// Start HTTP server for Prometheus metrics
	go func() {
		log.Println("Starting Prometheus metrics server on :9092/metrics")
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9092", nil); err != nil {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	}()

	// Data fetching and processing loop
	ticker := time.NewTicker(30 * time.Second) // Fetch data every 30 seconds
	defer ticker.Stop()

	for {
		<-ticker.C
		log.Println("Fetching and processing device data...")
		// Fetch and process PWR data
		pwrUnitCount := processPWRData()

		// Fetch and process BAT data
		processBATData(pwrUnitCount)
		log.Println("Data processing complete. Waiting for next tick.")
	}
}

// processBATData fetches, parses, and updates metrics for BAT command
func processBATData(pwrUnitCount int8) {
	batLines, err := fetcher.FetchConsoleOutput("bat")
	if err != nil {
		log.Printf("Error fetching BAT data: %v", err)
		metrics.RecordError("bat_fetch")
		return
	}

	batData, err := parser.ParseBAT(batLines)
	if err != nil {
		log.Printf("Error parsing BAT data: %v", err)
		metrics.RecordError("bat_parse")
		return
	}

	if len(batData) == 0 {
		log.Println("No BAT data parsed.")
	}

	// Update Prometheus metrics for BAT data
	for _, status := range batData {
		metrics.UpdateBatteryMetrics(status)
	}
	log.Printf("Successfully processed %d BAT records.\n", len(batData))
}

// processPWRData fetches, parses, and updates metrics for PWR command
func processPWRData() int8 {
	pwrLines, err := fetcher.FetchConsoleOutput("pwr")
	if err != nil {
		log.Printf("Error fetching PWR data: %v", err)
		metrics.RecordError("pwr_fetch")
		return 0
	}

	pwrData, err := parser.ParsePWR(pwrLines)
	if err != nil {
		log.Printf("Error parsing PWR data: %v", err)
		metrics.RecordError("pwr_parse")
		return 0
	}

	if len(pwrData) == 0 {
		log.Println("No PWR data parsed.")
		return 0
	}

	// Update Prometheus metrics for PWR data
	for _, status := range pwrData {
		metrics.UpdatePowerMetrics(status)
	}

	log.Printf("Successfully processed %d PWR records.\n", len(pwrData))
	return int8(len(pwrData))
}
