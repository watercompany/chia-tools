package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseLogForTotalPlots(lines []string, CSVData *[][]string, dateIndexMap *map[string]int, csvDataFarmIndex int) error {
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
		if strings.Contains(line, "proofs") && strings.Contains(line, "Total") && strings.Contains(line, "plots") {
			totalPlotsStr := getNumberValue(line, strings.Index(line, "Total")+6)
			totalPlots, err := strconv.Atoi(totalPlotsStr)
			if err != nil {
				return err
			}

			currentTotalPlotsStr := (*CSVData)[(*dateIndexMap)[lineDateStr]][csvDataFarmIndex]
			if currentTotalPlotsStr == valuePlaceholder {
				currentTotalPlotsStr = "0"
			}
			currentTotalPlots, err := strconv.Atoi(currentTotalPlotsStr)
			if err != nil {
				return err
			}

			if currentTotalPlots == 0 || currentTotalPlots > totalPlots {
				(*CSVData)[(*dateIndexMap)[lineDateStr]][csvDataFarmIndex] = fmt.Sprintf("%07v", totalPlots)
			}
		}

	}
	return nil
}
