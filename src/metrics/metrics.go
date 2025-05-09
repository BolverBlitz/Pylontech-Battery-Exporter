package metrics

import (
	"log"
	"strconv"
	"strings"

	"bat-monitor/src/parser" // Assuming parser package is correctly located

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// General metric for tracking errors
	scrapeErrors *prometheus.CounterVec

	// Battery Metrics
	batteryVolt               *prometheus.GaugeVec
	batteryCurr               *prometheus.GaugeVec
	batteryTemp               *prometheus.GaugeVec
	batteryBaseState          *prometheus.GaugeVec
	batterySOC                *prometheus.GaugeVec
	batteryCoulomb            *prometheus.GaugeVec
	batteryBalanceActiveCount *prometheus.GaugeVec

	// Power Supply Metrics
	powerVolt      *prometheus.GaugeVec
	powerCurr      *prometheus.GaugeVec
	powerBoardTemp *prometheus.GaugeVec
	powerBaseState *prometheus.GaugeVec
	powerSOC       *prometheus.GaugeVec
	powerMosTemp   *prometheus.GaugeVec
)

// InitMetrics initializes all Prometheus metrics.
func InitMetrics() {
	namespace := "devicemon" // Namespace for all metrics

	scrapeErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "scraper",
			Name:      "errors_total",
			Help:      "Total number of errors encountered during data scraping or parsing.",
		},
		[]string{"type"}, // e.g., "bat_fetch", "pwr_parse"
	)

	// --- Battery Metrics Initialization ---
	batteryVolt = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "volt", // Matches struct field name
			Help:      "Battery voltage in millivolts.",
		},
		[]string{"id"}, // Using "id" to be generic, maps to battery_id
	)

	batteryCurr = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "curr", // Matches struct field name
			Help:      "Battery current in milliamps.",
		},
		[]string{"id"},
	)

	batteryTemp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "temp_celsius", // Clarified unit
			Help:      "Battery temperature in degrees Celsius. Assumes input is milli-degrees C (e.g., 17000 -> 17.0 C).",
		},
		[]string{"id"},
	)

	batteryBaseState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "base_state", // Matches struct field name
			Help:      "Battery base state code (0: Charge, 1: Dischg, 2: Idle, 3: Balance, -1: Unknown).",
		},
		[]string{"id"},
	)

	batterySOC = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "soc", // Matches struct field name
			Help:      "Battery State of Charge in percent.",
		},
		[]string{"id"},
	)

	batteryCoulomb = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "coulomb", // Matches struct field name
			Help:      "Battery remaining capacity in milliampere-hours.",
		},
		[]string{"id"},
	)

	batteryBalanceActiveCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "bal_active_count",
			Help:      "Number of active balancing channels. If BAL is 'N' or similar, this will be 0.",
		},
		[]string{"id"},
	)

	// --- Power Supply Metrics Initialization ---
	powerVolt = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "volt", // Matches struct field name
			Help:      "Power supply voltage in millivolts.",
		},
		[]string{"id"}, // Using "id" to be generic, maps to power_id
	)

	powerCurr = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "curr", // Matches struct field name
			Help:      "Power supply current in milliamps.",
		},
		[]string{"id"},
	)

	powerBoardTemp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "temp_celsius", // Board temperature, clarified unit
			Help:      "Power supply board temperature in degrees Celsius. Assumes input is milli-degrees C.",
		},
		[]string{"id"},
	)

	powerBaseState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "base_state", // Matches struct field name
			Help:      "Power supply base state code (e.g., 0: Charge, 1: Dischg, 2: Idle, -1: N/A).",
		},
		[]string{"id"},
	)

	powerSOC = promauto.NewGaugeVec( // Was 'Coulomb' in PowerStatus struct, but represents SOC %
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "soc_percent", // SoC Charge Percentage
			Help:      "Power supply State of Charge or equivalent percentage (from 'Coulomb' field).",
		},
		[]string{"id"},
	)

	powerMosTemp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "mos_temp_celsius", // milli-degrees C
			Help:      "Power supply MOS temperature in degrees Celsius. Assumes input is milli-degrees C if numeric.",
		},
		[]string{"id"},
	)
}

// UpdateBatteryMetrics updates Prometheus gauges with the latest battery status.
func UpdateBatteryMetrics(status parser.BatteryStatus) {
	idStr := strconv.Itoa(status.ID)

	batteryVolt.WithLabelValues(idStr).Set(float64(status.Volt))
	batteryCurr.WithLabelValues(idStr).Set(float64(status.Curr))
	batteryTemp.WithLabelValues(idStr).Set(float64(status.Temp) / 1000.0)

	batteryBaseState.WithLabelValues(idStr).Set(float64(status.BaseState))
	batterySOC.WithLabelValues(idStr).Set(float64(status.SOC))
	batteryCoulomb.WithLabelValues(idStr).Set(float64(status.Coulomb))

	// Handle BAL field: "Y" means 1 active, "N" means 0, otherwise count "1"s.
	activeBalanceChannels := 0
	if status.BAL == "Y" {
		activeBalanceChannels = 1 // Indicates balancing is active
	} else if status.BAL != "" && status.BAL != "N" {
		activeBalanceChannels = strings.Count(status.BAL, "1")
	}
	// If status.BAL is "N" or empty, activeBalanceChannels remains 0.
	batteryBalanceActiveCount.WithLabelValues(idStr).Set(float64(activeBalanceChannels))
}

// UpdatePowerMetrics updates Prometheus gauges with the latest power supply status.
func UpdatePowerMetrics(status parser.PowerStatus) {
	idStr := strconv.Itoa(status.ID)

	powerVolt.WithLabelValues(idStr).Set(float64(status.Volt))
	powerCurr.WithLabelValues(idStr).Set(float64(status.Curr))
	powerBoardTemp.WithLabelValues(idStr).Set(float64(status.Temp) / 1000.0)

	powerBaseState.WithLabelValues(idStr).Set(float64(status.BaseState))
	powerSOC.WithLabelValues(idStr).Set(float64(status.Coulomb)) // Coulomb field in PowerStatus is used as SOC %

	if mosTempFloat, err := strconv.ParseFloat(status.MosTemp, 64); err == nil {
		// Assuming status.MosTemp string represents value in 0.1°C
		powerMosTemp.WithLabelValues(idStr).Set(mosTempFloat / 10.0) // Corrected from / 1000.0
	} else {
		log.Printf("Could not parse MosTemp string '%s' to float for power_id %s: %v", status.MosTemp, idStr, err)
	}
}

// RecordError increments the error counter for a given type.
func RecordError(errorType string) {
	scrapeErrors.WithLabelValues(errorType).Inc()
}
