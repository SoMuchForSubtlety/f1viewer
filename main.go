package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type config struct {
	LiveRetryTimeout      int       `json:"live_retry_timeout"`
	Lang                  string    `json:"preferred_language"`
	CheckUpdate           bool      `json:"check_updates"`
	CustomPlaybackOptions []command `json:"custom_playback_options"`
}

type command struct {
	Title          string     `json:"title"`
	Concurrent     bool       `json:"concurrent"`
	Commands       [][]string `json:"commands"`
	Watchphrase    string     `json:"watchphrase"`
	CommandToWatch int        `json:"command_to_watch"`
}

type commandContext struct {
	EpID          string
	CustomOptions command
	Title         string
}

type viewerSession struct {
	con config

	vodTypes vodTypesStruct

	abortWritingInfo chan bool

	// cache
	episodeMap      map[string]episodeStruct
	driverMap       map[string]driverStruct
	teamMap         map[string]teamStruct
	episodeMapMutex sync.RWMutex
	teamMapMutex    sync.RWMutex
	driverMapMutex  sync.RWMutex

	// tview
	app       *tview.Application
	infoTable *tview.Table
	debugText *tview.TextView
	tree      *tview.TreeView
}

func setWorkingDirectory() {
	//  Get the absolute path this executable is located in.
	executablePath, err := os.Executable()
	if err != nil {
		log.Printf("Error: Couldn't determine working directory: %v", err)
	}
	//  Set the working directory to the path the executable is located in.
	os.Chdir(filepath.Dir(executablePath))
}

func main() {
	debugEnabled := flag.Bool("d", false, "enable debug mode")
	flag.Parse()

	setWorkingDirectory()

	session := viewerSession{}
	logFile, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	defer logFile.Close()
	// start UI
	session.app = tview.NewApplication()

	// set defaults
	session.con.CheckUpdate = true
	session.con.Lang = "en"
	session.con.LiveRetryTimeout = 60
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Println(err.Error())
	} else {
		err = json.Unmarshal(file, &session.con)
		if err != nil {
			log.Fatalf("malformed configuration file: %v", err)
		}
		log.Printf("Found %v custom commands", len(session.con.CustomPlaybackOptions))
	}
	session.abortWritingInfo = make(chan bool)
	// cache
	session.episodeMap = make(map[string]episodeStruct)
	session.driverMap = make(map[string]driverStruct)
	session.teamMap = make(map[string]teamStruct)
	// build base tree
	root := tview.NewTreeNode("VOD-Types").
		SetColor(tcell.ColorBlue).
		SetSelectable(false)
	session.tree = tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)
	var allSeasons allSeasonStruct
	// check for live session
	go func() {
		for {
			log.Println("checking for live session")
			isLive, liveNode, err := getLiveNode()
			if err != nil {
				log.Printf("error looking for live session: %v", err)
			} else if isLive {
				insertNodeAtTop(root, liveNode)
				if session.app != nil {
					session.app.Draw()
				}
				return
			} else if session.con.LiveRetryTimeout < 0 {
				log.Println("no live session found")
				return
			} else {
				log.Println("no live session found")
			}
			time.Sleep(time.Second * time.Duration(session.con.LiveRetryTimeout))
		}
	}()
	// check if an update is available
	if session.con.CheckUpdate {
		go func() {
			node, err := getUpdateNode()
			if err != nil {
				log.Println(err.Error())
			} else {
				insertNodeAtTop(root, node)
				session.app.Draw()
			}
		}()
	}
	// set vod types nodes
	go func() {
		nodes, err := session.getVodTypeNodes()
		if err != nil {
			log.Println(err.Error())
		} else {
			appendNodes(root, nodes...)
			session.app.Draw()
		}
	}()
	// set full race weekends node
	fullSessions := tview.NewTreeNode("Full Race Weekends").
		SetSelectable(true).
		SetReference(allSeasons).
		SetColor(tcell.ColorYellow)
	root.AddChild(fullSessions)
	session.tree.SetChangedFunc(session.nodeSwitched)
	session.tree.SetSelectedFunc(session.nodeSelected)
	// flex containing everything
	flex := tview.NewFlex()
	// flex containing metadata and debug
	rowFlex := tview.NewFlex()
	rowFlex.SetDirection(tview.FlexRow)
	// metadata window
	session.infoTable = tview.NewTable()
	session.infoTable.SetBorder(true).SetTitle(" Info ")
	// debug window
	session.debugText = tview.NewTextView()
	session.debugText.SetBorder(true).SetTitle("Debug")
	session.debugText.SetChangedFunc(func() {
		session.app.Draw()
	})
	mw = io.MultiWriter(session.debugText, logFile)
	log.SetOutput(mw)

	flex.AddItem(session.tree, 0, 2, true)
	// flag -d enables debug window
	if *debugEnabled {
		flex.AddItem(rowFlex, 0, 2, false)
		rowFlex.AddItem(session.infoTable, 0, 2, false)
		rowFlex.AddItem(session.debugText, 0, 1, false)

	}
	session.app.SetRoot(flex, true).Run()
}

// takes year/race ID and returns full year and race nuber as strings
func getYearAndRace(input string) (string, string, error) {
	var fullYear string
	var raceNumber string
	if len(input) < 4 {
		return fullYear, raceNumber, errors.New("not long enough")
	}
	_, err := strconv.Atoi(input[:4])
	if err != nil {
		return fullYear, raceNumber, errors.New("not a valid RearRaceID")
	}
	// TODO fix before 2020
	if input[:4] == "2018" || input[:4] == "2019" {
		return input[:4], "0", nil
	}
	year := input[:2]
	intYear, _ := strconv.Atoi(year)
	// TODO: change before 2030
	if intYear < 30 {
		fullYear = "20" + year
	} else {
		fullYear = "19" + year
	}
	raceNumber = input[2:4]
	return fullYear, raceNumber, nil
}

func (session *viewerSession) monitorCommand(node *tview.TreeNode, watchphrase string, stdout io.Reader, stderr io.Reader) {
	done := false
	outScanner := bufio.NewScanner(stdout)
	outScanner.Split(scanLinesCustom)
	errScanner := bufio.NewScanner(stderr)
	errScanner.Split(scanLinesCustom)
	go func() {
		for outScanner.Scan() {
			sText := outScanner.Text()
			fmt.Fprintln(session.debugText, sText)
			if strings.Contains(strings.ToLower(sText), strings.ToLower(watchphrase)) {
				done = true
			}
		}
	}()
	go func() {
		for errScanner.Scan() {
			sText := errScanner.Text()
			fmt.Fprintln(session.debugText, sText)
			if strings.Contains(sText, watchphrase) {
				done = true
			}
		}
	}()
	session.blinkNode(node, &done, tcell.ColorWhite)
	session.app.Draw()
}

// stolen from https://golang.org/src/bufio/scan.go?s=11799:11877#L335
func scanLinesCustom(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	} else if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func (session *viewerSession) nodeSwitched(node *tview.TreeNode) {
	reference := node.GetReference()
	if index, ok := reference.(int); ok && index < len(session.vodTypes.Objects) {
		v, t := getTableValuesFromInterface(session.vodTypes.Objects[index])
		go session.fillTableFromSlices(v, t, session.abortWritingInfo)
	} else if x := reflect.ValueOf(reference); x.Kind() == reflect.Struct {
		v, t := getTableValuesFromInterface(reference)
		go session.fillTableFromSlices(v, t, session.abortWritingInfo)
	} else if len(node.GetChildren()) != 0 {
		session.infoTable.Clear()
	}
	session.infoTable.ScrollToBeginning()
	session.app.Draw()
}

func (session *viewerSession) nodeSelected(node *tview.TreeNode) {
	reference := node.GetReference()
	children := node.GetChildren()
	if reference == nil || node.GetText() == "loading..." {
		// Selecting the root node or a loading node does nothing
		return
	} else if len(children) > 0 {
		// Collapse if visible, expand if collapsed.
		node.SetExpanded(!node.IsExpanded())
	} else if ep, ok := reference.(episodeStruct); ok {
		// if regular episode is selected for the first time
		nodes := session.getPlaybackNodes(ep.Title, ep.Items[0])
		appendNodes(node, nodes...)
	} else if ep, ok := reference.(channelUrlsStruct); ok {
		// if single perspective is selected (main feed, driver onboards, etc.) from full race weekends
		// TODO: better name
		nodes := session.getPlaybackNodes(node.GetText(), ep.Self)
		appendNodes(node, nodes...)
	} else if event, ok := reference.(eventStruct); ok {
		// if event (eg. Australian GP 2018) is selected from full race weekends
		done := false
		hasSessions := false
		go func() {
			sessions, err := session.getSessionNodes(event)
			if err != nil {
				log.Println(err.Error())
				hasSessions = true
			} else {
				for _, session := range sessions {
					if session != nil && len(session.GetChildren()) > 0 {
						hasSessions = true
						node.AddChild(session)
					}
				}
			}
			done = true
		}()
		go func() {
			session.blinkNode(node, &done, tcell.ColorWhite)
			if !hasSessions {
				node.SetColor(tcell.ColorRed)
				node.SetText(node.GetText() + " - NO CONTENT AVAILABLE")
				node.SetSelectable(false)
			}
			session.app.Draw()
		}()
	} else if season, ok := reference.(seasonStruct); ok {
		// if full season is selected from full race weekends
		done := false
		go func() {
			events, err := getEventNodes(season)
			if err != nil {
				log.Println(err.Error())
			} else {
				for _, event := range events {
					if event != nil {
						layout := "2006-01-02"
						e := event.GetReference().(eventStruct)
						t, _ := time.Parse(layout, e.StartDate)
						if t.Before(time.Now().AddDate(0, 0, 1)) {
							node.AddChild(event)
						}
					}
				}
			}
			done = true
		}()
		go session.blinkNode(node, &done, tcell.ColorWheat)
	} else if context, ok := reference.(commandContext); ok {
		go func() {
			err := session.runCustomCommand(context, node)
			if err != nil {
				log.Println(err.Error())
			}
		}()
	} else if i, ok := reference.(int); ok {
		// if episodes for category are not loaded yet
		if i < len(session.vodTypes.Objects) {
			done := false
			go func() {
				episodes, err := session.getEpisodeNodes(session.vodTypes.Objects[i].ContentUrls)
				if err != nil {
					log.Println(err.Error())
				} else {
					appendNodes(node, episodes...)
				}
				done = true
			}()
			go session.blinkNode(node, &done, tcell.ColorYellow)
		}
	} else if _, ok := reference.(allSeasonStruct); ok {
		done := false
		go func() {
			seasons, err := getSeasonNodes()
			if err != nil {
				log.Println(err.Error())
			} else {
				appendNodes(node, seasons...)
				node.SetReference(seasons)
			}
			done = true
		}()
		go session.blinkNode(node, &done, tcell.ColorYellow)
	} else if node.GetText() == "Play with MPV" {
		go func() {
			url, err := getPlayableURL(reference.(string))
			if err != nil {
				log.Println(err.Error())
				return
			}
			cmd := exec.Command("mpv", url, "--alang="+session.con.Lang, "--start=0")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Println(err)
				return
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				log.Println(err)
				return
			}
			err = cmd.Start()
			if err != nil {
				log.Println(err.Error())
				return
			}
			go session.monitorCommand(node, "Video", stdout, stderr)
		}()
	} else if node.GetText() == "Download .m3u8" {
		go func() {
			node.SetColor(tcell.ColorBlue)
			urlAndTitle := reference.([]string)
			url, err := getPlayableURL(urlAndTitle[0])
			if err != nil {
				log.Println(err.Error())
				return
			}
			_, _, err = downloadAsset(url, urlAndTitle[1])
			if err != nil {
				log.Println(err.Error())
			}
		}()
	} else if node.GetText() == "GET URL" || node.GetText() == "URL copied to clipboard" {
		go func() {
			url, err := getPlayableURL(reference.(string))
			if err != nil {
				log.Println(err.Error())
				return
			}
			err = clipboard.WriteAll(url)
			if err != nil {
				log.Println(err.Error())
				return
			}
			node.SetText("URL copied to clipboard")
			node.SetColor(tcell.ColorBlue)
			session.app.Draw()
		}()
	} else if node.GetText() == "download update" {
		err := openbrowser("https://github.com/SoMuchForSubtlety/F1viewer/releases/latest")
		if err != nil {
			log.Println(err.Error())
		}
	} else if node.GetText() == "don't tell me about updates" {
		session.con.CheckUpdate = false
		err := session.con.save()
		if err != nil {
			log.Println(err.Error())
		}
		node.SetColor(tcell.ColorBlue)
		node.SetText("update notifications turned off")
	}
}

func (session *viewerSession) runCustomCommand(cc commandContext, node *tview.TreeNode) error {
	// custom command
	monitor := false
	com := cc.CustomOptions
	if com.Watchphrase != "" && com.CommandToWatch >= 0 && com.CommandToWatch < len(com.Commands) {
		monitor = true
	}
	url, err := getPlayableURL(cc.EpID)
	if err != nil {
		return err
	}
	var filepath string
	var cookie string
	// run every command
	for j := range com.Commands {
		if len(com.Commands[j]) == 0 {
			continue
		}
		tmpCommand := make([]string, len(com.Commands[j]))
		copy(tmpCommand, com.Commands[j])
		// replace $url, $file and $cookie
		for x, s := range tmpCommand {
			tmpCommand[x] = s
			if (strings.Contains(s, "$file") || strings.Contains(s, "$cookie")) && filepath == "" {
				filepath, cookie, err = downloadAsset(url, cc.Title)
				if err != nil {
					return err
				}
			}
			tmpCommand[x] = strings.Replace(tmpCommand[x], "$file", filepath, -1)
			tmpCommand[x] = strings.Replace(tmpCommand[x], "$cookie", cookie, -1)
			tmpCommand[x] = strings.Replace(tmpCommand[x], "$url", url, -1)
		}
		// run command
		log.Println(append([]string{"starting: "}, tmpCommand...))
		cmd := exec.Command(tmpCommand[0], tmpCommand[1:]...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		err = cmd.Start()
		if err != nil {
			return err
		}
		if monitor && com.CommandToWatch == j {
			go session.monitorCommand(node, com.Watchphrase, stdout, stderr)
		}
		// wait for exit code if commands should not be executed concurrently
		if !com.Concurrent {
			err := cmd.Wait()
			if err != nil {
				return err
			}
		}
	}
	if !monitor {
		node.SetColor(tcell.ColorBlue)
		session.app.Draw()
	}
	return nil
}

func (cfg *config) save() error {
	d, err := json.MarshalIndent(&cfg, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	err = ioutil.WriteFile("config.json", d, 0600)
	if err != nil {
		return fmt.Errorf("error saving config: %v", err)
	}
	return nil
}
