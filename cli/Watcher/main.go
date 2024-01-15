package main

import (
	"chia-tools/cli/Watcher/telegrambot"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var (
	srcPath  string
	botToken string
	chatID   string
)

const (
	timeFormatFromLogs = "2006-01-02T15:04:05Z07:00"
	logName            = "debug.log"
)

func init() {
	flag.StringVar(&srcPath, "src", "~/chia-farm-logs", "Directory that contains the chia debug logs")
	flag.StringVar(&botToken, "bot-token", "", "Telegram bot token to be used for sending message to telegram")
	flag.StringVar(&chatID, "chat-id", "", "Telegram chat id of where the message to be sent")

}

func main() {
	flag.Parse()

	if !strings.HasSuffix(srcPath, "/") {
		srcPath = srcPath + "/"
	}

	if !isPathExist(srcPath) {
		fmt.Printf("srcPath %s does not exist", srcPath)
		os.Exit(1)
	}

	farmFoldersCount, err := countFarmFolders(srcPath)
	if err != nil {
		fmt.Printf("error counting farm folders: %v", err)
		os.Exit(1)
	}

	var proofsFoundHistory []string
	lastProofCheckTimeHistory := make([]int64, farmFoldersCount+1)
	// add initial values
	for i := 1; i <= farmFoldersCount; i++ {
		lastProofCheckTimeHistory[i] = time.Now().AddDate(0, 0, -1).UTC().Unix()
	}

	sentErrorMsg := make([]bool, farmFoldersCount+1)

	for {
		for i := 1; i <= farmFoldersCount; i++ {
			farmName := fmt.Sprintf("farm-%02v", i)
			foundProofsArr, lastProofCheckTime, err := processScraping(srcPath + farmName + "/debug.log")
			if err != nil {
				fmt.Printf("error running watcher: %v", err)
				os.Exit(1)
			}

			for _, val := range foundProofsArr {
				if !isExistInArray(val, proofsFoundHistory) {
					proofsFoundHistory = append(proofsFoundHistory, val)
					err := telegrambot.SendMessage(botToken, chatID, fmt.Sprintf("%s: %s", farmName, val))
					if err != nil {
						fmt.Printf("error sending message to telegram: %v", err)
						os.Exit(1)
					}
				}
			}

			lastProofCheckTimeSeconds := lastProofCheckTime.UTC().Unix()
			if lastProofCheckTime == (time.Time{}) {
				lastProofCheckTimeSeconds = lastProofCheckTimeHistory[i]
			}

			if lastProofCheckTimeHistory[i] != lastProofCheckTimeSeconds {
				sentErrorMsg[i] = false
			}
			// if more than 15 minutes, send outage message
			timeNowUTC8Unix := time.Now().UTC().Unix()
			if timeNowUTC8Unix-lastProofCheckTimeHistory[i] > (60*15) && !sentErrorMsg[i] && lastProofCheckTimeHistory[i] != 0 {
				timeStr := fmt.Sprintf("%v", time.Unix(lastProofCheckTimeHistory[i], 0).UTC().Add(time.Hour*time.Duration(8)))
				timeStr = timeStr[:len(timeStr)-9] + "0800 UTC"
				msg := fmt.Sprintf("WARNING: %s: It has been more than 15 minutes since last proof check (%s)", farmName, timeStr)

				if timeNowUTC8Unix-lastProofCheckTimeHistory[i] > (60*15) && timeNowUTC8Unix-lastProofCheckTimeSeconds < (60*15) {
					msg = fmt.Sprintf("INFO: %s: harvester has started doing plot checks", farmName)
				}
				err := telegrambot.SendMessage(botToken, chatID, msg)
				if err != nil {
					fmt.Printf("error sending message to telegram: %v", err)
					os.Exit(1)
				}

				sentErrorMsg[i] = true
			}

			lastProofCheckTimeHistory[i] = lastProofCheckTimeSeconds

		}

		time.Sleep(60 * time.Second)
		if len(proofsFoundHistory) > farmFoldersCount*3 {
			proofsFoundHistory = []string{}
		}
	}

}

func isExistInArray(n string, nArr []string) bool {
	for _, val := range nArr {
		if val == n {
			return true
		}
	}
	return false
}

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

func isPathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}

func getProofFound(lines []string) ([]string, error) {
	var lineStrArr []string
	for _, line := range lines {
		if strings.Contains(line, "Found 1") {
			lineStrArr = append(lineStrArr, line)
		}
	}

	return lineStrArr, nil
}

func getLastProofCheckTime(lines []string) (time.Time, error) {
	var lastProofCheck time.Time
	linesLength := len(lines)
	for i := linesLength - 1; i > 0; i-- {
		line := lines[i]
		if len(line) < 23 {
			continue
		}

		s := line[0:19]
		// Manually add UTC+8 because
		// chia logs doesnt put UTC
		s = s + "+08:00"

		lineDate, err := time.Parse(timeFormatFromLogs, s)
		if err != nil {
			continue
		}
		if strings.Contains(line, "Found 0") && !strings.Contains(line, "0 plots were eligible for farming") {
			lastProofCheck = lineDate
			break
		}
	}

	return lastProofCheck, nil
}

func processScraping(filePath string) ([]string, time.Time, error) {
	var proofsFoundArr []string
	var lastProofCheckTime time.Time

	file, err := os.Open(filePath)
	if err != nil {
		return []string{}, time.Time{}, fmt.Errorf("failed to open log file: %s, skipping reading", err)
	}
	defer file.Close()

	stat, err := os.Stat(filePath)
	if err != nil {
		return []string{}, time.Time{}, fmt.Errorf("error reading log file: %s, skipping reading", err)
	}

	if stat.Size() > 0 {
		buf := make([]byte, stat.Size())
		_, err = file.ReadAt(buf, int64(0))
		if err == nil {
			lines := strings.Split(strings.ReplaceAll(string(buf), "\r\n", "\n"), "\n")
			proofsFoundArr, err = getProofFound(lines)
			if err != nil {
				return []string{}, time.Time{}, fmt.Errorf("error getting proofs found array: %v", err)
			}

			lastProofCheckTime, err = getLastProofCheckTime(lines)
			if err != nil {
				return []string{}, time.Time{}, fmt.Errorf("error getting last proof check time: %v", err)
			}

		} else {
			return []string{}, time.Time{}, fmt.Errorf("error when reading bytes from log file: %s", err)
		}
	}

	return proofsFoundArr, lastProofCheckTime, nil
}
