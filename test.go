package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {

	args := os.Args[1:]
	download := false

	//check arguments
	if len(args) < 1 {
		fmt.Println("please provide an asset ID")
		os.Exit(1)
	} else if len(args) >= 2 && args[1] == "-d" {
		download = true
	}

	//trim asset ID
	id := ""
	if args[0][:17] == "/api/assets/asse_" {
		id = args[0][17:]
		id = id[:len(id)-1]
	} else {
		id = args[0]
	}

	response := GetProperURL(id)

	//write json response to file
	/*
		file, err := os.Create("result.json")
		if err != nil {
			log.Fatal("Cannot create file", err)
		}
		defer file.Close()
		fmt.Fprintf(file, response)
	*/

	//checks for errors
	if response == `{"form_validation_errors": null, "skylark_error_code": null, "error": "Resource not found."}` {
		fmt.Println("There was an error, please review result.json for details and double check the asset ID")
		os.Exit(1)
	}

	//extract url form json
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

	//download .m3u8 file
	if download {
		if err := DownloadFile(id+"_master.m3u8", urlString); err != nil {
			panic(err)
		}
	}
}

//DownloadFile downloads m3u8 to specified path and applies patch
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

	//TODO: skip writing to file
	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	//create patched version
	fixm3u8(filepath, url)
	//delete original
	out.Close()
	os.Remove(filepath)
	return err
}

//GetProperURL returns the body of the json request
func GetProperURL(assetID string) string {

	extendedJSON := `{"asset_url":"/api/assets/asse_`

	var finalJSON = extendedJSON + assetID + `/"}`

	//make request
	body := strings.NewReader(finalJSON)
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
