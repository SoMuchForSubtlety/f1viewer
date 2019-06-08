package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

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

var vodTypes vodTypesStruct
var con config
var abortWritingInfo chan bool
var episodeMap map[string]episodeStruct
var driverMap map[string]driverStruct
var teamMap map[string]teamStruct

var episodeMapMutex = sync.RWMutex{}
var driverMapMutex = sync.RWMutex{}
var teamMapMutex = sync.RWMutex{}

var app *tview.Application
var infoTable *tview.Table
var debugText *tview.TextView
var tree *tview.TreeView

func setWorkingDirectory() {
	//  Get the absolute path this executable is located in.
	executablePath, err := os.Executable()
	if err != nil {			
		debugPrint("Error: Couldn't determine working directory:")
		debugPrint(err.Error())
	}
	//  Set the working directory to the path the executable is located in.
	os.Chdir(filepath.Dir(executablePath))
}

func main() {
	// start UI
	app = tview.NewApplication()
	setWorkingDirectory()
	file, err := ioutil.ReadFile("config.json")
	// set defaults
	con.CheckUpdate = true
	con.Lang = "en"
	con.LiveRetryTimeout = 60
	if err != nil {
		debugPrint(err.Error())
	} else {
		err = json.Unmarshal(file, &con)
		if err != nil {
			debugPrint("malformed configuration file:")
			debugPrint(err.Error())
		}
	}
	abortWritingInfo = make(chan bool)
	// cache
	episodeMap = make(map[string]episodeStruct)
	driverMap = make(map[string]driverStruct)
	teamMap = make(map[string]teamStruct)
	// build base tree
	root := tview.NewTreeNode("VOD-Types").
		SetColor(tcell.ColorBlue).
		SetSelectable(false)
	tree = tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)
	var allSeasons allSeasonStruct
	// check for live session
	go func() {
		for {
			debugPrint("checking for live session")
			isLive, liveNode, err := getLiveNode()
			if err != nil {
				debugPrint("error looking for live session:")
				debugPrint(err.Error())
			} else if isLive {
				insertNodeAtTop(root, liveNode)
				if app != nil {
					app.Draw()
				}
				return
			} else if con.LiveRetryTimeout < 0 {
				debugPrint("no live session found")
				return
			} else {
				debugPrint("no live session found")
			}
			time.Sleep(time.Second * time.Duration(con.LiveRetryTimeout))
		}
	}()
	// check if an update is available
	if con.CheckUpdate {
		go func() {
			node, err := getUpdateNode()
			if err != nil {
				debugPrint(err.Error())
			} else {
				insertNodeAtTop(root, node)
				app.Draw()
			}
		}()
	}
	// set vod types nodes
	go func() {
		nodes, err := getVodTypeNodes()
		if err != nil {
			debugPrint(err.Error())
		} else {
			appendNodes(root, nodes...)
			app.Draw()
		}
	}()
	// set full race weekends node
	fullSessions := tview.NewTreeNode("Full Race Weekends").
		SetSelectable(true).
		SetReference(allSeasons).
		SetColor(tcell.ColorYellow)
	root.AddChild(fullSessions)
	tree.SetChangedFunc(nodeSwitched)
	tree.SetSelectedFunc(nodeSelected)
	// flex containing everything
	flex := tview.NewFlex()
	// flex containing metadata and debug
	rowFlex := tview.NewFlex()
	rowFlex.SetDirection(tview.FlexRow)
	// metadata window
	infoTable = tview.NewTable()
	infoTable.SetBorder(true).SetTitle(" Info ")
	// debug window
	debugText = tview.NewTextView()
	debugText.SetBorder(true).SetTitle("Debug")
	debugText.SetChangedFunc(func() {
		app.Draw()
	})

	flex.AddItem(tree, 0, 2, true)
	flex.AddItem(rowFlex, 0, 2, false)
	rowFlex.AddItem(infoTable, 0, 2, false)
	// flag -d enables debug window
	if checkArgs("-d") {
		rowFlex.AddItem(debugText, 0, 1, false)
	}
	app.SetRoot(flex, true).Run()
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

// prints to debug window
func debugPrint(i interface{}) {
	output := fmt.Sprintf("%v", i)
	if debugText != nil {
		fmt.Fprintf(debugText, output+"\n")
		debugText.ScrollToEnd()
	}
}

func checkArgs(searchArg string) bool {
	for _, arg := range os.Args {
		if arg == searchArg {
			return true
		}
	}
	return false
}

func monitorCommand(node *tview.TreeNode, watchphrase string, output io.ReadCloser) {
	scanner := bufio.NewScanner(output)
	done := false
	go func() {
		for scanner.Scan() {
			sText := scanner.Text()
			debugPrint(sText)
			if strings.Contains(sText, watchphrase) {
				break
			}
		}
		done = true
	}()
	blinkNode(node, &done, tcell.ColorWhite)
	app.Draw()
}

func nodeSwitched(node *tview.TreeNode) {
	reference := node.GetReference()
	if index, ok := reference.(int); ok && index < len(vodTypes.Objects) {
		v, t := getTableValuesFromInterface(vodTypes.Objects[index])
		go fillTableFromSlices(v, t, abortWritingInfo)
	} else if x := reflect.ValueOf(reference); x.Kind() == reflect.Struct {
		v, t := getTableValuesFromInterface(reference)
		go fillTableFromSlices(v, t, abortWritingInfo)
	} else if len(node.GetChildren()) != 0 {
		infoTable.Clear()
	}
	infoTable.ScrollToBeginning()
	app.Draw()
}

func nodeSelected(node *tview.TreeNode) {
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
		nodes := getPlaybackNodes(ep.Title, ep.Items[0])
		appendNodes(node, nodes...)
	} else if ep, ok := reference.(channelUrlsStruct); ok {
		// if single perspective is selected (main feed, driver onboards, etc.) from full race weekends
		// TODO: better name
		nodes := getPlaybackNodes(node.GetText(), ep.Self)
		appendNodes(node, nodes...)
	} else if event, ok := reference.(eventStruct); ok {
		// if event (eg. Australian GP 2018) is selected from full race weekends
		done := false
		hasSessions := false
		go func() {
			sessions, err := getSessionNodes(event)
			if err != nil {
				debugPrint(err.Error())
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
			blinkNode(node, &done, tcell.ColorWhite)
			if !hasSessions {
				node.SetColor(tcell.ColorRed)
				node.SetText(node.GetText() + " - NO CONTENT AVAILABLE")
				node.SetSelectable(false)
			}
			app.Draw()
		}()
	} else if season, ok := reference.(seasonStruct); ok {
		// if full season is selected from full race weekends
		done := false
		go func() {
			events, err := getEventNodes(season)
			if err != nil {
				debugPrint(err.Error())
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
		go blinkNode(node, &done, tcell.ColorWheat)
	} else if context, ok := reference.(commandContext); ok {
		go func() {
			err := runCustomCommand(context, node)
			if err != nil {
				debugPrint(err.Error())
			}
		}()
	} else if i, ok := reference.(int); ok {
		// if episodes for category are not loaded yet
		if i < len(vodTypes.Objects) {
			done := false
			go func() {
				episodes, err := getEpisodeNodes(vodTypes.Objects[i].ContentUrls)
				if err != nil {
					debugPrint(err.Error())
				} else {
					appendNodes(node, episodes...)
				}
				done = true
			}()
			go blinkNode(node, &done, tcell.ColorYellow)
		}
	} else if _, ok := reference.(allSeasonStruct); ok {
		done := false
		go func() {
			seasons, err := getSeasonNodes()
			if err != nil {
				debugPrint(err.Error())
			} else {
				appendNodes(node, seasons...)
				node.SetReference(seasons)
			}
			done = true
		}()
		go blinkNode(node, &done, tcell.ColorYellow)
	} else if node.GetText() == "Play with MPV" {
		go func() {
			url, err := getPlayableURL(reference.(string))
			if err != nil {
				debugPrint(err.Error())
				return
			}
			cmd := exec.Command("mpv", url, "--alang="+con.Lang, "--start=0")
			stdoutIn, _ := cmd.StdoutPipe()
			err = cmd.Start()
			if err != nil {
				debugPrint(err.Error())
				return
			}
			go monitorCommand(node, "Video", stdoutIn)
		}()
	} else if node.GetText() == "Download .m3u8" {
		go func() {
			node.SetColor(tcell.ColorBlue)
			urlAndTitle := reference.([]string)
			url, err := getPlayableURL(urlAndTitle[0])
			if err != nil {
				debugPrint(err.Error())
				return
			}
			_, _, err = downloadAsset(url, urlAndTitle[1])
			if err != nil {
				debugPrint(err.Error())
			}
		}()
	} else if node.GetText() == "GET URL" {
		go func() {
			url, err := getPlayableURL(reference.(string))
			if err != nil {
				debugPrint(err.Error())
				return
			}
			debugPrint(url)
		}()
	} else if node.GetText() == "download update" {
		err := openbrowser("https://github.com/SoMuchForSubtlety/F1viewer/releases/latest")
		if err != nil {
			debugPrint(err.Error())
		}
	} else if node.GetText() == "don't tell me about updates" {
		con.CheckUpdate = false
		err := con.save()
		if err != nil {
			debugPrint(err.Error())
		}
		node.SetColor(tcell.ColorBlue)
		node.SetText("update notifications turned off")
	}
}

func runCustomCommand(cc commandContext, node *tview.TreeNode) error {
	// custom command
	monitor := false
	com := cc.CustomOptions
	if com.Watchphrase != "" && com.CommandToWatch >= 0 && com.CommandToWatch < len(com.Commands) {
		monitor = true
	}
	var stdoutIn io.ReadCloser
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
		debugPrint(append([]string{"starting: "}, tmpCommand...))
		cmd := exec.Command(tmpCommand[0], tmpCommand[1:]...)
		stdoutIn, _ = cmd.StdoutPipe()
		err := cmd.Start()
		if err != nil {
			return err
		}
		if monitor && com.CommandToWatch == j {
			go monitorCommand(node, com.Watchphrase, stdoutIn)
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
		app.Draw()
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
