package scraper

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func parseLogForGapsProofTime(lines []string, CSVData *[][]string, processDataMap *map[FarmDateMap][]float64, dateIndexMap *map[string]int, csvDataFarmIndex int) error {
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
			// add time in seconds
			(*processDataMap)[FarmDateMap{FarmIndex: csvDataFarmIndex, Date: lineDateStr}] = append((*processDataMap)[FarmDateMap{FarmIndex: csvDataFarmIndex, Date: lineDateStr}], float64(lineDate.Unix()))

		}

	}
	return nil
}

func getGapsCount(N float64, nArr ...float64) int {
	lenArray := len(nArr)
	if lenArray == 0 {
		return 0
	}

	// sort
	sort.Slice(nArr, func(i, j int) bool { return nArr[i] < nArr[j] })

	fmt.Printf("arr=%v\n\n", nArr)
	var count int = 0
	var prev float64 = 0
	for x, val := range nArr {
		if x == 0 {
			prev = val
			continue
		}

		if val-prev >= N {
			count++
		}

		prev = val
	}

	return count
}

func processGapsProofTime(gapTime float64, CSVData *[][]string, processDataMap *map[FarmDateMap][]float64, dateIndexMap *map[string]int) error {
	for i, farm := range *CSVData {
		if i == 0 {
			continue
		}
		for x := range farm {
			if x == 0 {
				continue
			}
			date := farm[0]

			gapsCount := getGapsCount(gapTime, (*processDataMap)[FarmDateMap{FarmIndex: x, Date: date}]...)
			newVal := fmt.Sprintf("%07v", gapsCount)

			// // Have to manually add 0 padding
			// diffLen := 7 - len(newVal)
			// if diffLen != 0 {
			// 	for i := 0; i < diffLen; i++ {
			// 		newVal = "0" + newVal
			// 	}
			// }

			(*CSVData)[i][x] = newVal
		}
	}

	return nil
}
