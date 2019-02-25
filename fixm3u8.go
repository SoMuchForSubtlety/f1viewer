package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func fixm3u8(filePath string, url string) {
	lines, _ := readLines(filePath)
	var newLines []string

	//trim url
	var re1 = regexp.MustCompile(`[^\/]*$`)
	url = re1.ReplaceAllString(url, "")

	//fix URLs in m3u8
	for _, line := range lines {
		if line[:5] == "layer" {
			line = url + line
		} else {
			var re = regexp.MustCompile(`[^"]*"$`)
			tempString := re.FindString(line)
			line = re.ReplaceAllString(line, url+tempString)
		}
		newLines = append(newLines, line)
	}
	writeLines(newLines, "fixed_"+filePath)
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
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
