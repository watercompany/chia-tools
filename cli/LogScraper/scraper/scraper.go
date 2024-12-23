package scraper

import (
	"chia-tools/cli/Watcher/telegrambot"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	valuePlaceholder   = "-------"
	formatTimeStr      = "2006-01-02"
	timeFormatFromLogs = "2006-01-02T15:04:05.000"
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

func processScraping(cfg *ScraperCfg, filePath string, CSVData *[][]string, processDataMap *map[FarmDateMap][]float64, dateIndexMap *map[string]int, csvDataFarmIndex int) error {
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

			if cfg.Proofs {
				err := parseLogForProofsFound(&cfg.TotalProofsFoundInt, lines, CSVData, dateIndexMap, csvDataFarmIndex)
				if err != nil {
					return fmt.Errorf("error finding proofs: %v", err)
				}
			} else if cfg.TotalPlots {
				err = parseLogForTotalPlots(lines, CSVData, dateIndexMap, csvDataFarmIndex)
				if err != nil {
					return fmt.Errorf("error finding total plots: %v", err)
				}
			} else if cfg.TotalEligiblePlots {
				err := parseLogForTotalEligiblePlots(lines, CSVData, dateIndexMap, csvDataFarmIndex)
				if err != nil {
					return fmt.Errorf("error finding total eligible plots: %v", err)
				}
			} else if cfg.MaxProofTime {
				err := parseLogForMaxProofTime(lines, CSVData, dateIndexMap, csvDataFarmIndex)
				if err != nil {
					return fmt.Errorf("error finding max proof time: %v", err)
				}
			} else if cfg.MedianProofTime {
				err := parseLogForMedianProofTime(lines, CSVData, processDataMap, dateIndexMap, csvDataFarmIndex)
				if err != nil {
					return fmt.Errorf("error finding median proof time: %v", err)
				}
			} else if cfg.MeanProofTime {
				err := parseLogForMeanProofTime(lines, CSVData, processDataMap, dateIndexMap, csvDataFarmIndex)
				if err != nil {
					return fmt.Errorf("error finding mean proof time: %v", err)
				}
			} else if cfg.PercentProofTime != 0 {
				err := parseLogForPercentProofTime(lines, CSVData, processDataMap, dateIndexMap, csvDataFarmIndex)
				if err != nil {
					return fmt.Errorf("error finding percent proof time: %v", err)
				}
			} else if cfg.GapsProofChecks != 0 {
				err := parseLogForGapsProofTime(lines, CSVData, processDataMap, dateIndexMap, csvDataFarmIndex)
				if err != nil {
					return fmt.Errorf("error finding gaps proof checks: %v", err)
				}
			}

		} else {
			return fmt.Errorf("error when reading bytes from log file: %s", err)
		}
	}

	return nil
}

func ScrapeLogs(cfg ScraperCfg) error {
	var CSVData = [][]string{}
	var processDataMap = make(map[FarmDateMap][]float64)
	dateIndexMap := make(map[string]int)
	farmIndexMap := make(map[string]int)
	CSVFilename := cfg.DestDir + time.Now().Format(formatTimeStr)
	cfg.TotalProofsFoundInt = 0

	farmFoldersCount, err := countFarmFolders(cfg.SrcDir)
	if err != nil {
		return fmt.Errorf("error counting farm folders: %v", err)
	}

	// CSV Header
	CSVData = append(CSVData, []string{"Date      "})
	dataToBeAdded := []string{}
	for i := 0; i < farmFoldersCount; i++ {
		dataToBeAdded = append(dataToBeAdded, valuePlaceholder)
		farmName := fmt.Sprintf("farm-%02v", i+1)
		CSVData[0] = append(CSVData[0], farmName)
		farmIndexMap[farmName] = i + 1
	}

	files, err := filePathWalkDir(cfg.SrcDir)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}

	// strDateIndexStart := strings.LastIndex(files[0], "/") + 1
	// strDateIndexEnd := strDateIndexStart + len(formatTimeStr)
	// timeStr := files[0][strDateIndexStart:strDateIndexEnd]
	// if strings.HasPrefix(timeStr, ".") {
	// 	timeStr = files[0][strDateIndexStart+1 : strDateIndexEnd+1]
	// }

	// use the past 30 days
	oldestDate := time.Now().AddDate(0, 0, -30)
	startDate := oldestDate
	untilDate := time.Now()

	if cfg.StartDate != "" {
		oldestDate, err = time.Parse(formatTimeStr, cfg.StartDate)
		startDate = oldestDate
		if err != nil {
			return fmt.Errorf("error parsing time: %v", err)
		}

		// untilDate = oldestDate.AddDate(0, 1, 0)
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
	} else if cfg.MaxProofTime {
		CSVFilename = CSVFilename + "-max-proof-time-summary"
	} else if cfg.MedianProofTime {
		CSVFilename = CSVFilename + "-median-proof-time-summary"
	} else if cfg.MeanProofTime {
		CSVFilename = CSVFilename + "-mean-proof-time-summary"
	} else if cfg.PercentProofTime != 0 {
		CSVFilename = CSVFilename + "-percent-proof-time-summary"
	} else if cfg.GapsProofChecks != 0 {
		CSVFilename = CSVFilename + "-gaps-proof-checks-summary"
	}

	for _, file := range files {
		if !strings.Contains(file, "farm") || strings.Contains(file, "lock") || !(strings.HasSuffix(file, "txt") || strings.HasSuffix(file, "log")) {
			continue
		}
		fileDate := time.Time{}
		lastSlash := strings.LastIndex(file, "/")

		if strings.HasSuffix(file, "log") {
			fileDate, err = time.Parse(formatTimeStr, time.Now().String()[:10])
			if err != nil {
				return fmt.Errorf("error parsing time: %v: %v", err, file)
			}
		} else {
			fileDate, err = time.Parse(formatTimeStr, file[lastSlash+1:lastSlash+11])
			if err != nil {
				return fmt.Errorf("error parsing time: %v: %v", err, file)
			}
		}

		if fileDate.AddDate(0, 0, 1).UnixNano() < startDate.UnixNano() || fileDate.UnixNano() > untilDate.UnixNano() {
			continue
		}

		// Get farm name
		lastSlash = strings.LastIndex(file, "/")
		farmName := file[lastSlash-7 : lastSlash]

		if strings.Contains(file, "live") {
			// farm-00/live -> 12 characters
			// farm-00 -> 7 characters
			farmName = file[lastSlash-12 : lastSlash-5]
		}

		csvDataFarmIndex := farmIndexMap[farmName]

		err = processScraping(&cfg, file, &CSVData, &processDataMap, &dateIndexMap, csvDataFarmIndex)
		if err != nil {
			panic(fmt.Sprintf("error scraping: %v", err))
		}
	}

	// Process median from process data
	if cfg.MedianProofTime {
		err = processMedianProofTime(&CSVData, &processDataMap, &dateIndexMap)
		if err != nil {
			return fmt.Errorf("error processing median proof time: %v", err)
		}
	} else if cfg.MeanProofTime {
		err = processMeanProofTime(&CSVData, &processDataMap, &dateIndexMap)
		if err != nil {
			return fmt.Errorf("error processing mean proof time: %v", err)
		}
	} else if cfg.PercentProofTime != 0 {
		err = processPercentProofTime(float64(cfg.PercentProofTime), &CSVData, &processDataMap, &dateIndexMap)
		if err != nil {
			return fmt.Errorf("error processing percent proof time: %v", err)
		}
	} else if cfg.GapsProofChecks != 0 {
		err = processGapsProofTime(float64(cfg.GapsProofChecks), &CSVData, &processDataMap, &dateIndexMap)
		if err != nil {
			return fmt.Errorf("error processing gaps proof checks: %v", err)
		}
	}

	if cfg.Print {
		for _, line := range CSVData {
			fmt.Println(line)
		}
		fmt.Println(CSVData[0])

		if cfg.TotalProofsFound {
			fmt.Printf("Total Proofs Found = %v\n", cfg.TotalProofsFoundInt)
		}
	}

	if cfg.Save {
		err := saveCSV(CSVData, CSVFilename+".csv")
		if err != nil {
			return fmt.Errorf("error saving csv: %v", err)
		}
	}

	// send to telegram
	// remove farm-2 to farm-5 data for tg for now
	if cfg.SendTelegram {
		// create message
		tgMessage := "-=================-"
		err := telegrambot.SendMessage(cfg.BotToken, cfg.ChatID, tgMessage)
		if err != nil {
			fmt.Printf("error sending message to telegram: %v\n", err)
			os.Exit(1)
		}

		line_message := ""
		for _, line := range CSVData {
			// remove farm-3 to farm-4 data for tg for now
			//line = removeIndex(line, 2)
			//line = removeIndex(line, 3)
			line = removeIndex(line, 3)
			line = removeIndex(line, 3)
			// remove farm-8 data for tg for now
			line = removeIndex(line, 6)

			line_message += fmt.Sprint(line) + "%0D%0A"
		}

		err = telegrambot.SendMessage(cfg.BotToken, cfg.ChatID, line_message)
		if err != nil {
			fmt.Printf("error sending message to telegram: %v\n", err)
			os.Exit(1)
		}

		if cfg.TotalProofsFound {
			formattedLine := fmt.Sprint("Total Proofs Found = ", cfg.TotalProofsFoundInt)
			err := telegrambot.SendMessage(cfg.BotToken, cfg.ChatID, formattedLine)
			if err != nil {
				fmt.Printf("error sending message to telegram: %v\n", err)
				os.Exit(1)
			}
		}
	}
	return nil
}
