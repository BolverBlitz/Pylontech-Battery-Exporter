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
		"Dsg Cap         :  6621177",
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
	if got.DsgCap != 6621177 {
		t.Fatalf("DsgCap = %v, want 6621177", got.DsgCap)
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

func TestParsePWRCurrentColumnLayout(t *testing.T) {
	lines := []string{
		"pwr",
		"@",
		"Power Volt   Curr   Tempr  Tlow   Tlow.Id  Thigh  Thigh.Id Vlow   Vlow.Id  Vhigh  Vhigh.Id Base.St  Volt.St  Curr.St  Temp.St  Coulomb  Time                 B.V.St   B.T.St  MosTempr M.T.St   SysAlarm.St",
		"1     51516  -1459  32900  29400  12       31300  0        3429   2        3438   1        Dischg   Normal   Normal   Normal   100%     2026-06-18 22:49:12  Normal   Normal  32400    Normal   Normal",
		"4     -      -      -      -      -        -      -        Absent",
	}

	got, err := ParsePWR(lines)
	if err != nil {
		t.Fatalf("ParsePWR returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(ParsePWR) = %d, want 1: %#v", len(got), got)
	}

	status := got[0]
	if status.ID != 1 || status.BaseState != 1 || status.Coulomb != 100 {
		t.Fatalf("parsed identity/state/SOC incorrectly: %#v", status)
	}
	if status.VoltState != "Normal" || status.CurrState != "Normal" || status.TempState != "Normal" {
		t.Fatalf("parsed state columns incorrectly: %#v", status)
	}
	if status.BVState != "Normal" || status.BTState != "Normal" || status.MosTemp != "32400" || status.MTState != "Normal" {
		t.Fatalf("parsed trailing columns incorrectly: %#v", status)
	}
}

func TestParsePWRLegacyColumnLayoutWithoutHeader(t *testing.T) {
	lines := []string{
		"1 51516 -1459 32900 0 0 0 0 Dischg Normal Normal Normal 100% 2026-06-18 22:49:12 Normal Normal 32400 Normal",
	}

	got, err := ParsePWR(lines)
	if err != nil {
		t.Fatalf("ParsePWR returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(ParsePWR) = %d, want 1: %#v", len(got), got)
	}

	status := got[0]
	if status.BaseState != 1 || status.Coulomb != 100 || status.MosTemp != "32400" {
		t.Fatalf("legacy layout parsed incorrectly: %#v", status)
	}
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
