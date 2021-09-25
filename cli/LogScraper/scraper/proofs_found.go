package scraper

import (
	"strings"
	"time"
)

func parseLogForProofsFound(lines []string, CSVData *[][]string, dateIndexMap *map[string]int, csvDataFarmIndex int) error {
	s := ""

	for _, line := range lines {
		if len(line) < 23 {
			continue
		}

		s = line[0:23]
		if strings.HasPrefix(line, ".") {
			s = line[1:24]
		}
		lineDate, err := time.Parse(timeFormatFromLogs, s)
		if err != nil {
			continue
		}

		lineDateStr := lineDate.Format(formatTimeStr)
		if strings.Contains(line, "Found 1") {
			(*CSVData)[(*dateIndexMap)[lineDateStr]][csvDataFarmIndex] = proofFound
		}

	}
	return nil
}
