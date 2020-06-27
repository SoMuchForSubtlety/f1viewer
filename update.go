package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/rivo/tview"
)

const githubURL = "https://api.github.com/repos/SoMuchForSubtlety/F1viewer/releases/latest"

type release struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Body       string `json:"body"`
}

func getRelease() (release, error) {
	var re release
	err := doGet(githubURL, &re)
	return re, err
}

func (session *viewerSession) CheckUpdate() {
	if !session.cfg.CheckUpdate {
		return
	}
	release, err := getRelease()
	if err != nil {
		session.logError("could not check for release: ", err)
		return
	}
	if release.TagName == version ||
		release.TagName == "v"+version ||
		version == "dev" {
		return
	}

	session.logInfo("New version found!")
	session.logInfo(release.TagName)
	fmt.Fprintln(session.textWindow, "\n[blue::bu]"+release.Name+"[-::-]\n")
	fmt.Fprintln(session.textWindow, release.Body+"\n")

	updateNode := tview.NewTreeNode("UPDATE AVAILABLE").
		SetColor(activeTheme.UpdateColor).
		SetExpanded(false)
	getUpdateNode := tview.NewTreeNode("download update").
		SetColor(activeTheme.ActionNodeColor).
		SetSelectedFunc(func() {
			err := openbrowser("https://github.com/SoMuchForSubtlety/F1viewer/releases/latest")
			if err != nil {
				session.logError(err)
			}
		})
	stopCheckingNode := tview.NewTreeNode("don't tell me about updates")

	stopCheckingNode.SetColor(activeTheme.ActionNodeColor).
		SetSelectedFunc(func() {
			session.cfg.CheckUpdate = false
			err := session.cfg.save()
			if err != nil {
				session.logError(err)
			}
			stopCheckingNode.SetText("update checks turned off")
		})
	updateNode.AddChild(getUpdateNode)
	updateNode.AddChild(stopCheckingNode)

	insertNodeAtTop(session.tree.GetRoot(), updateNode)
	session.app.Draw()
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
