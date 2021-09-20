package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

var lastTime time.Time
var currentTime time.Time
var destPath string
var srcPath string
var logname string


func initCLI() *cli.App {

	app := cli.NewApp()
	app.Name = "log plumber"
	app.Usage = "doing plumber work for logs"
	app.Version = "1.0.0"
	addCLICommands(app)

	app.Action = func(cnx *cli.Context) error {
		cli.ShowAppHelpAndExit(cnx, 0)
		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	cli.VersionFlag = &cli.BoolFlag{
		Name:  "print-version",
		Usage: "print version",
	}

	return app
}

func addCLICommands(app *cli.App) {
	app.Commands = []*cli.Command{
		{
			Name:      "setTargets",
			Usage:     "set source and destination folders for log files",
			UsageText: "set source and destination folders for log files",
			Action: func(cnx *cli.Context) error {
				srcPath = cnx.Args().Get(0)
				destPath = cnx.Args().Get(1)
				if !isFolderExist(srcPath) {
					fmt.Println("source folder does not exist, please try again")
					os.Exit(1)
				}
				if !isFolderExist(destPath) {
					fmt.Println("target folder does not exist, please try again")
					os.Exit(1)
				}
				 
				processLogDir(srcPath)
				return nil
			},
		},
	}
}

func main() {

	app := initCLI()

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
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

func processLogDir(logDir string) {

	err := filepath.Walk(logDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(info.Name(), "log") || strings.HasSuffix(info.Name(), "lock") {
				return nil
			}
			err = processFile(path)
			if err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		log.Fatal(err)
	}

}


func processFile(fname string) error {

	file, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("failed to open log file: %s, skipping reading", err)
	}
	defer file.Close()

	stat, err := os.Stat(fname)
	if err != nil {
		return fmt.Errorf("error reading log file: %s, skipping reading", err)
	}
	start := int64(0)

	if stat.Size() > 0 { //not empty file
		buf := make([]byte, stat.Size())
		_, err = file.ReadAt(buf, start)
		if err == nil {
			lines := strings.Split(strings.ReplaceAll(string(buf), "\r\n", "\n"), "\n")
			parseLines(lines)
		} else {
			return fmt.Errorf("error when reading bytes from log file: %s", err)
		}
	}

	newFile, err := os.Create(destPath + logname)
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()
 
	_, err = io.Copy(newFile, file)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func parseLines(lines []string) error{
	s := ""
	var err error = nil
	formatTimeStr := "2006-01-02T15:04:05.000"
	for i, line := range lines {
		s = line[0:23]
		if i == 0 {
			lastTime,err = time.Parse(formatTimeStr,s)
			if err != nil{
				return err
			}
			
			logname =  strings.Replace(lastTime.Format("2006-01-02 15:04:05"), " ", "-", 1)  
			logname =  strings.Replace(logname, ":", "-", 2)  + "-chia-logs.txt"
		}else{
			currentTime,err = time.Parse(formatTimeStr,s)
			if err == nil {
				if  lastTime.After(currentTime) {
					return fmt.Errorf("error: log timestamp is out of order")
				}
				lastTime = currentTime
			}else{
				return err
			}
		}
		
		
	}
	return err
}

 