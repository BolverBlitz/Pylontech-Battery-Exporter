package parser

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// BatteryStatus holds the parsed data for a single battery entry.
type BatteryStatus struct {
	ID        int    `json:"id"`
	Volt      int    `json:"volt"` // Voltage in mV
	Curr      int    `json:"curr"` // Current in mA
	Temp      int    `json:"temp"`
	BaseState int8   `json:"base_state"` // 0: Charge, 1: Dischg, 2: Idle, 3: Balance
	VoltState string `json:"volt_state"`
	CurrState string `json:"curr_state"`
	TempState string `json:"temp_state"`
	SOC       int8   `json:"soc"`     // State of Charge in %
	Coulomb   int    `json:"coulomb"` // Remaining capacity in mAH
	BAL       string `json:"bal"`     // Balance status (e.g., "0000000000000000")
}

// PowerStatus holds the parsed data for a single power supply entry.
type PowerStatus struct {
	ID        int    `json:"id"`
	Volt      int    `json:"volt"` // Voltage in mV
	Curr      int    `json:"curr"` // Current in mA
	Temp      int    `json:"temp"`
	BaseState int8   `json:"base_state"`
	VoltState string `json:"volt_state"`
	CurrState string `json:"curr_state"`
	TempState string `json:"temp_state"`
	Coulomb   int8   `json:"coulomb"`
	BVState   string `json:"bv_state"`
	BTState   string `json:"bt_state"`
	MosTemp   string `json:"mos_temp"`
	MTState   string `json:"mt_state"`
}

// baseStateMap maps string representations of base states to their int8 values.
var baseStateMap = map[string]int8{
	"Charge":  0,
	"Dischg":  1,
	"Idle":    2,
	"Balance": 3,
	"N/A":     -1, // Placeholder for unknown or not applicable states
}

// parseSOC converts a string like "85%" to an int8 value 85.
func parseSOC(s string) (int8, error) {
	s = strings.TrimSuffix(s, "%")
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse SOC '%s': %w", s, err)
	}
	return int8(n), nil
}

// parseCoulomb expects the numeric part and the unit (though unit isn't used here).
func parseCoulomb(s1 string, s2 string /* unit */) (int, error) {
	s := strings.TrimSpace(s1) // Trim any spaces from the number part
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Coulomb '%s': %w", s, err)
	}
	return n, nil
}

func parseBaseState(s string) int8 {
	if val, ok := baseStateMap[s]; ok {
		return val
	}
	return -1 // Unknown state
}

func parseInt(s string, fieldName string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s '%s': %w", fieldName, s, err)
	}
	return n, nil
}

// ParseBAT parses the raw lines from the 'bat' command output.
func ParseBAT(lines []string) ([]BatteryStatus, error) {
	var results []BatteryStatus
	// Regex to identify data lines. Example: "0   3750  0    301 Charge Normal Normal Normal 85% 3450 mAH 0000000000000000"
	// It should match lines starting with numbers, followed by various fields.
	// Adjust regex if header lines or other non-data lines are present and need skipping.
	dataRegex := regexp.MustCompile(`^\s*\d+\s+\d+`) // Matches lines starting with at least two numbers (ID, Volt)

	for lineIdx, line := range lines { // Added lineIdx for logging
		line = strings.TrimSpace(line)
		if !dataRegex.MatchString(line) || line == "" {
			// log.Printf("Skipping non-data line (BAT) or empty line: '%s'", line) // Example logging
			continue // Skip header or malformed lines
		}

		fields := strings.Fields(line)
		// Expected fields: ID, Volt, Curr, Temp, BaseState, VoltState, CurrState, TempState, SOC, CoulombVal, CoulombUnit, BAL
		if len(fields) < 12 { // Ensure enough fields are present
			log.Printf("Skipping line %d (BAT) due to insufficient fields (got %d, expected at least 12): '%s'", lineIdx+1, len(fields), line)
			continue
		}

		var status BatteryStatus
		var err error

		status.ID, err = parseInt(fields[0], "BAT ID")
		if err != nil {
			log.Printf("Error parsing BAT ID on line %d: %v. Line: '%s'", lineIdx+1, err, line)
			continue
		}

		status.Volt, err = parseInt(fields[1], "BAT Volt") // Assuming mV
		if err != nil {
			log.Printf("Error parsing BAT Volt for ID %d on line %d: %v. Line: '%s'", status.ID, lineIdx+1, err, line)
			continue
		}

		status.Curr, err = parseInt(fields[2], "BAT Curr") // Assuming mA
		if err != nil {
			log.Printf("Error parsing BAT Curr for ID %d on line %d: %v. Line: '%s'", status.ID, lineIdx+1, err, line)
			continue
		}

		// Temperature is in 0.1 C, e.g., "301" means 30.1 C
		status.Temp, err = parseInt(fields[3], "BAT Temp")
		if err != nil {
			log.Printf("Error parsing BAT Temp for ID %d on line %d: %v. Line: '%s'", status.ID, lineIdx+1, err, line)
			continue
		}

		status.BaseState = parseBaseState(fields[4])
		status.VoltState = fields[5]
		status.CurrState = fields[6]
		status.TempState = fields[7]

		status.SOC, err = parseSOC(fields[8])
		if err != nil {
			log.Printf("Warning parsing SOC for BAT ID %d on line %d: %v. Line: '%s'", status.ID, lineIdx+1, err, line)
			status.SOC = -1 // Indicate parsing failure for SOC
		}

		// Coulomb parsing: fields[9] is value, fields[10] is unit "mAH"
		status.Coulomb, err = parseCoulomb(fields[9], fields[10])
		if err != nil {
			log.Printf("Warning parsing Coulomb for BAT ID %d on line %d: %v. Line: '%s'", status.ID, lineIdx+1, err)
			status.Coulomb = -1 // Indicate parsing failure
		}

		status.BAL = fields[11]

		results = append(results, status)
	}
	if len(results) == 0 && len(lines) > 0 {
		// Check if any lines were data-like but failed parsing
		foundDataLikeLine := false
		for _, line := range lines {
			if dataRegex.MatchString(strings.TrimSpace(line)) {
				foundDataLikeLine = true
				break
			}
		}
		if foundDataLikeLine {
			log.Println("Warning: No BAT data records were successfully parsed, though some lines appeared to be data lines. Check format and parsing logic.")
		} else if len(lines) > 0 {
			// log.Println("Note: No BAT data lines matched the expected format.") // Less critical if lines are just headers etc.
		}
	}
	return results, nil
}

// ParsePWR parses the raw lines from the 'pwr' command output.
func ParsePWR(lines []string) ([]PowerStatus, error) {
	var results []PowerStatus
	// Regex for data lines, e.g., "0  5000   0    250  ..."
	// Based on field access, it seems to expect a line that can be split into many fields.
	dataRegex := regexp.MustCompile(`^\s*\d+\s+`) // Matches lines starting with a number (ID)

	for lineIdx, line := range lines { // Added lineIdx for logging
		line = strings.TrimSpace(line)
		// Skip lines explicitly containing "Absent" or if they don't look like data lines or are empty.
		if !dataRegex.MatchString(line) || strings.Contains(line, "Absent") || line == "" {
			// log.Printf("Skipping non-data, 'Absent', or empty line (PWR): '%s'", line)
			continue
		}

		fields := strings.Fields(line)
		// Expected fields based on indices used: ID(0), Volt(1), Curr(2), Temp(3), ..., BaseState(8), VoltState(9), CurrState(10), TempState(11), SOC/Coulomb(12), Time_p1(13), Time_p2(14), BVState(15), BTState(16), MosTemp(17), MTState(18)
		if len(fields) < 19 {
			log.Printf("Skipping line %d (PWR) due to insufficient fields (got %d, expected at least 19): '%s'", lineIdx+1, len(fields), line)
			continue
		}

		var status PowerStatus
		var err error

		status.ID, err = parseInt(fields[0], "PWR ID")
		if err != nil {
			log.Printf("Error parsing PWR ID on line %d: %v. Line: '%s'", lineIdx+1, err, line)
			continue
		}

		status.Volt, err = parseInt(fields[1], "PWR Volt") // Assuming mV
		if err != nil {
			log.Printf("Error parsing PWR Volt for ID %d on line %d: %v. Line: '%s'", status.ID, lineIdx+1, err, line)
			continue
		}

		status.Curr, err = parseInt(fields[2], "PWR Curr") // Assuming mA
		if err != nil {
			log.Printf("Error parsing PWR Curr for ID %d on line %d: %v. Line: '%s'", status.ID, lineIdx+1, err, line)
			continue
		}

		status.Temp, err = parseInt(fields[3], "PWR Temp (Board)") // Temp in 0.1C
		if err != nil {
			log.Printf("Error parsing PWR Temp for ID %d on line %d: %v. Line: '%s'", status.ID, lineIdx+1, err, line)
			continue
		}

		status.BaseState = parseBaseState(fields[8]) // BaseState is field 8
		status.VoltState = fields[9]
		status.CurrState = fields[10]
		status.TempState = fields[11] // State for board temperature

		socVal, err := parseSOC(fields[12]) // SOC is field 12
		if err != nil {
			log.Printf("Warning parsing SOC/Coulomb for PWR ID %d on line %d: %v. Line: '%s'", status.ID, lineIdx+1, err)
			status.Coulomb = -1 // Indicate parsing failure
		} else {
			status.Coulomb = socVal // Storing SOC (as int8) into Coulomb field as per struct def
		}

		status.BVState = fields[15]
		status.BTState = fields[16]

		status.MosTemp = fields[17] // MosTemp is field 17 (string, in 0.1C)
		// No direct parsing to int here, kept as string. Conversion happens in metrics.go

		status.MTState = fields[18]

		results = append(results, status)
	}
	if len(results) == 0 && len(lines) > 0 {
		foundDataLikeLine := false
		for _, line := range lines {
			if dataRegex.MatchString(strings.TrimSpace(line)) && !strings.Contains(line, "Absent") {
				foundDataLikeLine = true
				break
			}
		}
		if foundDataLikeLine {
			log.Println("Warning: No PWR data records were successfully parsed, though some lines appeared to be data lines. Check format and parsing logic.")
		} else if len(lines) > 0 {
			// log.Println("Note: No PWR data lines matched the expected format or were not 'Absent'.")
		}
	}
	return results, nil
}
