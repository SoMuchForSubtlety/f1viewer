package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/cmd"
)

type Config struct {
	LiveRetryTimeout          int                `toml:"live_retry_timeout,omitempty"`
	Lang                      []string           `toml:"preferred_languages,omitempty"`
	CheckUpdate               bool               `toml:"check_updates,omitempty"`
	SaveLogs                  bool               `toml:"save_logs,omitempty,omitempty"`
	LogLocation               string             `toml:"log_location,omitempty"`
	UseEnvironmentCredentials bool               `toml:"use_environment_credentials,omitempty"`
	CustomPlaybackOptions     []cmd.Command      `json:"custom_playback_options" toml:"custom_playback_options,omitempty"`
	LiveSessionHooks          []cmd.MultiCommand `toml:"live_session_hooks,omitempty"`
	MultiCommand              []cmd.MultiCommand `json:"multi_commands" toml:"multi_commands,omitempty"`
	HorizontalLayout          bool               `toml:"horizontal_layout,omitempty"`
	Theme                     Theme              `toml:"theme,omitempty"`
	TreeRatio                 int                `toml:"tree_ratio,omitempty"`
	OutputRatio               int                `toml:"output_ratio,omitempty"`
	TerminalWrap              bool               `toml:"terminal_wrap,omitempty"`
	DisableTeamColors         bool               `toml:"disable_team_colors,omitempty"`
	EnableMouse               bool               `toml:"enable_mouse,omitempty"`
}

type ConversionConfig struct {
	CustomPlaybackOptions []cmd.Command      `toml:"custom_playback_options,omitempty"`
	MultiCommand          []cmd.MultiCommand `toml:"multi_commands,omitempty"`
}

type Theme struct {
	BackgroundColor     string `toml:"background_color"`
	BorderColor         string `toml:"border_color"`
	CategoryNodeColor   string `toml:"category_node_color"`
	FolderNodeColor     string `toml:"folder_node_color"`
	ItemNodeColor       string `toml:"item_node_color"`
	ActionNodeColor     string `toml:"action_node_color"`
	LoadingColor        string `toml:"loading_color"`
	LiveColor           string `toml:"live_color"`
	UpdateColor         string `toml:"update_color"`
	NoContentColor      string `toml:"no_content_color"`
	InfoColor           string `toml:"info_color"`
	ErrorColor          string `toml:"error_color"`
	TerminalAccentColor string `toml:"terminal_accent_color"`
	TerminalTextColor   string `toml:"terminal_text_color"`
	MultiCommandColor   string `toml:"multi_command_color"`
}

//go:embed default-config.toml
var defaultConfig []byte

const (
	configName = "config.toml"
)

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

func customOptsAsToml(path string) ([]byte, error) {
	oldCfg, err := os.ReadFile(filepath.Join(path, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("could not open old config file: %w", err)
	}
	var tmpCfg Config
	err = json.Unmarshal(oldCfg, &tmpCfg)
	if err != nil {
		return nil, fmt.Errorf("invalid old config: %w", err)
	}
	tmpCfg2 := ConversionConfig{
		CustomPlaybackOptions: tmpCfg.CustomPlaybackOptions,
		MultiCommand:          tmpCfg.MultiCommand,
	}
	var data bytes.Buffer
	err = toml.NewEncoder(&data).Encode(tmpCfg2)
	if err != nil {
		return nil, fmt.Errorf("could not encode old config as toml: %w", err)
	}

	return data.Bytes(), nil
}

func LoadConfig() (Config, error) {
	var cfg Config
	p, err := GetConfigPath()
	if err != nil {
		return cfg, err
	}

	if _, err = os.Stat(path.Join(p, configName)); os.IsNotExist(err) {
		cfgData := defaultConfig
		customOptsToml, err := customOptsAsToml(p)
		if err == nil {
			cfgData = append(cfgData, 0x0A)              // newline
			cfgData = append(cfgData, customOptsToml...) // add existing custom opts
		}
		err = os.WriteFile(path.Join(p, configName), cfgData, fs.ModePerm)
		if err != nil {
			return cfg, err
		}
	}

	data, err := ioutil.ReadFile(path.Join(p, configName))
	if err != nil {
		return cfg, err
	}
	err = toml.Unmarshal(data, &cfg)
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
	for i, lang := range cfg.Lang {
		if val, ok := languageCodeRemapping[lang]; ok {
			cfg.Lang[i] = val
		}
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

func GetLogPath() (string, error) {
	var p string

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

	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		err = os.MkdirAll(p, os.ModePerm)
	}
	return p, err
}

func configureLogging(cfg Config) (*os.File, error) {
	if !cfg.SaveLogs {
		log.SetOutput(ioutil.Discard)
		return nil, nil
	}
	logPath, err := GetLogPath()
	if err != nil {
		return nil, fmt.Errorf("Could not get log path: %w", err)
	}
	completePath := path.Join(logPath, time.Now().Format("2006-01-02")+".log")
	logFile, err := os.OpenFile(completePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return nil, fmt.Errorf("Could not open log file: %w", err)
	}
	log.SetOutput(logFile)
	return logFile, nil
}
