package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type config struct {
	AuthToken string `json:"auth_token"`
	Cookie    string `json:"cookie"`
}

func main() {

	//load config with auth toke and cookie

	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println("please provide a config file")
		os.Exit(1)
	}

	var localConfig config

	err = json.Unmarshal(configFile, &localConfig)
	if err != nil {
		log.Fatalf("malformed configuration file: %v\n", err)
	}

	args := os.Args[1:]
	download := false

	if len(args) < 1 {
		fmt.Println("please provide an asset ID")
		os.Exit(1)
	} else if len(args) >= 2 && args[1] == "-d" {
		download = true
	}

	//writes response to file
	file, err := os.Create("result.json")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	response := GetProperURL(args[0], localConfig)
	fmt.Fprintf(file, response)

	//checks for errors
	if response == `{"form_validation_errors": null, "skylark_error_code": null, "error": "Resource not found."}` {
		fmt.Println("There was an error, please review result.json for details and double check the asset ID")
		os.Exit(1)
	}

	//handles json and downloading
	type URLStruct struct {
		Objects []struct {
			Tata struct {
				TokenisedURL string `json:"tokenised_url"`
			} `json:"tata"`
		} `json:"objects"`
	}

	var finalURL URLStruct

	json.Unmarshal([]byte(response), &finalURL)
	var urlString = finalURL.Objects[0].Tata.TokenisedURL
	fmt.Println(urlString)

	if download {
		if err := DownloadFile("master.m3u8", urlString); err != nil {
			panic(err)
		}
	}
}

//DownloadFile downloads url to path
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

//GetProperURL returns the body of the json request
func GetProperURL(assetID string, localConfig config) string {
	baseJSON := `{"asset_url":"`
	extendedJSON := `{"asset_url":"/api/assets/asse_`

	var finalJSON = ""

	if assetID[:17] == "/api/assets/asse_" {
		finalJSON = baseJSON + assetID + `"}`
	} else {
		finalJSON = extendedJSON + assetID + `/"}`
	}

	body := strings.NewReader(finalJSON)
	req, err := http.NewRequest("POST", "https://f1tv.formula1.com/api/viewings/", body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:65.0) Gecko/20100101 Firefox/65.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en, en")
	req.Header.Set("Referer", "https://f1tv.formula1.com/en/episode/1996-spanish-grand-prix")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("Authorization", localConfig.AuthToken)
	req.Header.Set("X-Countrycode", "zero")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", localConfig.Cookie)
	req.Header.Set("Te", "Trailers")

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
