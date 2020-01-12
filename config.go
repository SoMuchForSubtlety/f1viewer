package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type config struct {
	LiveRetryTimeout      int       `json:"live_retry_timeout"`
	Lang                  string    `json:"preferred_language"`
	CheckUpdate           bool      `json:"check_updates"`
	SaveLogs              bool      `json:"save_logs"`
	LogLocation           string    `json:"log_location"`
	DownloadLocation      string    `json:"download_location"`
	CustomPlaybackOptions []command `json:"custom_playback_options"`
	HorizontalLayout      bool      `json:"horizontal_layout"`
	Theme                 theme     `json:"theme"`
}

func loadConfig() (cfg config, err error) {
	path, err := getConfigPath()
	if err != nil {
		return
	}

	if _, err = os.Stat(path + "config.json"); os.IsNotExist(err) {
		err = nil
		cfg.LiveRetryTimeout = 60
		cfg.Lang = "en"
		cfg.CheckUpdate = true
		cfg.SaveLogs = true
		cfg.save()
		return
	}

	var data []byte
	data, err = ioutil.ReadFile(path + "config.json")
	if err != nil {
		return
	}
	json.Unmarshal(data, &cfg)
	return
}

func (cfg config) save() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(&cfg, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	err = ioutil.WriteFile(path+"config.json", data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error saving config: %v", err)
	}
	return nil
}
