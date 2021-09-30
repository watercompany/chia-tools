package main

import (
	"chia-tools/cli/Watcher/telegrambot"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	srcPath  string
	botToken string
	chatID   string
	farmName string
)

const (
	timeFormatFromLogs = "2006-01-02T15:04:05.000"
	logName            = "debug.log"
)

func init() {
	flag.StringVar(&srcPath, "src", "/home/cf/.chia/mainnet/log", "Directory that contains the chia debug logs")
	flag.StringVar(&botToken, "bot-token", "", "Telegram bot token to be used for sending message to telegram")
	flag.StringVar(&chatID, "chat-id", "", "Telegram chat id of where the message to be sent")
	flag.StringVar(&farmName, "farm-name", "farm-0", "Farm name of where the tool is running")

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

	srcPath = srcPath + logName

	var proofsFoundHistory []string
	var lastProofCheckTimeHistory int64
	for {
		foundProofsArr, lastProofCheckTime, err := processScraping(srcPath)
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

		if lastProofCheckTimeHistory != lastProofCheckTime.Unix() {
			// if more than 15 minutes, send outage message
			if time.Now().Unix()-lastProofCheckTime.Unix() > (60 * 15) {
				err := telegrambot.SendMessage(botToken, chatID, fmt.Sprintf("WARNING: %s: It has been more than 15 minutes since last proof check (%v)", farmName, lastProofCheckTime))
				if err != nil {
					fmt.Printf("error sending message to telegram: %v", err)
					os.Exit(1)
				}
			}

			lastProofCheckTimeHistory = lastProofCheckTime.Unix()
		}

		time.Sleep(60 * time.Second)
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

		s := line[0:23]
		lineDate, err := time.Parse(timeFormatFromLogs, s)
		if err != nil {
			continue
		}
		if strings.Contains(line, "Found 0") {
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
