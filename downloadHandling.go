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

//takes asset ID and downloads corresponding .m3u8
func downloadAsset(url string, title string) string {
	//sanitize title
	title = strings.Replace(title, ":", "", -1)
	//get JSON containing .m3u8 url

	//abort if no proper URL was found
	if len(url) < 10 {
		return ""
	}

	//download and patch .m3u8 file
	//TODO: switch id to title
	downloadM3U8(title+".m3u8", url)
	return `./downloaded/` + title + ".m3u8"
}

//returns valid m3u8 URL as string
func getProperURL(assetID string) string {
	formattedID := ""
	isChannel := false
	if strings.Contains(assetID, "/api/channels/") {
		isChannel = true
		formattedID = `{"channel_url":"` + assetID + `"}`
	} else {
		formattedID = `{"asset_url":"` + assetID + `"}`
	}
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

	type channelURLstruct struct {
		TokenisedURL string `json:"tokenised_url"`
	}

	var urlString = ""
	if isChannel {
		var finalURL channelURLstruct
		err = json.Unmarshal([]byte(repsAsString), &finalURL)
		if err != nil {
			fmt.Println(err)
		}
		urlString = finalURL.TokenisedURL

	} else {
		var finalURL urlStruct
		json.Unmarshal([]byte(repsAsString), &finalURL)
		urlString = finalURL.Objects[0].Tata.TokenisedURL
	}
	//debugPrint(urlString)
	return strings.Replace(urlString, "&", "\x26", -1)
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
		debugPrint(line)
		if strings.Contains(line, "https") {
		} else if len(line) > 6 && (line[:5] == "layer" || line[:4] == "clip") {
			line = url + line
		} else {
			var re = regexp.MustCompile(`[^"]*m3u8"`)
			tempString := re.FindString(line)
			line = re.ReplaceAllString(line, url+tempString)
		}
		var re2 = regexp.MustCompile(`https:\/\/f1tv-cdn[^\.]*\.formula1\.com`)
		line = re2.ReplaceAllString(line, "https://f1tv.secure.footprint.net")
		debugPrint(line)
		newLines = append(newLines, line)
	}
	writeLines(newLines, filePath)
}

//write m3u8 to file
func writeLines(lines []string, path string) error {
	//create downloads folder if it doesnt exist
	if _, err := os.Stat(`/downloaded/`); os.IsNotExist(err) {
		os.MkdirAll(`./downloaded/`, os.ModePerm)
	}
	file, err := os.Create(`./downloaded/` + path)
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
