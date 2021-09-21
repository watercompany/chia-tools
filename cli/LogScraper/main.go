package main

import (
	"encoding/binary"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	srcPath  string
	destPath string
	wins     *bool
	save     *bool
	print    *bool
)

const (
	proofFound    = "---1---"
	formatTimeStr = "2006-01-02"
)

func init() {
	flag.StringVar(&srcPath, "src", "/mnt/skynas-log/HarvesterLog", "Directory that contains the harvester logs")
	flag.StringVar(&destPath, "dest", "/mnt/skynas-log/HarvesterLog/summary", "destPath of scraped data")
	wins = flag.Bool("wins", false, "set if tool will scrape proof wins")
	save = flag.Bool("save", false, "set if csv will be saved")
	print = flag.Bool("print", false, "set if summary will be printed")
}

func main() {
	flag.Parse()

	if !strings.HasSuffix(destPath, "/") {
		destPath = destPath + "/"
	}

	if !isFolderExist(srcPath) {
		fmt.Printf("srcPath %s does not exist", srcPath)
		os.Exit(1)
	}

	if *save {
		if !isFolderExist(destPath) {
			fmt.Printf("destPath %s does not exist", destPath)
			os.Exit(1)
		}
	}

	err := scrapeLogs(srcPath)
	if err != nil {
		fmt.Printf("error scraping logs: %v", err)
		os.Exit(1)
	}
}

func isFolderExist(path string) bool {
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

// sortName returns a filename sort key with
// non-negative integer suffixes in numeric order.
// For example, amt, amt0, amt2, amt10, amt099, amt100, ...
func sortName(filename string) string {
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	// split numeric suffix
	i := len(name) - 1
	for ; i >= 0; i-- {
		if '0' > name[i] || name[i] > '9' {
			break
		}
	}
	i++
	// string numeric suffix to uint64 bytes
	// empty string is zero, so integers are plus one
	b64 := make([]byte, 64/8)
	s64 := name[i:]
	if len(s64) > 0 {
		u64, err := strconv.ParseUint(s64, 10, 64)
		if err == nil {
			binary.BigEndian.PutUint64(b64, u64+1)
		}
	}
	// prefix + numeric-suffix + ext
	return name[:i] + string(b64) + ext
}

func filePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	sort.Slice(
		files,
		func(i, j int) bool {
			return sortName(files[i][strings.LastIndex(files[i], "/"):]) < sortName(files[j][strings.LastIndex(files[j], "/"):])
		},
	)
	return files, err
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
	var err error = nil

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
	return err
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

func scrapeLogs(logDir string) error {
	var CSVData = [][]string{}
	dateIndexMap := make(map[string]int)
	farmIndexMap := make(map[string]int)
	CSVFilename := destPath + time.Now().Format(formatTimeStr)

	farmFoldersCount, err := countFarmFolders(logDir)
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

	files, err := filePathWalkDir(logDir)
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

	for _, file := range files {
		if !strings.Contains(file, "farm") {
			continue
		}

		// Get farm name
		lastSlash := strings.LastIndex(file, "/")
		farmName := file[lastSlash-7 : lastSlash]
		csvDataFarmIndex := farmIndexMap[farmName]

		if *wins {
			CSVFilename = CSVFilename + "-found-proofs-summary"
			err = findProofsFound(file, &CSVData, &dateIndexMap, csvDataFarmIndex)
			if err != nil {
				return fmt.Errorf("error finding proofs: %v", err)
			}
		}

	}

	if *print {
		for _, line := range CSVData {
			fmt.Println(line)
		}
	}

	if *save {
		err := saveCSV(CSVData, CSVFilename+".csv")
		if err != nil {
			return fmt.Errorf("error saving csv: %v", err)
		}
	}

	return nil
}
