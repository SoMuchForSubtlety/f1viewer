package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type release struct {
	Name   string `json:"name"`
	Body   string `json:"body"`
	Assets []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getRelease() (release, error) {
	var re release
	jsonString, err := getJSON("https://api.github.com/repos/SoMuchForSubtlety/F1viewer/releases/latest")
	if err != nil {
		return re, err
	}
	err = json.Unmarshal([]byte(jsonString), &re)
	if err != nil {
		return re, err
	}
	return re, nil
}

func updateAvailable() (bool, error) {
	url, err := getHashfileLink()
	if err != nil {
		return false, err
	}
	checksums, err := downloadData(url)
	if err != nil {
		return false, err
	}
	hash, err := calculateHash()
	if err != nil {
		return false, err
	}
	re1 := regexp.MustCompile(`^[a-z0-9]+`)
	for _, line := range checksums {
		match := re1.FindString(line)
		if strings.ToLower(match) == strings.ToLower(hash) {
			return false, nil
		}
	}
	return true, nil
}

func getHashfileLink() (string, error) {
	re, err := getRelease()
	if err != nil {
		return "", err
	}
	for _, asset := range re.Assets {
		if asset.Name == "checksums.txt" {
			return asset.BrowserDownloadURL, nil
		}
	}
	return "", errors.New("no checksums.txt found")
}

func calculateHash() (string, error) {
	hasher := sha256.New()
	f, err := os.Open(os.Args[0])
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	f.Close()
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func getUpdateNode() (*tview.TreeNode, error) {
	hasUpdate, err := updateAvailable()
	if !hasUpdate {
		err = errors.New("no update available")
	}
	re, _ := getRelease()
	updateNode := tview.NewTreeNode("UPDATE AVAILABLE").SetColor(tcell.ColorRed).SetExpanded(false).SetReference(re)
	getUpdateNode := tview.NewTreeNode("download update").SetColor(tcell.ColorRed).SetReference("update")
	stopCheckingNode := tview.NewTreeNode("don't tell me about updates").SetColor(tcell.ColorRed).SetReference("don't check")
	updateNode.AddChild(getUpdateNode)
	updateNode.AddChild(stopCheckingNode)
	return updateNode, err
}

func openbrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		return err
	}
	return nil
}
