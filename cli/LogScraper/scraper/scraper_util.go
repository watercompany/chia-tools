package scraper

import (
	"io/ioutil"
	"strings"
)

// sortName returns a filename sort key with
// non-negative integer suffixes in numeric order.
// For example, amt, amt0, amt2, amt10, amt099, amt100, ...
// func sortName(filename string) string {
// 	ext := filepath.Ext(filename)
// 	name := filename[:len(filename)-len(ext)]
// 	// split numeric suffix
// 	i := len(name) - 1
// 	for ; i >= 0; i-- {
// 		if '0' > name[i] || name[i] > '9' {
// 			break
// 		}
// 	}
// 	i++
// 	// string numeric suffix to uint64 bytes
// 	// empty string is zero, so integers are plus one
// 	b64 := make([]byte, 64/8)
// 	s64 := name[i:]
// 	if len(s64) > 0 {
// 		u64, err := strconv.ParseUint(s64, 10, 64)
// 		if err == nil {
// 			binary.BigEndian.PutUint64(b64, u64+1)
// 		}
// 	}
// 	// prefix + numeric-suffix + ext
// 	return name[:i] + string(b64) + ext
// }

func filePathWalkDir(root string) ([]string, error) {
	// var files []string
	// err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
	// 	if !info.IsDir() {
	// 		files = append(files, path)
	// 	}
	// 	return nil
	// })

	// sort.Slice(
	// 	files,
	// 	func(i, j int) bool {
	// 		return sortName(files[i][strings.LastIndex(files[i], "/"):]) < sortName(files[j][strings.LastIndex(files[j], "/"):])
	// 	},
	// )
	// return files, err

	var listFiles []string
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() && strings.Contains(file.Name(), "farm") {
			farmFiles, err := ioutil.ReadDir(root + "/" + file.Name())
			if err != nil {
				return nil, err
			}

			for _, farmFile := range farmFiles {
				if !farmFile.IsDir() {
					listFiles = append(listFiles, root+"/"+file.Name()+"/"+farmFile.Name())
				}

				if farmFile.IsDir() && strings.Contains(farmFile.Name(), "live") {
					farmLiveFiles, err := ioutil.ReadDir(root + "/" + file.Name() + "/" + farmFile.Name())
					if err != nil {
						return nil, err
					}

					for _, farmLiveFile := range farmLiveFiles {
						if !farmLiveFile.IsDir() {
							listFiles = append(listFiles, root+"/"+file.Name()+"/"+farmFile.Name()+"/"+farmLiveFile.Name())
						}
					}
				}

			}

		}
	}

	return listFiles, nil
}

func getNumberValue(line string, start int) string {
	end := strings.Index(line[start:], " ") + start
	return line[start:end]
}

func getIndexUntilSpaceToTheLeft(line string, start int) int {
	for {
		if line[start-1:start] != " " {
			start--
		} else {
			break
		}

		if start == 0 {
			break
		}
	}

	return start
}
