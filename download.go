package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

//takes asset ID and downloads corresponding .m3u8
func downloadAsset(url string, title string) (string, error) {
	//sanitize title
	title = strings.Replace(title, ":", "", -1)
	//abort if the URL is not valid
	if len(url) < 10 {
		return "", errors.New("invalid url")
	}
	//download and patch .m3u8 file
	data, err := downloadData(url)
	if err != nil {
		return "", err
	}
	fixedLineArray := fixData(data, url)
	path, err := writeToFile(fixedLineArray, title+".m3u8")
	if err != nil {
		return "", err
	}
	return strings.Replace(path, " ", "\x20", -1), nil
}

//returns valid m3u8 URL as string
func getPlayableURL(assetID string) (string, error) {
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
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
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
			return "", err
		}
		urlString = finalURL.TokenisedURL
	} else {
		var finalURL urlStruct
		err = json.Unmarshal([]byte(repsAsString), &finalURL)
		if err != nil {
			return "", err
		}
		urlString = finalURL.Objects[0].Tata.TokenisedURL
	}
	return strings.Replace(urlString, "&", "\x26", -1), nil
}

//downloads m3u8 data and returns it as slice
func downloadData(url string) ([]string, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// convert body to string array
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	lineArray := strings.Split(buf.String(), "\n")
	return lineArray, nil
}

//chage URIs in m3u8 data to full URLs
func fixData(lines []string, url string) []string {
	var newLines []string
	//trim url
	var re1 = regexp.MustCompile(`[^\/]*$`)
	url = re1.ReplaceAllString(url, "")

	//fix URLs in m3u8
	for _, line := range lines {
		if strings.Contains(line, "https") {
		} else if len(line) > 6 && (line[:5] == "layer" || line[:4] == "clip" || line[:3] == "OTT") {
			line = url + line
		} else {
			var re = regexp.MustCompile(`[^"]*m3u8"`)
			tempString := re.FindString(line)
			line = re.ReplaceAllString(line, url+tempString)
		}
		var re2 = regexp.MustCompile(`https:\/\/f1tv-cdn[^\.]*\.formula1\.com`)
		line = re2.ReplaceAllString(line, "https://f1tv.secure.footprint.net")
		newLines = append(newLines, line)
	}
	return newLines
}

//write slice of lines to file and return the full file path
func writeToFile(lines []string, path string) (string, error) {
	//create downloads folder if it doesnt exist
	if _, err := os.Stat(`/downloads/`); os.IsNotExist(err) {
		os.MkdirAll(`./downloads/`, os.ModePerm)
	}
	path = `./downloads/` + path
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	err = w.Flush()
	if err != nil {
		return "", err
	}
	return path, nil
}
