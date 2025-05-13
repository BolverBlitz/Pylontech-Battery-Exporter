package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"pylontech_exporter/src/fetcher"
	"pylontech_exporter/src/metrics"
	"pylontech_exporter/src/parser"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	verbose bool
)

func logVerbose(format string, v ...interface{}) {
	if verbose {
		log.Printf(format, v...)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	refreshSecondsStr := os.Getenv("REFRESH_SECONDS")
	if refreshSecondsStr == "" {
		refreshSecondsStr = "30"
	}
	refreshSeconds, err := strconv.Atoi(refreshSecondsStr)
	if err != nil || refreshSeconds < 1 {
		log.Printf("Invalid REFRESH_SECONDS value '%s', defaulting to 30", refreshSecondsStr)
		refreshSeconds = 30
	}

	verbose = strings.ToLower(os.Getenv("LOG_VERBOSE")) == "true"

	// Initialize Prometheus metrics and get the custom registry
	customRegistry := metrics.InitMetrics()

	// Start HTTP server for Prometheus metrics
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "9100" // fallback default
		}

		// Use HandlerFor with the custom registry
		http.Handle("/metrics", promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{}))
		log.Printf("Starting HTTP server on :%s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	}()

	// Data fetching and processing loop
	ticker := time.NewTicker(time.Duration(refreshSeconds) * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		logVerbose("Fetching and processing device data...")
		pwrUnitCount := processPWRData()
		processBATData(pwrUnitCount)
		logVerbose("Data processing complete. Waiting for next tick.")
	}
}

// processBATData fetches, parses, and updates metrics for BAT command
func processBATData(pwrUnitCount int8) {
	if pwrUnitCount <= 0 {
		log.Println("No power units specified for BAT data processing (pwrUnitCount <= 0).")
		return
	}

	totalRecordsProcessedOverall := 0
	unitsSuccessfullyProcessed := 0

	for unitNum := int8(1); unitNum <= pwrUnitCount; unitNum++ {
		var commandToFetch string
		var unitMetricLabel string

		// This logic for commandToFetch and unitMetricLabel seems to have a slight off-by-one
		// potential in how it's creating labels vs commands for subsequent units.
		// For unitNum = 1: command "bat+1", label "bat1" (Correct)
		// For unitNum = 2: command "bat+2", label "bat2" (Original code had "bat+1", label "bat1" effectively again due to suffix logic)
		// Assuming the intent is that unitNum maps directly to the +N in bat+N and the label suffix.
		suffix := strconv.Itoa(int(unitNum))
		commandToFetch = "bat+" + suffix
		unitMetricLabel = "bat" + suffix

		logVerbose("Fetching BAT data for unit %s (command: %s)...", unitMetricLabel, commandToFetch)
		batLines, err := fetcher.FetchConsoleOutput(commandToFetch)
		if err != nil {
			log.Printf("Error fetching BAT data for unit %s: %v", unitMetricLabel, err)
			metrics.RecordError("bat_fetch_" + unitMetricLabel)
			continue
		}

		logVerbose("Parsing BAT data for unit %s...", unitMetricLabel)
		batDataForUnit, err := parser.ParseBAT(batLines)
		if err != nil {
			log.Printf("Error parsing BAT data for unit %s: %v", unitMetricLabel, err)
			metrics.RecordError("bat_parse_" + unitMetricLabel)
			continue
		}

		if len(batDataForUnit) == 0 {
			log.Printf("No BAT data parsed for unit %s.", unitMetricLabel)
		}

		for _, status := range batDataForUnit {
			metrics.UpdateBatteryMetrics(unitMetricLabel, status)
		}

		if len(batDataForUnit) > 0 {
			logVerbose("Successfully processed %d BAT records for unit %s.", len(batDataForUnit), unitMetricLabel)
		}
		totalRecordsProcessedOverall += len(batDataForUnit)
		unitsSuccessfullyProcessed++
	}

	if unitsSuccessfullyProcessed > 0 {
		logVerbose("Finished processing BAT data for %d unit(s). Total records processed: %d.\n", unitsSuccessfullyProcessed, totalRecordsProcessedOverall)
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

	for _, status := range pwrData {
		metrics.UpdatePowerMetrics(status)
	}

	logVerbose("Successfully processed %d PWR records.\n", len(pwrData))
	return int8(len(pwrData))
}
