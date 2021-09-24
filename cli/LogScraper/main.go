package main

import (
	"chia-tools/cli/LogScraper/scraper"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	srcPath  string
	destPath string

	proofs             *bool
	totalPlots         *bool
	totalEligiblePlots *bool
	maxProofTime       *bool
	medianProofTime    *bool
	meanProofTime      *bool

	save  *bool
	print *bool
)

func init() {
	flag.StringVar(&srcPath, "src", "/mnt/skynas-log/HarvesterLog", "Directory that contains the harvester logs")
	flag.StringVar(&destPath, "dest", "/mnt/skynas-log/HarvesterLog/summary", "destPath of scraped data")

	proofs = flag.Bool("proofs", false, "set if tool will scrape for proof found")
	totalPlots = flag.Bool("total-plots", false, "set if tool will scrape for minimum total plots")
	totalEligiblePlots = flag.Bool("total-eligible-plots", false, "set if tool will scrape for total eligible plots")
	maxProofTime = flag.Bool("max-proof-time", false, "set if tool will scrape for max proof time")
	medianProofTime = flag.Bool("median-proof-time", false, "set if tool will scrape for median proof time")
	meanProofTime = flag.Bool("mean-proof-time", false, "set if tool will scrape for mean proof time")

	save = flag.Bool("save", false, "set if csv will be saved")
	print = flag.Bool("print", false, "set if summary will be printed")
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

	scraperCfg := scraper.ScraperCfg{
		DestDir:            destPath,
		SrcDir:             srcPath,
		Save:               *save,
		Print:              *print,
		Proofs:             *proofs,
		TotalPlots:         *totalPlots,
		TotalEligiblePlots: *totalEligiblePlots,
		MaxProofTime:       *maxProofTime,
		MedianProofTime:    *medianProofTime,
		MeanProofTime:      *meanProofTime,
	}

	err := scraper.ScrapeLogs(scraperCfg)
	if err != nil {
		fmt.Printf("error scraping logs: %v", err)
		os.Exit(1)
	}
}
