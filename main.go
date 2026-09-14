package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type MeterRecord struct {
	Date     time.Time
	Import   float64
	Export   float64
	SolarGen float64
}

const dataFile = "solar_readings.txt"

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("==========================================================================")
	fmt.Println(" SOLAR & GRID METER SYSTEM ")
	fmt.Println("==========================================================================")

	// 1. Read all previous records from file
	history, err := loadHistory(dataFile)
	if err != nil {
		fmt.Printf("Error reading history file: %v\n", err)
		return
	}

	// 2. Display period-wise summary if historical records exist
	if len(history) > 0 {
		displayHistorySummary(history)
	} else {
		fmt.Println("\n[No prior reading history found in database]")
	}

	// 3. Prompt for New Current Reading
	fmt.Println("\n--------------------------------------------------------------------------")
	fmt.Println("ENTER NEW READING")
	fmt.Println("--------------------------------------------------------------------------")

	currDateStr := promptInput(reader, "Enter Current Date (YYYY-MM-DD) [Default: Today]: ")
	if currDateStr == "" {
		currDateStr = time.Now().Format("2006-01-02")
	}
	currDate, err := time.Parse("2006-01-02", currDateStr)
	if err != nil {
		fmt.Printf("Error parsing current date: %v\n", err)
		return
	}

	currImport := promptFloat(reader, "Enter Current Imported Units (kWh): ")
	currExport := promptFloat(reader, "Enter Current Exported Units (kWh): ")

	solarGenInput := promptInput(reader, "Enter Total Solar Inverter Yield for this period (kWh) [Press Enter if unknown]: ")
	var solarGen float64
	var hasSolarData bool
	if strings.TrimSpace(solarGenInput) != "" {
		val, err := strconv.ParseFloat(solarGenInput, 64)
		if err == nil && val >= 0 {
			solarGen = val
			hasSolarData = true
		}
	}

	// 4. Processing logic: First run vs. Comparative calculation
	if len(history) == 0 {
		fmt.Println("\n⚠️ Setting this reading as your INITIAL BASELINE.")
		newRec := MeterRecord{
			Date:     currDate,
			Import:   currImport,
			Export:   currExport,
			SolarGen: solarGen,
		}
		if err := appendRecord(dataFile, newRec); err != nil {
			fmt.Printf("Error saving initial record: %v\n", err)
			return
		}
		fmt.Printf("✓ Baseline recorded: %s | Import: %.2f kWh | Export: %.2f kWh | Solar: %.2f kWh\n",
			currDate.Format("02-Jan-2006"), currImport, currExport, solarGen)
		fmt.Println("Run again on your next reading date to compute the next period summary.")
		return
	}

	// Calculate period results relative to the last recorded entry
	lastRecord := history[len(history)-1]

	// Validation for chronological order
	if !currDate.After(lastRecord.Date) {
		fmt.Printf("\nWarning: Current date (%s) is not after previous reading date (%s).\n",
			currDate.Format("2006-01-02"), lastRecord.Date.Format("2006-01-02"))
	}

	importDiff := currImport - lastRecord.Import
	exportDiff := currExport - lastRecord.Export
	netBalance := exportDiff - importDiff
	days := currDate.Sub(lastRecord.Date).Hours() / 24

	// 5. Display Current Period Summary
	fmt.Println("\n==========================================================================")
	fmt.Println(" NEW PERIOD SUMMARY RESULT ")
	fmt.Println("==========================================================================")
	fmt.Printf("Period      : %s to %s (%.0f days)\n",
		lastRecord.Date.Format("02-Jan-2006"), currDate.Format("02-Jan-2006"), days)
	fmt.Printf("Grid Import : %.2f kWh -> %.2f kWh (Diff: +%.2f kWh)\n",
		lastRecord.Import, currImport, importDiff)
	fmt.Printf("Grid Export : %.2f kWh -> %.2f kWh (Diff: +%.2f kWh)\n",
		lastRecord.Export, currExport, exportDiff)
	fmt.Println("--------------------------------------------------------------------------")

	if netBalance >= 0 {
		fmt.Printf("Net Grid Status : NET EXPORT +%.2f kWh (Surplus to Grid)\n", netBalance)
	} else {
		fmt.Printf("Net Grid Status : NET IMPORT %.2f kWh (Drawn from Grid)\n", -netBalance)
	}

	if hasSolarData {
		consumption := (solarGen + importDiff) - exportDiff
		fmt.Println("--------------------------------------------------------------------------")
		fmt.Printf("Solar Generation : %.2f kWh Yield\n", solarGen)
		if days > 0 {
			fmt.Printf("Home Consumption : %.2f kWh (Avg %.2f kWh/day)\n", consumption, consumption/days)
		} else {
			fmt.Printf("Home Consumption : %.2f kWh\n", consumption)
		}
	}
	fmt.Println("==========================================================================\n")

	// 6. Append new record to database
	newRec := MeterRecord{
		Date:     currDate,
		Import:   currImport,
		Export:   currExport,
		SolarGen: solarGen,
	}

	if err := appendRecord(dataFile, newRec); err != nil {
		fmt.Printf("Error updating file: %v\n", err)
		return
	}

	fmt.Printf("✓ New reading saved to %s\n", dataFile)
}

func loadHistory(filename string) ([]MeterRecord, error) {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []MeterRecord
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		d, err1 := time.Parse("2006-01-02", parts[0])
		imp, err2 := strconv.ParseFloat(parts[1], 64)
		exp, err3 := strconv.ParseFloat(parts[2], 64)

		if err1 == nil && err2 == nil && err3 == nil {
			var gen float64
			if len(parts) >= 4 {
				if parsedGen, err4 := strconv.ParseFloat(parts[3], 64); err4 == nil {
					gen = parsedGen
				}
			}

			records = append(records, MeterRecord{
				Date:     d,
				Import:   imp,
				Export:   exp,
				SolarGen: gen,
			})
		}
	}

	return records, scanner.Err()
}

func displayHistorySummary(history []MeterRecord) {
	fmt.Println("\nHISTORICAL PERIOD-WISE SUMMARY")
	fmt.Println("----------------------------------------------------------------------------------------------------------")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "PERIOD\tDAYS\tIMPORT (+)\tEXPORT (+)\tSOLAR YIELD\tHOME CONSUMPTION\tNET GRID STATUS\t")
	fmt.Fprintln(w, "------\t----\t----------\t----------\t-----------\t----------------\t---------------\t")

	for i := 1; i < len(history); i++ {
		prev := history[i-1]
		curr := history[i]

		days := curr.Date.Sub(prev.Date).Hours() / 24
		impDiff := curr.Import - prev.Import
		expDiff := curr.Export - prev.Export
		net := expDiff - impDiff

		periodStr := fmt.Sprintf("%s to %s", prev.Date.Format("02/01/06"), curr.Date.Format("02/01/06"))

		var netStr string
		if net >= 0 {
			netStr = fmt.Sprintf("+%.2f kWh (Export)", net)
		} else {
			netStr = fmt.Sprintf("%.2f kWh (Import)", net)
		}

		solarStr := "-"
		consumptionStr := "-"
		if curr.SolarGen > 0 {
			solarStr = fmt.Sprintf("%.2f kWh", curr.SolarGen)
			consumption := (curr.SolarGen + impDiff) - expDiff
			if days > 0 {
				consumptionStr = fmt.Sprintf("%.2f kWh (%.2f/d)", consumption, consumption/days)
			} else {
				consumptionStr = fmt.Sprintf("%.2f kWh", consumption)
			}
		}

		fmt.Fprintf(w, "%s\t%.0f\t+%.2f kWh\t+%.2f kWh\t%s\t%s\t%s\t\n",
			periodStr, days, impDiff, expDiff, solarStr, consumptionStr, netStr)
	}
	w.Flush()

	last := history[len(history)-1]
	fmt.Printf("\nLast Recorded Meter Baseline (%s): Import = %.2f kWh | Export = %.2f kWh | Solar Yield = %.2f kWh\n",
		last.Date.Format("02-Jan-2006"), last.Import, last.Export, last.SolarGen)
}

func appendRecord(filename string, r MeterRecord) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	line := fmt.Sprintf("%s|%.2f|%.2f|%.2f\n", r.Date.Format("2006-01-02"), r.Import, r.Export, r.SolarGen)
	_, err = file.WriteString(line)
	return err
}

func promptInput(r *bufio.Reader, label string) string {
	fmt.Print(label)
	str, _ := r.ReadString('\n')
	return strings.TrimSpace(str)
}

func promptFloat(r *bufio.Reader, label string) float64 {
	for {
		s := promptInput(r, label)
		val, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return val
		}
		fmt.Println("Invalid input. Please enter a valid numerical value.")
	}
}