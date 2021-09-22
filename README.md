# chia-tools

## LogPlumber  
        check timestamp order and copy log files from source to destination, sample run: go run main.go setTargets "/home/mike/.chia/mainnet/log/" "/home/mike/logtemp/"

## LogScraper
LogScraper scrapes data from the logs. These data can be
1. Proofs found
2. Total eligible plots
3. Total plots

To get proofs found, use:
```
sudo go run ./cli/LogScraper/main.go 
-src [Directory that contains the harvester logs] 
-dest [Destination directory of saved csv] 
-wins [Set if data scraped will be proofs found] 
-save [Set if csv file will be saved to the dest dir] 
-print [Set if summary will be printed in the cli]
```
Note:
Source Directory must be a directory that contain folders "farm-01", "farm-02", "farm-03", and so on which then contains the actual harvester logs.

Example command:
```
sudo go run ./cli/LogScraper/main.go -src /mnt/skynas-log/HarvesterLog -dest /mnt/skynas-log/HarvesterLog/summary -wins -print -save
```