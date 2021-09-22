package scraper

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func findProofsFound(filePath string, CSVData *[][]string, dateIndexMap *map[string]int, csvDataFarmIndex int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %s, skipping reading", err)
	}
	defer file.Close()

	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error reading log file: %s, skipping reading", err)
	}

	if stat.Size() > 0 {
		buf := make([]byte, stat.Size())
		_, err = file.ReadAt(buf, int64(0))
		if err == nil {
			lines := strings.Split(strings.ReplaceAll(string(buf), "\r\n", "\n"), "\n")
			err := parseLog(lines, CSVData, dateIndexMap, csvDataFarmIndex)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("error when reading bytes from log file: %s", err)
		}
	}

	return nil
}

func parseLog(lines []string, CSVData *[][]string, dateIndexMap *map[string]int, csvDataFarmIndex int) error {
	s := ""

	for _, line := range lines {
		if len(line) < 23 {
			continue
		}

		s = line[0:23]
		lineDate, err := time.Parse("2006-01-02T15:04:05.000", s)
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
