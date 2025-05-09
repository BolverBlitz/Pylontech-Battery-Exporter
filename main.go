package main

import (
	"log"
	"net/http"
	"strconv"
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
	// Check if there are any power units to process
	if pwrUnitCount <= 0 {
		log.Println("No power units specified for BAT data processing (pwrUnitCount <= 0).")
		return
	}

	totalRecordsProcessedOverall := 0
	unitsSuccessfullyProcessed := 0

	// Loop through each power unit, from 1 to pwrUnitCount
	// unitNum is 1-indexed for user-friendliness (Unit 1, Unit 2, ...)
	for unitNum := int8(1); unitNum <= pwrUnitCount; unitNum++ {
		var commandToFetch string
		var unitMetricLabel string // Label for metrics, e.g., "bat0", "bat1"

		if unitNum == 1 {
			commandToFetch = "bat+1"
			unitMetricLabel = "bat1"
		} else {
			// Subsequent units (Unit 2, Unit 3, ...) use "bat+<index>"
			// For Unit 2 (unitNum=2), command is "bat+1", label "bat1"
			// For Unit 3 (unitNum=3), command is "bat+2", label "bat2"
			suffix := strconv.Itoa(int(unitNum))
			commandToFetch = "bat+" + suffix
			unitMetricLabel = "bat" + suffix
		}

		log.Printf("Fetching BAT data for unit %s (command: %s)...", unitMetricLabel, commandToFetch)
		batLines, err := fetcher.FetchConsoleOutput(commandToFetch)
		if err != nil {
			log.Printf("Error fetching BAT data for unit %s: %v", unitMetricLabel, err)
			metrics.RecordError("bat_fetch_" + unitMetricLabel)
			continue
		}

		log.Printf("Parsing BAT data for unit %s...", unitMetricLabel)
		batDataForUnit, err := parser.ParseBAT(batLines)
		if err != nil {
			log.Printf("Error parsing BAT data for unit %s: %v", unitMetricLabel, err)
			metrics.RecordError("bat_parse_" + unitMetricLabel)
			continue
		}

		if len(batDataForUnit) == 0 {
			log.Printf("No BAT data parsed for unit %s.", unitMetricLabel)
		}

		// Update Prometheus metrics for each battery status record from this unit
		for _, status := range batDataForUnit {
			metrics.UpdateBatteryMetrics(unitMetricLabel, status)
		}

		if len(batDataForUnit) > 0 {
			log.Printf("Successfully processed %d BAT records for unit %s.", len(batDataForUnit), unitMetricLabel)
		}
		totalRecordsProcessedOverall += len(batDataForUnit)
		unitsSuccessfullyProcessed++
	}

	// Final summary log
	if unitsSuccessfullyProcessed > 0 {
		log.Printf("Finished processing BAT data for %d unit(s). Total records processed: %d.\n", unitsSuccessfullyProcessed, totalRecordsProcessedOverall)
	} else if pwrUnitCount > 0 {
		log.Println("Attempted to process BAT data, but no units were successfully fetched or parsed.")
	}
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
