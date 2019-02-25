package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {

	assetID := os.Args[1:]
	download := false

	if len(assetID) < 1 {
		fmt.Println("please provide an asset ID")
		os.Exit(1)
	} else if len(assetID) >= 2 && assetID[1] == "-d" {
		download = true
	}

	baseJSON := `{"asset_url":"`
	extendedJSON := `{"asset_url":"/api/assets/asse_`

	var finalJSON = ""

	if assetID[0][:17] == "/api/assets/asse_" {
		finalJSON = baseJSON + assetID[0] + `"}`
	} else {
		finalJSON = extendedJSON + assetID[0] + `/"}`
	}

	body := strings.NewReader(finalJSON)
	req, err := http.NewRequest("POST", "https://f1tv.formula1.com/api/viewings/", body)
	if err != nil {
		// handle err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:65.0) Gecko/20100101 Firefox/65.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en, en")
	req.Header.Set("Referer", "https://f1tv.formula1.com/en/episode/1996-spanish-grand-prix")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("Authorization", "JWT eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6bnVsbCwiZXhwIjoxNTUyMjU1MjU5LCJpZCI6NDI0OTgwNH0.VWPxBw4ewxMLImeSsxt060-XjR_z5n_pYsdXIJIRGmM")
	req.Header.Set("X-Countrycode", "zero")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "__zlcmid=mLhKZK6Or2EeNC; fw-uuid=0f6b656c-cd85-4cba-89a5-5935b5; _fbp=fb.1.1550572504615.2068341976; login=%7B%22event%22:%22login%22,%22componentId%22:%22component_login_page%22,%22actionType%22:%22success%22%7D; userOriginCountry=GBR; re-html5-cc-options=%7B%22ccOn%22%3Afalse%2C%22font%22%3A%22Verdana%22%2C%22fontColor%22%3A%22%23FFF%22%2C%22fontSize%22%3A%22large%22%2C%22fontEdge%22%3A%22None%22%2C%22backgroundColor%22%3A%22%23000000%22%2C%22opacity%22%3A1%7D; player-audio-language=fx; user-locale=en; account-info=%7B%22data%22%3A%7B%22subscriptionStatus%22%3A%22active%22%2C%22subscriptionToken%22%3A%22eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJTdWJzY3JpYmVySWQiOiIzNTIzODk1MyIsIlNlc3Npb25JZCI6IjIuM0hxTlwvYmFKYlBmVllqTFVHU3oybmpOWkJycGpQQTdRb3VaMXQrZDhBZDRSWnRpaTFxb3huRjZwTG41QSs0KzVCc0lPMFhOM25oc0tnVlJcL1ZyRVN1VGRuRHNUdUJQU3hZaTJ4QjhrNG9HZmRGUkhMREI5TE9XU3c0TlEzMlZ2R3ZseTZLV21pRU9QbDhKblV1UzYxZ09uVmdcL3ZrZGl2cFI4VXA0Mk55UWVzYmJ4ZmM4bkV0NHp2UCt2d3pKU2ZXaGpvSndQNGdKaFhldmduOThDN2tEZWJ2UXVQb1duUytkUE53aVpoZ090MmV5UGZGYk1TbVdpaUdjd0JTcGpUdTllcVA4czZTY0hzQlJUTkErY3J3dkJWaXc0NTZaS252bjh3NjBEYXhvNEN6c0R1NGhVOXZEWDVaZ2ljQW96Q3pLbHpFb29qUW5VRTBFcTF6eHFcL2lrdHVBOWNNTmlNa0RNTlp1T3JFMnpreTZFZDhpVnFHQlRWMWxhUm9yWjZxU1F1a0pRTHhtQ2RnZkEzdlFnOGtWaExRdjlWVnh0dlZkb1wvbVkySTNsSUNKS3oyU0RQOHNcL1Y3dkF6TEt2dmFRMFJoWnZGaFlOVndiVDBaWkZUYVY2R3poMWFzVzNTXC8wYmltcDJFVjk0YlBkZ1dEY2VuaHdtRVlESll0MUtDaFBTTVBoTGNycm00MWhyN0hXeUtmSzZEUFZQeENCb1pCOXVFdE85Wmt3b0pGR1kzQU5tZ2NDWWY3YlRwRlIzZndmZWM1OXpadnh0QWk2M3ptU1k2RDBuaEw5aFwvd3M2WkZTa3h3ZFBYNWZORTdISjJWWm53NXdDUEF6cW42ZjJmdVNDa3RYMmwyRFdXXC9SUjd3ZWdPU2FUXC8zTWpJaFZ5VktEdWEzdUY0bWFiWTBjemVMTVY4OFNtWGZXQnhuRGNjWUVaSTY2RDNiWDB3eDRQa28yTFJJNndROTNtQTVBWXdtVnlucEhVcjFsMlJFK294SXZkUjZHYlQzNmFSSHlyRUtSMFZRVG1QdXJZaFwvNEpxdzduaVR6RERqRFNGU21KUXp0SUJZMlpUdG05SjU4Vk5oUzMxYVNiV2U0NCtBTW82V01kc3E1aEdDWktnS0ptbVBuZ3NoZGVcL0E9PSIsIkZpcnN0TmFtZSI6Ikpha29iIiwiTGFzdE5hbWUiOiJBaHJlciIsIlN1YnNjcmliZWRQcm9kdWN0IjoiRjEgVFYgUHJvIEFubnVhbCIsIlN1YnNjcmlwdGlvblN0YXR1cyI6ImFjdGl2ZSIsIkVudGl0bGVtZW50IjoiRXhjbHVzaXZlIGxpdmUgdGltaW5nfEFjY2VzcyB0byBmdWxsIHRlYW0gcmFkaW98RHJpdmVyIHRyYWNrZXIgbWFwfFR5cmUgdXNhZ2UgaGlzdG9yeXxBY2Nlc3MgdG8gaGlzdG9yaWMgcmFjZXMgYXJjaGl2ZXxSZXBsYXlzIG9mIGFsbCBGMSwgRjIsIEdQMywgUG9yc2NoZSBTdXBlciBDdXAgcmFjZXN8U2t5IFNwb3J0cyBGMSBDaGFubmVsfExpdmUgY292ZXJhZ2Ugb2YgYWxsIEYxIHJhY2VzfExpdmUgY292ZXJhZ2Ugb2YgRjIsIEdQMywgUG9yc2NoZSBTdXBlcmN1cCByYWNlc3xBY2Nlc3MgdG8gYWxsIEYxIG9uLWJvYXJkIGNhbWVyYXMiLCJFeHRlcm5hbEF1dGhvcml6YXRpb25zQ29udGV4dERhdGEiOiJBVVQiLCJpYXQiOiIxNTUxMDQ1NjU4IiwiZXhwIjoiMTU1MTIxODQ1OCJ9.Hg95ciqwQjGB1HjRXhr_SYuWBXB-Ztwry3tjgskdfvE%22%7D%2C%22profile%22%3A%7B%22SubscriberId%22%3A35238953%2C%22country%22%3A%22GBR%22%2C%22firstName%22%3A%22Jakob%22%7D%7D; user-metadata=undefined; jwtToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6bnVsbCwiZXhwIjoxNTUyMjU1MjU5LCJpZCI6NDI0OTgwNH0.VWPxBw4ewxMLImeSsxt060-XjR_z5n_pYsdXIJIRGmM; userPlan=%5B%22%2Fapi%2Fplans%2Fplan_fce5d57c014c4245a4674f11c7a5ee20%2F%22%5D")
	req.Header.Set("Te", "Trailers")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	//converts response body to string
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	responseBody := buf.String()

	//writes response to file
	file, err := os.Create("result.json")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	fmt.Fprintf(file, responseBody)

	//checks for errors
	if responseBody == `{"form_validation_errors": null, "skylark_error_code": null, "error": "Resource not found."}` {
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

	json.Unmarshal([]byte(responseBody), &finalURL)
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
