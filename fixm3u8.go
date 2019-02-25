package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func fixm3u8(lines []string, url string, filePath string) {
	//lines, _ := readLines(filePath)
	var newLines []string

	//trim url
	var re1 = regexp.MustCompile(`[^\/]*$`)
	url = re1.ReplaceAllString(url, "")

	//fix URLs in m3u8
	for _, line := range lines {
		if len(line) > 0 && line[:5] == "layer" {
			line = url + line
		} else {
			var re = regexp.MustCompile(`[^"]*m3u8"`)
			tempString := re.FindString(line)
			line = re.ReplaceAllString(line, url+tempString)
		}
		newLines = append(newLines, line)
	}
	writeLines(newLines, "fixed_"+filePath)
}

func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
