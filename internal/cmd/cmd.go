package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/SoMuchForSubtlety/f1viewer/internal/util"
	"github.com/gdamore/tcell/v2"
)

type Store struct {
	Commands     []Command
	MultiCommads []MultiCommand
	logger       util.Logger
	lang         string
	accentColor  tcell.Color
}

type commandAndArgs []string

type Command struct {
	Title      string         `json:"title"`
	Command    commandAndArgs `json:"command"`
	registry   string
	registry32 string
}

type MultiCommand struct {
	Title   string           `json:"title,omitempty"`
	Targets []ChannelMatcher `json:"targets,omitempty"`
}

type ChannelMatcher struct {
	MatchTitle string         `json:"match_title,omitempty"`
	Command    commandAndArgs `json:"command,omitempty"`
	CommandKey string         `json:"command_key,omitempty"`
}

type CommandContext struct {
	CustomOptions Command
	MetaData      MetaData
	URL           func() (string, error)
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

func NewStore(customCommands []Command, multiCommands []MultiCommand, lang string, logger util.Logger, accentColor tcell.Color) *Store {
	store := Store{
		logger:       logger,
		lang:         lang,
		accentColor:  accentColor,
		MultiCommads: multiCommands,
	}

	commands := []Command{
		{
			Title:   "Play with MPV",
			Command: []string{"mpv", "$url", "--alang=" + lang, "--start=0", "--quiet", "--title=$title"},
		},
		{
			Title:      "Play with VLC",
			registry:   "SOFTWARE\\WOW6432Node\\VideoLAN\\VLC",
			registry32: "SOFTWARE\\VideoLAN\\VLC",
			Command:    []string{"vlc", "$url", "--meta-title=$title"},
		},
		{
			Title:   "Play with IINA",
			Command: []string{"iina", "--no-stdin", "$url"},
		},
	}

	for _, c := range commands {
		_, err := exec.LookPath(c.Command[0])
		if err == nil {
			store.Commands = append(store.Commands, c)
		} else if c, found := checkRegistry(c); found {
			store.Commands = append(store.Commands, c)
		}
	}

	if runtime.GOOS == "darwin" {
		store.Commands = append(store.Commands, Command{
			Title:   "Play with QuickTime Player",
			Command: []string{"open", "-a", "quicktime player", "$url"},
		})
	}

	if len(store.Commands) == 0 {
		store.logger.Error("No compatible players found, make sure they are in your PATH environmen variable")
	}

	store.Commands = append(store.Commands, customCommands...)

	return &store
}

func (s *Store) RunCommand(cc CommandContext) error {
	url, err := cc.URL()
	if err != nil {
		return fmt.Errorf("could not get video URL: %w", err)
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
	return s.runCmd(exec.Command(tmpCommand[0], tmpCommand[1:]...))
}

func (s *Store) runCmd(cmd *exec.Cmd) error {
	wdir, err := os.Getwd()
	if err != nil {
		// session.logError("unable to get working directory: ", err)
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

	accentColorString := util.ColortoHexString(s.accentColor)
	fmt.Fprintf(s.logger, "[%s::b][[-]%s[%s]]$[-::-] %s\n", accentColorString, wdir, accentColorString, strings.Join(cmd.Args, " "))

	cmd.Stdout = s.logger
	cmd.Stderr = s.logger

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

func sanitizeFileName(s string) string {
	whitespace := regexp.MustCompile(`\s+`)
	var illegal *regexp.Regexp
	if runtime.GOOS == "windows" {
		illegal = regexp.MustCompile(`[<>:"/\\|?*]`)
	} else {
		illegal = regexp.MustCompile(`/`)
	}
	s = illegal.ReplaceAllString(s, " ")
	s = whitespace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	return s
}
