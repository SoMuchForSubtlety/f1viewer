package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/rivo/tview"
)

const githubURL = "https://api.github.com/repos/SoMuchForSubtlety/F1viewer/releases/latest"

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
	err := doGet(githubURL, &re)
	return re, err
}

func updateAvailable() (bool, error) {
	url, err := getHashfileLink()
	if err != nil {
		return false, err
	}
	checksums, _, err := downloadData(url)
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

func (session *viewerSession) getUpdateNode() (*tview.TreeNode, error) {
	hasUpdate, err := updateAvailable()
	if err != nil {
		return nil, err
	}
	if !hasUpdate {
		return nil, errors.New("no update available")
	}
	updateNode := tview.NewTreeNode("UPDATE AVAILABLE").SetColor(activeTheme.UpdateColor).SetExpanded(false)
	getUpdateNode := tview.NewTreeNode("download update").SetColor(activeTheme.ActionNodeColor)
	getUpdateNode.SetSelectedFunc(func() {
		err := openbrowser("https://github.com/SoMuchForSubtlety/F1viewer/releases/latest")
		if err != nil {
			session.logError(err)
		}
	})
	stopCheckingNode := tview.NewTreeNode("don't tell me about updates").SetColor(activeTheme.ActionNodeColor)
	stopCheckingNode.SetSelectedFunc(func() {
		session.cfg.CheckUpdate = false
		err := session.cfg.save()
		if err != nil {
			session.logError(err)
		}
		session.logInfo("Checking for updates turned off.")
	})
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

func doGet(url string, v interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}
