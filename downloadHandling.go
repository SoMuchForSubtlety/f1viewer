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

//chage URIs to full URLS
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

//takes asset ID and downloads corresponding .m3u8
func downloadAsset(assetID string, title string) {
	//trim asset ID
	id := ""
	if assetID[:17] == "/api/assets/asse_" {
		id = assetID[17 : len(assetID)-1]
	} else {
		id = assetID
	}

	//get JSON containing .m3u8 url
	response := getProperURL(id)

	//checks for errors
	if response == `{"form_validation_errors": null, "skylark_error_code": null, "error": "Resource not found."}` {
		fmt.Println("There was an error, please review result.json for details and double check the asset ID")
		os.Exit(1)
	}

	//extract url form json
	type urlStruct struct {
		Objects []struct {
			Tata struct {
				TokenisedURL string `json:"tokenised_url"`
			} `json:"tata"`
		} `json:"objects"`
	}

	var finalURL urlStruct

	json.Unmarshal([]byte(response), &finalURL)
	var urlString = finalURL.Objects[0].Tata.TokenisedURL

	//download and patch .m3u8 file
	//TODO: switch id to title
	if err := downloadM3U8(id+".m3u8", urlString); err != nil {
		panic(err)
	}
}

//downloads m3u8
func downloadM3U8(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// convert body to string array
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	lineArray := strings.Split(buf.String(), "\n")
	//apply fix and save
	fixm3u8(lineArray, url, filepath)
	return err
}

//returns the body of the json request as string
func getProperURL(assetID string) string {

	json := `{"asset_url":"/api/assets/asse_` + assetID + `/"}`

	//make request
	body := strings.NewReader(json)
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
	return buf.String()
}
