package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/cmd"
)

type Config struct {
	LiveRetryTimeout      int                `json:"live_retry_timeout"`
	Lang                  string             `json:"preferred_language"`
	CheckUpdate           bool               `json:"check_updates"`
	SaveLogs              bool               `json:"save_logs"`
	LogLocation           string             `json:"log_location"`
	CustomPlaybackOptions []cmd.Command      `json:"custom_playback_options"`
	MultiCommand          []cmd.MultiCommand `json:"multi_commands"`
	HorizontalLayout      bool               `json:"horizontal_layout"`
	Theme                 Theme              `json:"theme"`
	TreeRatio             int                `json:"tree_ratio"`
	OutputRatio           int                `json:"output_ratio"`
	TerminalWrap          bool               `json:"terminal_wrap"`
	DisableTeamColors     bool               `json:"disable_team_colors"`
}

type Theme struct {
	BackgroundColor     string `json:"background_color"`
	BorderColor         string `json:"border_color"`
	CategoryNodeColor   string `json:"category_node_color"`
	FolderNodeColor     string `json:"folder_node_color"`
	ItemNodeColor       string `json:"item_node_color"`
	ActionNodeColor     string `json:"action_node_color"`
	LoadingColor        string `json:"loading_color"`
	LiveColor           string `json:"live_color"`
	UpdateColor         string `json:"update_color"`
	NoContentColor      string `json:"no_content_color"`
	InfoColor           string `json:"info_color"`
	ErrorColor          string `json:"error_color"`
	TerminalAccentColor string `json:"terminal_accent_color"`
	TerminalTextColor   string `json:"terminal_text_color"`
	MultiCommandColor   string `json:"multi_command_color"`
}

// Old configs (e.g. the old default config) may be using ISO 639-1 2-letter
// codes. We remap those codes to ISO 639-2 3-letter codes to prevent those
// configs from breaking.
// https://www.iso.org/iso-639-language-codes.html
var languageCodeRemapping = map[string]string{
	"de": "deu",
	"fr": "fra",
	"es": "spa",
	"nl": "nld",
	"pt": "por",
	"en": "eng",
}

func LoadConfig() (Config, error) {
	var cfg Config
	p, err := GetConfigPath()
	if err != nil {
		return cfg, err
	}

	if _, err = os.Stat(path.Join(p, "config.json")); os.IsNotExist(err) {
		cfg.LiveRetryTimeout = 60
		cfg.Lang = "eng"
		cfg.CheckUpdate = true
		cfg.SaveLogs = true
		cfg.TreeRatio = 1
		cfg.OutputRatio = 1
		cfg.TerminalWrap = true
		err = cfg.Save()
		return cfg, err
	}

	var data []byte
	data, err = ioutil.ReadFile(path.Join(p, "config.json"))
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, err
	}
	if cfg.TreeRatio < 1 {
		cfg.TreeRatio = 1
	}
	if cfg.OutputRatio < 1 {
		cfg.OutputRatio = 1
	}
	// Remap 2-letter code to 3-letter code
	if lang, ok := languageCodeRemapping[cfg.Lang]; ok {
		cfg.Lang = lang
	}

	// TODO: move?
	_, err = configureLogging(cfg)
	return cfg, err
}

func GetConfigPath() (string, error) {
	p, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	p = path.Join(p, "f1viewer")

	_, err = os.Stat(p)
	if os.IsNotExist(err) {
		err = os.MkdirAll(p, os.ModePerm)
	}
	return p, err
}

func GetLogPath(cfg Config) (string, error) {
	var p string
	if cfg.LogLocation == "" {
		// windows, macos
		switch runtime.GOOS {
		case "windows", "darwin":
			configPath, err := GetConfigPath()
			if err != nil {
				return "", err
			}
			p = path.Join(configPath, "logs")
		default:
			// linux, etc.
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			p = path.Join(home, "/.local/share/f1viewer/")
		}
	} else {
		p = cfg.LogLocation
	}

	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		err = os.MkdirAll(p, os.ModePerm)
	}
	return p, err
}

func (cfg Config) Save() error {
	p, err := GetConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(&cfg, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	err = ioutil.WriteFile(path.Join(p, "config.json"), data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error saving config: %v", err)
	}
	return nil
}

func configureLogging(cfg Config) (*os.File, error) {
	if !cfg.SaveLogs {
		log.SetOutput(ioutil.Discard)
		return nil, nil
	}
	logPath, err := GetLogPath(cfg)
	if err != nil {
		return nil, fmt.Errorf("Could not get log path: %w", err)
	}
	completePath := path.Join(logPath, time.Now().Format("2006-01-02")+".log")
	logFile, err := os.OpenFile(completePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("Could not open log file: %w", err)
	}
	log.SetOutput(logFile)
	return logFile, nil
}
