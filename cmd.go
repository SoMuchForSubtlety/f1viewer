package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

type commandAndArgs []string

type command struct {
	Title   string         `json:"title"`
	Command commandAndArgs `json:"command"`
}

type multiCommand struct {
	Title   string           `json:"title,omitempty"`
	Targets []channelMatcher `json:"targets,omitempty"`
}

type channelMatcher struct {
	MatchTitle string         `json:"match_title,omitempty"`
	Command    commandAndArgs `json:"command,omitempty"`
	CommandKey string         `json:"command_key,omitempty"`
}

type commandContext struct {
	EpID          string
	CustomOptions command
	Titles        Titles
}

// Titles contains title metadata
type Titles struct {
	PerspectiveTitle string
	SessionTitle     string
	EventTitle       string
	CategoryTitle    string
	EpisodeTitle     string
	SeasonTitle      string
}

func (session *viewerSession) loadCommands() {
	commands := []command{
		{
			Title:   "Play with MPV",
			Command: []string{"mpv", "$url", "--alang=" + session.cfg.Lang, "--start=0", "--quiet", "--title=$title"},
		},
		{
			Title:   "Play with VLC",
			Command: []string{"vlc", "$url", "--meta-title=$title"},
		},
		{
			Title:   "Play with IINA",
			Command: []string{"iina", "--no-stdin", "$url"},
		},
	}

	for _, c := range commands {
		_, err := exec.LookPath(c.Command[0])
		if err == nil {
			session.commands = append(session.commands, c)
		}
	}

	if runtime.GOOS == "darwin" {
		session.commands = append(session.commands, command{
			Title:   "Play with QuickTime Player",
			Command: []string{"open", "-a", "quicktime player", "$url"},
		})
	}

	if len(session.commands) == 0 {
		session.logError("No compatible players found, make sure they are in your PATH environmen variable")
	}

}

func (session *viewerSession) runCustomCommand(cc commandContext) error {
	url, err := getPlayableURL(cc.EpID, session.authtoken)
	if err != nil {
		return err
	}
	// replace variables
	tmpCommand := make([]string, len(cc.CustomOptions.Command))
	copy(tmpCommand, cc.CustomOptions.Command)
	for i := range tmpCommand {
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
	fmt.Fprintf(session.textWindow, "[%s::b][[-]%s[%s]]$[-::-] %s\n", accentColorString, wdir, accentColorString, strings.Join(cmd.Args, " "))

	cmd.Stdout = session.textWindow
	cmd.Stderr = session.textWindow

	err = cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Process.Release()
}

func (t Titles) String() string {
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
