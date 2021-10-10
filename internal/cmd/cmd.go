package cmd

import (
	"context"
	"encoding/json"
	"errors"
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

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/proxy"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/util"
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
	Proxy      bool           `json:"proxy"`
	registry   string
	registry32 string
	flatpak    bool
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
	Event            string
	Category         string
	Title            string
	Session          string
	Date             time.Time
	Year             string
	Country          string
	Series           string
	EpisodeNumber    int64
	OrdinalNumber    int64
	Circuit          string

	Source interface{}
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
			Command: []string{"mpv", "$url", "--alang=" + lang, "--quiet", "--title=$title"},
			Proxy:   true,
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
			Proxy:   true,
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

	if runtime.GOOS == "linux" {
		if lookFlatpak("VLC") {
			store.Commands = append(store.Commands, Command{
				Title:   "Play with VLC Flatpak",
				Command: []string{"org.videolan.VLC", "$url", "--meta-title=$title"},
				flatpak: true,
			})
		}

		if lookFlatpak("mpv") {
			store.Commands = append(store.Commands, Command{
				Title:   "Play with MPV Flatpak",
				Command: []string{"io.mpv.Mpv", "$url", "--meta-title=$title"},
				flatpak: true,
			})

		}
	}

	if len(store.Commands) == 0 {
		store.logger.Error("No compatible players found, make sure they are in your PATH environmen variable")
	}

	store.Commands = append(store.Commands, customCommands...)

	return &store
}

func (s *Store) GetCommand(multi ChannelMatcher) Command {
	if multi.CommandKey != "" {
		for _, c := range s.Commands {
			if strings.EqualFold(multi.CommandKey, c.Title) {
				return c
			}
		}
	}

	return Command{
		Title:   "matcher for " + multi.MatchTitle,
		Command: multi.Command,
	}
}

func (s *Store) RunCommand(cc CommandContext) error {
	url, err := cc.URL()
	if err != nil {
		return fmt.Errorf("could not get video URL: %w", err)
	}

	var proxyEnabled bool
	ctx, cancel := context.WithCancel(context.Background())
	if cc.CustomOptions.Proxy {
		prxy, err := proxy.NewProxyServer(url, s.logger)
		if err != nil && !errors.Is(err, proxy.ErrNotRequired) {
			cancel()
			return err
		} else if err == nil {
			tmpUrl, err := prxy.Listen(ctx)
			if err != nil {
				s.logger.Errorf("failed to start proxy: %s", err)
			} else {
				s.logger.Info("proxy started")
				url = tmpUrl
				proxyEnabled = true
			}
		} else {
			cancel()
			s.logger.Info("proxy not required")
		}
	}

	// replace variables
	tmpCommand := make([]string, len(cc.CustomOptions.Command))
	copy(tmpCommand, cc.CustomOptions.Command)
	metadataJson, err := json.MarshalIndent(cc.MetaData, "", "\t")
	if err != nil {
		s.logger.Error("failed to convert metadata to JSON:", err)
	}
	for i := range tmpCommand {
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$url", url)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$json", string(metadataJson))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$session", cc.MetaData.Session)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$event", cc.MetaData.Event)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$perspective", cc.MetaData.PerspectiveTitle)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$category", cc.MetaData.Category)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$episodenumber", strconv.FormatInt(cc.MetaData.EpisodeNumber, 10))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$season", cc.MetaData.Year)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$title", cc.MetaData.Title)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$filename", sanitizeFileName(cc.MetaData.Title))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$series", cc.MetaData.Series)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$country", cc.MetaData.Country)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$circuit", cc.MetaData.Circuit)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$ordinal", strconv.FormatInt(cc.MetaData.OrdinalNumber, 10))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$time", cc.MetaData.Date.Format(time.RFC3339))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$date", cc.MetaData.Date.Format("2006-01-02"))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$year", cc.MetaData.Year)
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$month", cc.MetaData.Date.Format("01"))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$day", cc.MetaData.Date.Format("02"))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$hour", cc.MetaData.Date.Format("15"))
		tmpCommand[i] = strings.ReplaceAll(tmpCommand[i], "$minute", cc.MetaData.Date.Format("04"))
	}

	if cc.CustomOptions.flatpak {
		return execFlatpak(cc.CustomOptions.Command[0], url, cancel)
	}

	return s.runCmd(exec.Command(tmpCommand[0], tmpCommand[1:]...), proxyEnabled, cancel)
}

func (s *Store) runCmd(cmd *exec.Cmd, proxy bool, cancel func()) error {
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
		cancel()
		return err
	}
	if !proxy {
		cancel()
		return cmd.Process.Release()
	} else {
		go func() {
			_, err := cmd.Process.Wait()
			if err != nil {
				s.logger.Error("process exited with errod: %s", err)
			}
			cancel()
		}()
		return nil
	}
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

// this method is used to check if there exists a flatpak installation of given program
// only relevant for linux systems

func lookFlatpak(file string) bool {

	// use find executable to check if flatpak is installed

	_, err := exec.LookPath("flatpak")

	if err != nil {
		return false
	}

	// official flatpak image for vlc is called org.videolan.VLC
	// we can access installed flatpak programs by calling 'flatpak list'

	// output of flatpak list is (name, appID, version, branch, installation)
	// use the awk program to filter output and only gain name of programs

	flatpak_programs_cmd := exec.Command("bash", "-c", "flatpak list | awk '{print $1}'")
	flatpak_programs, err := flatpak_programs_cmd.Output()

	if err != nil {
		return false
	}

	// now we have a list with programs seperated by newline
	// transform it into an array to loop and check if the desired program is installed

	list := strings.Split(string(flatpak_programs), "\n")

	for _, program := range list {
		if program == file {
			return true
		}
	}

	return false

}

// todo: add support for cancel func since idk what it is

func execFlatpak(cmd string, url string, _ context.CancelFunc) error {
	command := "flatpak run " + cmd + " " + url
	ex := exec.Command("sh", "-c", command)

	_, err := ex.Output()
	return err
}
