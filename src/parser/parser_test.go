package parser

import "testing"

func TestParseSTATFirmwareLabelValueOutput(t *testing.T) {
	lines := []string{
		"stat 2",
		"@",
		"Device address           2",
		"SOH Times       :      282",
		"CYCLE Times     :      123",
		"SOH             :       99",
		"ChgCurr 0~0.2C Secs      :   580588",
		"ChgCurr 0.2C~0.5C Secs   :   161292",
		"ChgCurr 0.5C~0.8C Secs   :      234",
		"ChgCurr 0.8C~1C Secs     :        0",
		"ChgCurr 1C above Secs    :        0",
		"DsgCurr 0~0.2C Secs      :  1724862",
		"DsgCurr 0.2C~0.5C Secs   :        2",
		"DsgCurr 0.5C~0.8C Secs   :        0",
		"DsgCurr 0.8C~1C Secs     :        0",
		"DsgCurr 1C above Secs    :        0",
		"Soc 0~20% Secs           :    64802",
		"Soc 20~60% Secs          :  1058316",
		"Soc 60% above Secs       :  1760878",
		"Soc Low Thd1 Sec         :        0",
	}

	got, err := ParseSTAT(lines)
	if err != nil {
		t.Fatalf("ParseSTAT returned error: %v", err)
	}

	if got.Cycles != 123 {
		t.Fatalf("Cycles = %v, want 123", got.Cycles)
	}
	if got.SOH != 99 {
		t.Fatalf("SOH = %v, want 99", got.SOH)
	}

	wantChg := map[string]float64{
		"0-0.2c":    580588,
		"0.2c-0.5c": 161292,
		"0.5c-0.8c": 234,
		"0.8c-1c":   0,
		"gt1c":      0,
	}
	assertFloatMap(t, "ChgCurrSec", got.ChgCurrSec, wantChg)

	wantDsg := map[string]float64{
		"0-0.2c":    1724862,
		"0.2c-0.5c": 2,
		"0.5c-0.8c": 0,
		"0.8c-1c":   0,
		"gt1c":      0,
	}
	assertFloatMap(t, "DsgCurrSec", got.DsgCurrSec, wantDsg)

	wantSoc := map[string]float64{
		"0-20":  64802,
		"20-60": 1058316,
		"gt60":  1760878,
	}
	assertFloatMap(t, "SocSec", got.SocSec, wantSoc)
}

func assertFloatMap(t *testing.T, name string, got, want map[string]float64) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d: %#v", name, len(got), len(want), got)
	}

	for key, wantValue := range want {
		if gotValue, ok := got[key]; !ok || gotValue != wantValue {
			t.Fatalf("%s[%q] = %v (exists %v), want %v", name, key, gotValue, ok, wantValue)
		}
	}
}
