package domain

func (so SystemOverview) TotalHomeConsumption() float64 {
	var totalGen, totalExport, totalImport float64

	for _, r := range so.Records {
		totalGen += r.SolarGen
	}
	for _, s := range so.Summaries {
		totalExport += s.ExportDiff
		totalImport += s.ImportDiff
	}

	return totalGen + totalImport - totalExport
}
