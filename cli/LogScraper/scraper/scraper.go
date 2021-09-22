package scraper

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	proofFound    = "---1---"
	formatTimeStr = "2006-01-02"
)

func countFarmFolders(logDir string) (int, error) {
	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, file := range files {
		if file.IsDir() && strings.Contains(file.Name(), "farm") {
			count++
		}
	}

	return count, nil
}

func saveCSV(data [][]string, dest string) error {
	file, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("cannot create csv file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		err := writer.Write(value)
		if err != nil {
			return fmt.Errorf("cannot write to file: %v", err)
		}

	}
	return nil
}

func ScrapeLogs(cfg ScraperCfg) error {
	var CSVData = [][]string{}
	dateIndexMap := make(map[string]int)
	farmIndexMap := make(map[string]int)
	CSVFilename := cfg.DestDir + time.Now().Format(formatTimeStr)

	farmFoldersCount, err := countFarmFolders(cfg.SrcDir)
	if err != nil {
		return fmt.Errorf("error counting farm folders: %v", err)
	}

	// CSV Header
	CSVData = append(CSVData, []string{"Date      "})
	dataToBeAdded := []string{}
	for i := 0; i < farmFoldersCount; i++ {
		dataToBeAdded = append(dataToBeAdded, "-------")
		farmName := fmt.Sprintf("farm-%02v", i+1)
		CSVData[0] = append(CSVData[0], farmName)
		farmIndexMap[farmName] = i + 1
	}

	files, err := filePathWalkDir(cfg.SrcDir)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}

	strDateIndexStart := strings.LastIndex(files[0], "/") + 1
	strDateIndexEnd := strDateIndexStart + len(formatTimeStr)
	oldestDate, err := time.Parse(formatTimeStr, files[0][strDateIndexStart:strDateIndexEnd])
	if err != nil {
		return fmt.Errorf("error parsing time: %v", err)
	}

	x := 1
	for d := oldestDate; !d.After(time.Now()); d = d.AddDate(0, 0, 1) {
		data := []string{d.Format(formatTimeStr)}
		data = append(data, dataToBeAdded...)
		CSVData = append(CSVData, data)

		dateIndexMap[d.Format(formatTimeStr)] = x
		x++
	}

	if cfg.Proofs {
		CSVFilename = CSVFilename + "-found-proofs-summary"
	} else if cfg.TotalPlots {
		CSVFilename = CSVFilename + "-total-plots-summary"
	} else if cfg.TotalEligiblePlots {
		CSVFilename = CSVFilename + "-total-eligible-plots-summary"
	}

	for _, file := range files {
		if !strings.Contains(file, "farm") {
			continue
		}

		// Get farm name
		lastSlash := strings.LastIndex(file, "/")
		farmName := file[lastSlash-7 : lastSlash]
		csvDataFarmIndex := farmIndexMap[farmName]

		if cfg.Proofs {
			err = findProofsFound(file, &CSVData, &dateIndexMap, csvDataFarmIndex)
			if err != nil {
				return fmt.Errorf("error finding proofs: %v", err)
			}
		} else if cfg.TotalPlots {
			err = findTotalPlots(file, &CSVData, &dateIndexMap, csvDataFarmIndex)
			if err != nil {
				return fmt.Errorf("error finding total plots: %v", err)
			}
		} else if cfg.TotalEligiblePlots {
			err = findTotalEligiblePlots(file, &CSVData, &dateIndexMap, csvDataFarmIndex)
			if err != nil {
				return fmt.Errorf("error finding total eligible plots: %v", err)
			}
		}

	}

	if cfg.Print {
		for _, line := range CSVData {
			fmt.Println(line)
		}
	}

	if cfg.Save {
		err := saveCSV(CSVData, CSVFilename+".csv")
		if err != nil {
			return fmt.Errorf("error saving csv: %v", err)
		}
	}

	return nil
}
