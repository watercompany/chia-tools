package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseLogForTotalEligiblePlots(lines []string, CSVData *[][]string, dateIndexMap *map[string]int, csvDataFarmIndex int) error {
	s := ""

	for _, line := range lines {
		if len(line) < 23 {
			continue
		}

		s = line[0:23]
		lineDate, err := time.Parse(timeFormatFromLogs, s)
		if err != nil {
			continue
		}

		lineDateStr := lineDate.Format(formatTimeStr)
		if strings.Contains(line, "plots were eligible") {
			totalEligiblePlotsStr := line[getIndexUntilSpaceToTheLeft(line, strings.Index(line, "plots were eligible")-1) : strings.Index(line, "plots were eligible")-1]
			totalEligiblePlots, err := strconv.Atoi(totalEligiblePlotsStr)
			if err != nil {
				return err
			}

			currentTotalEligiblePlotsStr := (*CSVData)[(*dateIndexMap)[lineDateStr]][csvDataFarmIndex]
			if currentTotalEligiblePlotsStr == valuePlaceholder {
				currentTotalEligiblePlotsStr = "0"
			}
			currentTotalEligiblePlots, err := strconv.Atoi(currentTotalEligiblePlotsStr)
			if err != nil {
				return err
			}

			(*CSVData)[(*dateIndexMap)[lineDateStr]][csvDataFarmIndex] = fmt.Sprintf("%07v", currentTotalEligiblePlots+totalEligiblePlots)

		}

	}
	return nil
}
