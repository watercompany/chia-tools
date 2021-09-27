package scraper

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func parseLogForMedianProofTime(lines []string, CSVData *[][]string, processDataMap *map[FarmDateMap][]float64, dateIndexMap *map[string]int, csvDataFarmIndex int) error {
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
			proofTimeStr := getNumberValue(line, strings.Index(line, "Time")+6)
			proofTime, err := strconv.ParseFloat(proofTimeStr, 32)
			if err != nil {
				return err
			}

			// add proof time
			(*processDataMap)[FarmDateMap{FarmIndex: csvDataFarmIndex, Date: lineDateStr}] = append((*processDataMap)[FarmDateMap{FarmIndex: csvDataFarmIndex, Date: lineDateStr}], float64(proofTime))

		}

	}
	return nil
}

func getMedian(n ...float64) float64 {
	if len(n) == 0 {
		return 0
	}

	// sort
	sort.Slice(n, func(i, j int) bool { return n[i] < n[j] })

	medianIndex := len(n) / 2

	if len(n)%2 != 0 {
		return n[medianIndex]
	}

	return (n[medianIndex-1] + n[medianIndex]) / 2
}

func processMedianProofTime(CSVData *[][]string, processDataMap *map[FarmDateMap][]float64, dateIndexMap *map[string]int) error {
	for i, farm := range *CSVData {
		if i == 0 {
			continue
		}
		for x := range farm {
			if x == 0 {
				continue
			}
			date := farm[0]

			median := getMedian((*processDataMap)[FarmDateMap{FarmIndex: x, Date: date}]...)
			if median == 0 {
				continue
			}
			newVal := fmt.Sprintf("%.3fs", median)

			// Have to manually add 0 padding
			// because %2.3f doesnt work
			diffLen := 7 - len(newVal)
			if diffLen != 0 {
				for i := 0; i < diffLen; i++ {
					newVal = "0" + newVal
				}
			}

			(*CSVData)[i][x] = newVal
		}
	}

	return nil
}
