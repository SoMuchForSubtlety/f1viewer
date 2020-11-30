package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
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
	MetaData      MetaData
}

// MetaData contains title metadata
type MetaData struct {
	PerspectiveTitle string
	SessionTitle     string
	EventTitle       string
	CategoryTitle    string
	EpisodeTitle     string
	SeasonTitle      string
	Date             time.Time
	OrdinalNumber    int
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
	var url string
	var err error
	if cc.EpID != "" {
		url, err = getPlayableURL(cc.EpID, session.authtoken)
		if err != nil {
			return err
		}
	} else {
		url, err = getBackupStream()
		if err != nil {
			return err
		}
	}
	// replace variables
	tmpCommand := make([]string, len(cc.CustomOptions.Command))
	copy(tmpCommand, cc.CustomOptions.Command)
	for i := range tmpCommand {
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$url", url)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$session", cc.MetaData.SessionTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$event", cc.MetaData.EventTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$perspective", cc.MetaData.PerspectiveTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$category", cc.MetaData.CategoryTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$episode", cc.MetaData.EpisodeTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$season", cc.MetaData.SeasonTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$title", cc.MetaData.String())
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$ordinal", strconv.Itoa(cc.MetaData.OrdinalNumber))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$time", cc.MetaData.Date.Format(time.RFC3339))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$date", cc.MetaData.Date.Format("2006-01-02"))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$year", cc.MetaData.Date.Format("2006"))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$month", cc.MetaData.Date.Format("01"))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$day", cc.MetaData.Date.Format("02"))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$hour", cc.MetaData.Date.Format("15"))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$minute", cc.MetaData.Date.Format("04"))
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

func (t MetaData) String() string {
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
