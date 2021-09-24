package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseLogForMaxProofTime(lines []string, CSVData *[][]string, dateIndexMap *map[string]int, csvDataFarmIndex int) error {
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
		if strings.Contains(line, "proofs") && strings.Contains(line, "Time") && strings.Contains(line, "plots") {
			maxProofTimeStr := getNumberValue(line, strings.Index(line, "Time")+6)
			maxProofTime, err := strconv.ParseFloat(maxProofTimeStr, 32)
			if err != nil {
				return err
			}

			currentmaxProofTimeStr := (*CSVData)[(*dateIndexMap)[lineDateStr]][csvDataFarmIndex]
			if currentmaxProofTimeStr == valuePlaceholder {
				currentmaxProofTimeStr = "0s"
			}
			currentmaxProofTime, err := strconv.ParseFloat(currentmaxProofTimeStr[:len(currentmaxProofTimeStr)-1], 32)
			if err != nil {
				return err
			}

			if currentmaxProofTime == 0 || currentmaxProofTime < maxProofTime {
				newVal := fmt.Sprintf("%.3fs", maxProofTime)

				// Have to manually add 0 padding
				// because %2.3f doesnt work
				diffLen := 7 - len(newVal)
				if diffLen != 0 {
					for i := 0; i < diffLen; i++ {
						newVal = "0" + newVal
					}
				}
				(*CSVData)[(*dateIndexMap)[lineDateStr]][csvDataFarmIndex] = newVal
			}
		}

	}
	return nil
}
