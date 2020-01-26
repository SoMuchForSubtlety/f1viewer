package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

type command struct {
	Title   string   `json:"title"`
	Command []string `json:"command"`
}

type commandContext struct {
	EpID          string
	CustomOptions command
	Titles        titles
}

type titles struct {
	PerspectiveTitle string
	SessionTitle     string
	EventTitle       string
	CategoryTitle    string
	EpisodeTitle     string
	SeasonTitle      string
}

func (session *viewerSession) runCustomCommand(cc commandContext) error {
	url, err := getPlayableURL(cc.EpID)
	if err != nil {
		return err
	}
	// replace variables
	var filepath string
	tmpCommand := make([]string, len(cc.CustomOptions.Command))
	copy(tmpCommand, cc.CustomOptions.Command)
	for i := range tmpCommand {
		if strings.Contains(tmpCommand[i], "$file") && filepath == "" {
			filepath, _, err = session.cfg.downloadAsset(url, cc.Titles.String())
			if err != nil {
				return err
			}
		}
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$file", filepath)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$url", url)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$session", cc.Titles.SessionTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$event", cc.Titles.EventTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$perspective", cc.Titles.PerspectiveTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$category", cc.Titles.CategoryTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$episode", cc.Titles.EpisodeTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$season", cc.Titles.SeasonTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$title", cc.Titles.String())
	}
	return session.runCmd(exec.Command(tmpCommand[0], tmpCommand[1:]...))
}

func (session *viewerSession) runCmd(cmd *exec.Cmd) error {
	wdir, err := os.Getwd()
	if err != nil {
		session.logError("unable to get working directory: ", err)
		wdir = "?"
	}
	user, err := user.Current()
	if err == nil {
		if wdir == user.HomeDir {
			wdir = "~"
		} else {
			wdir = filepath.Base(wdir)
		}
	}
	accentColorString := colortoHexString(activeTheme.TerminalAccentColor)
	fmt.Fprintln(session.textWindow, fmt.Sprintf("\n[%s::b][[-]%s[%s]]$[-::-] %s", accentColorString, wdir, accentColorString, strings.Join(cmd.Args, " ")))

	cmd.Stdout = session.textWindow
	cmd.Stderr = session.textWindow

	err = cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Process.Release()
}

func (t titles) String() string {
	var s []string
	if t.SeasonTitle != "" {
		s = append(s, t.SeasonTitle)
	}
	if t.EventTitle != "" {
		s = append(s, t.EventTitle)
	}
	if t.SessionTitle != "" {
		s = append(s, t.SessionTitle)
	}
	if t.PerspectiveTitle != "" {
		s = append(s, t.PerspectiveTitle)
	}
	if t.EpisodeTitle != "" {
		s = append(s, t.EpisodeTitle)
	}

	return sanitizeFileName(strings.Join(s, " - "))
}
