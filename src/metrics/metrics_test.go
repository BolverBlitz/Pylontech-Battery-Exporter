package metrics

import (
	"testing"

	"pylontech_exporter/src/parser"
)

func TestUpdateBatteryStatMetricsExportsDsgCap(t *testing.T) {
	t.Setenv("PROM_NAMESPACE", "devicemon")
	registry := InitMetrics()

	UpdateBatteryStatMetrics("bat3", parser.BatteryStatStatus{
		DsgCap: 6621177,
	})

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather returned error: %v", err)
	}

	for _, family := range metricFamilies {
		if family.GetName() != "devicemon_battery_stat_dsg_cap" {
			continue
		}

		metrics := family.GetMetric()
		if len(metrics) != 1 {
			t.Fatalf("dsg_cap metric count = %d, want 1", len(metrics))
		}
		if got := metrics[0].GetGauge().GetValue(); got != 6621177 {
			t.Fatalf("dsg_cap = %v, want 6621177", got)
		}
		for _, label := range metrics[0].GetLabel() {
			if label.GetName() == "unit" && label.GetValue() == "bat3" {
				return
			}
		}
		t.Fatal("dsg_cap metric is missing unit=\"bat3\"")
	}

	t.Fatal("devicemon_battery_stat_dsg_cap was not exported")
}
