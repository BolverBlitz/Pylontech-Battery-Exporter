package metrics

import (
	"log"
	"os"
	"strconv"
	"strings"

	"pylontech_exporter/src/parser"

	"github.com/prometheus/client_golang/prometheus"
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

func getNamespace() string {
	ns := os.Getenv("PROM_NAMESPACE")
	if ns == "" {
		ns = "default" // fallback if not set
	}
	return ns
}

// InitMetrics initializes all Prometheus metrics and returns a custom registry.
func InitMetrics() *prometheus.Registry {
	namespace := getNamespace()
	reg := prometheus.NewRegistry() // Create a new custom registry

	scrapeErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "scraper",
			Name:      "errors_total",
			Help:      "Total number of errors encountered during data scraping or parsing.",
		},
		[]string{"type"}, // e.g., "bat_fetch", "pwr_parse"
	)
	reg.MustRegister(scrapeErrors)

	// --- Battery Metrics Initialization ---
	batteryVolt = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "volt",
			Help:      "Battery voltage in millivolts.",
		},
		[]string{"unit", "id"},
	)
	reg.MustRegister(batteryVolt)

	batteryCurr = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "curr",
			Help:      "Battery current in milliamps.",
		},
		[]string{"unit", "id"},
	)
	reg.MustRegister(batteryCurr)

	batteryTemp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "temp_celsius",
			Help:      "Battery temperature in degrees Celsius. Assumes input is milli-degrees C (e.g., 17000 -> 17.0 C).",
		},
		[]string{"unit", "id"},
	)
	reg.MustRegister(batteryTemp)

	batteryBaseState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "base_state",
			Help:      "Battery base state code (0: Charge, 1: Dischg, 2: Idle, 3: Balance, -1: Unknown).",
		},
		[]string{"unit", "id"},
	)
	reg.MustRegister(batteryBaseState)

	batterySOC = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "soc",
			Help:      "Battery State of Charge in percent.",
		},
		[]string{"unit", "id"},
	)
	reg.MustRegister(batterySOC)

	batteryCoulomb = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "coulomb",
			Help:      "Battery remaining capacity in milliampere-hours.",
		},
		[]string{"unit", "id"},
	)
	reg.MustRegister(batteryCoulomb)

	batteryBalanceActiveCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "battery",
			Name:      "bal_active_count",
			Help:      "Number of active balancing channels. If BAL is 'N' or similar, this will be 0.",
		},
		[]string{"unit", "id"},
	)
	reg.MustRegister(batteryBalanceActiveCount)

	// --- Power Supply Metrics Initialization ---
	powerVolt = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "volt",
			Help:      "Power supply voltage in millivolts.",
		},
		[]string{"id"},
	)
	reg.MustRegister(powerVolt)

	powerCurr = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "curr",
			Help:      "Power supply current in milliamps.",
		},
		[]string{"id"},
	)
	reg.MustRegister(powerCurr)

	powerBoardTemp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "temp_celsius",
			Help:      "Power supply board temperature in degrees Celsius. Assumes input is milli-degrees C.",
		},
		[]string{"id"},
	)
	reg.MustRegister(powerBoardTemp)

	powerBaseState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "base_state",
			Help:      "Power supply base state code (e.g., 0: Charge, 1: Dischg, 2: Idle, -1: N/A).",
		},
		[]string{"id"},
	)
	reg.MustRegister(powerBaseState)

	powerSOC = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "soc_percent",
			Help:      "Power supply State of Charge or equivalent percentage (from 'Coulomb' field).",
		},
		[]string{"id"},
	)
	reg.MustRegister(powerSOC)

	powerMosTemp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "power",
			Name:      "mos_temp_celsius",
			Help:      "Power supply MOS temperature in degrees Celsius. Assumes input is milli-degrees C if numeric.",
		},
		[]string{"id"},
	)
	reg.MustRegister(powerMosTemp)

	return reg
}

// UpdateBatteryMetrics updates Prometheus gauges with the latest battery status.
func UpdateBatteryMetrics(unitLabel string, status parser.BatteryStatus) {
	idStr := strconv.Itoa(status.ID)

	batteryVolt.WithLabelValues(unitLabel, idStr).Set(float64(status.Volt))
	batteryCurr.WithLabelValues(unitLabel, idStr).Set(float64(status.Curr))
	batteryTemp.WithLabelValues(unitLabel, idStr).Set(float64(status.Temp) / 1000.0)
	batteryBaseState.WithLabelValues(unitLabel, idStr).Set(float64(status.BaseState))
	batterySOC.WithLabelValues(unitLabel, idStr).Set(float64(status.SOC))
	batteryCoulomb.WithLabelValues(unitLabel, idStr).Set(float64(status.Coulomb))

	activeBalanceChannels := 0
	if status.BAL == "Y" {
		activeBalanceChannels = 1
	} else if status.BAL != "" && status.BAL != "N" {
		activeBalanceChannels = strings.Count(status.BAL, "1")
	}
	batteryBalanceActiveCount.WithLabelValues(unitLabel, idStr).Set(float64(activeBalanceChannels))
}

// UpdatePowerMetrics updates Prometheus gauges with the latest power supply status.
func UpdatePowerMetrics(status parser.PowerStatus) {
	idStr := strconv.Itoa(status.ID)

	powerVolt.WithLabelValues(idStr).Set(float64(status.Volt))
	powerCurr.WithLabelValues(idStr).Set(float64(status.Curr))
	powerBoardTemp.WithLabelValues(idStr).Set(float64(status.Temp) / 1000.0)
	powerBaseState.WithLabelValues(idStr).Set(float64(status.BaseState))
	powerSOC.WithLabelValues(idStr).Set(float64(status.Coulomb))

	if mosTempFloat, err := strconv.ParseFloat(status.MosTemp, 64); err == nil {
		powerMosTemp.WithLabelValues(idStr).Set(mosTempFloat / 10.0)
	} else {
		log.Printf("Could not parse MosTemp string '%s' to float for power_id %s: %v", status.MosTemp, idStr, err)
	}
}

// RecordError increments the error counter for a given type.
func RecordError(errorType string) {
	scrapeErrors.WithLabelValues(errorType).Inc()
}
