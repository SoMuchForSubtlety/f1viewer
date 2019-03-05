package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

//returns URL of m3u8
func getM3U8URL(assetID string) string {
	//get playable URL of m3u8 file
	return getProperURL(assetID)
}

//takes asset ID and downloads corresponding .m3u8
func downloadAsset(assetID string, title string) {
	//get JSON containing .m3u8 url
	response := getProperURL(assetID)

	//download and patch .m3u8 file
	//TODO: switch id to title
	downloadM3U8(title+".m3u8", response)
}

//returns valid m3u8 URL as string
func getProperURL(assetID string) string {

	formattedID := `{"asset_url":"` + assetID + `"}`

	//make request
	body := strings.NewReader(formattedID)
	req, err := http.NewRequest("POST", "https://f1tv.formula1.com/api/viewings/", body)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	//converts response body to string
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	repsAsString := buf.String()

	//extract url form json
	type urlStruct struct {
		Objects []struct {
			Tata struct {
				TokenisedURL string `json:"tokenised_url"`
			} `json:"tata"`
		} `json:"objects"`
	}

	var finalURL urlStruct

	json.Unmarshal([]byte(repsAsString), &finalURL)
	var urlString = finalURL.Objects[0].Tata.TokenisedURL
	return urlString
}

//downloads m3u8
func downloadM3U8(filepath string, url string) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// convert body to string array
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	lineArray := strings.Split(buf.String(), "\n")
	//apply fix and save
	fixm3u8(lineArray, url, filepath)
}

//chage URIs in m3u8 to full URLs
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
	writeLines(newLines, filePath)
}

//write m3u8 to file
func writeLines(lines []string, path string) error {
	//create downloads folder if it doesnt exist
	if _, err := os.Stat(`\downloaded\`); os.IsNotExist(err) {
		os.MkdirAll(`.\downloaded\`, os.ModePerm)
	}
	file, err := os.Create(`.\downloaded\` + path)
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
