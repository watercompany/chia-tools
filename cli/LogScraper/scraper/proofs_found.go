package scraper

import (
	"fmt"
	"strconv"
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

		lineDate, err := time.Parse(timeFormatFromLogs, s)
		if err != nil {
			continue
		}

		lineDateStr := lineDate.Format(formatTimeStr)
		if strings.Contains(line, "Found 1") {
			currentTotalProofsStr := (*CSVData)[(*dateIndexMap)[lineDateStr]][csvDataFarmIndex]
			if currentTotalProofsStr == valuePlaceholder {
				currentTotalProofsStr = "0"
			} else {
				currentTotalProofsStr = currentTotalProofsStr[3:4]
			}
			currentTotalProofs, err := strconv.Atoi(currentTotalProofsStr)
			if err != nil {
				return err
			}
			(*CSVData)[(*dateIndexMap)[lineDateStr]][csvDataFarmIndex] = fmt.Sprintf("---%v---", currentTotalProofs+1)
		}

	}
	return nil
}
