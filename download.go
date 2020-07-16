package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// returns valid m3u8 URL as string
func getPlayableURL(assetID, token string) (string, error) {
	type channelContainer struct {
		ChannelURL string `json:"channel_url"`
	}

	type assetContainer struct {
		AssetURL string `json:"asset_url"`
	}

	isChannel := false
	var err error
	var body []byte
	if strings.Contains(assetID, "/api/channels/") {
		isChannel = true
		body, err = json.Marshal(channelContainer{assetID})
	} else {
		body, err = json.Marshal(assetContainer{assetID})
	}
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", "https://f1tv.formula1.com/api/viewings/", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "JWT "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	err = checkResponse(resp)
	if err != nil {
		return "", err
	}

	// extract url form json
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

	var urlString string
	if isChannel {
		var finalURL channelURLstruct
		err = json.NewDecoder(resp.Body).Decode(&finalURL)
		if err != nil {
			return "", err
		}
		urlString = finalURL.TokenisedURL
	} else {
		var finalURL urlStruct
		err = json.NewDecoder(resp.Body).Decode(&finalURL)
		if err != nil {
			return "", err
		}
		if len(finalURL.Objects) == 0 {
			return "", errors.New("no data received")
		}
		urlString = finalURL.Objects[0].Tata.TokenisedURL
	}
	parsed, err := url.Parse(urlString)
	return parsed.String(), err
}
