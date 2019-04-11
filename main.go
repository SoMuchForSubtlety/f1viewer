package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type config struct {
	Lang                  string    `json:"preferred_language"`
	CustomPlaybackOptions []command `json:"custom_playback_options"`
}

type command struct {
	Title          string     `json:"title"`
	Commands       [][]string `json:"commands"`
	Watchphrase    string     `json:"watchphrase"`
	CommandToWatch int        `json:"command_to_watch"`
}

type nodeContext struct {
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

func main() {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		con.Lang = "en"
	} else {
		err = json.Unmarshal(file, &con)
		if err != nil {
			log.Fatalf("malformed configuration file: %v\n", err)
		}
	}
	abortWritingInfo = make(chan bool)
	//cache
	episodeMap = make(map[string]episodeStruct)
	driverMap = make(map[string]driverStruct)
	teamMap = make(map[string]teamStruct)
	//build base tree
	root := tview.NewTreeNode("VOD-Types").
		SetColor(tcell.ColorBlue).
		SetSelectable(false)
	tree = tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)
	var allSeasons allSeasonStruct
	fullSessions := tview.NewTreeNode("Full Race Weekends").
		SetSelectable(true).
		SetReference(allSeasons).
		SetExpanded(false).
		SetColor(tcell.ColorYellow)
	root.AddChild(fullSessions)
	vodTypes = getVodTypes()
	for i, vType := range vodTypes.Objects {
		if len(vType.ContentUrls) > 0 {
			node := tview.NewTreeNode(vType.Name).
				SetSelectable(true).
				SetReference(i).
				SetColor(tcell.ColorYellow)
			root.AddChild(node)
		}
	}
	//display info for the episode or VOD type the cursor is on
	tree.SetChangedFunc(switchNode)
	//what happens when a node is selected
	tree.SetSelectedFunc(nodeSelected)
	//start UI
	app = tview.NewApplication()
	//flex containing everything
	flex := tview.NewFlex()
	//flex containing metadata and debug
	rowFlex := tview.NewFlex()
	rowFlex.SetDirection(tview.FlexRow)
	//metadata window
	infoTable = tview.NewTable()
	infoTable.SetBorder(true).SetTitle(" Info ")
	//debug window
	debugText = tview.NewTextView()
	debugText.SetBorder(true).SetTitle("Debug")
	debugText.SetChangedFunc(func() {
		app.Draw()
	})

	flex.AddItem(tree, 0, 2, true)
	flex.AddItem(rowFlex, 0, 2, false)
	rowFlex.AddItem(infoTable, 0, 2, false)
	//flag -d enables debug window
	if checkArgs("-d") {
		rowFlex.AddItem(debugText, 0, 1, false)
	}
	app.SetRoot(flex, true).Run()
}

//takes struct reflect Types and values and draws them as a table
func getTableValuesFromInterface(stru interface{}) ([]string, [][]string) {
	titles := reflect.TypeOf(stru)
	values := reflect.ValueOf(stru)
	t := make([]string, 1)
	v := make([][]string, 1)

	//iterate through titles and values and add them to the slices
	for i := 0; i < titles.NumField(); i++ {
		title := titles.Field(i)
		value := values.Field(i)

		if value.Kind() == reflect.Slice {
			lines := make([]string, value.Len())
			for j := 0; j < value.Len(); j++ {
				if value.Index(j).Kind() == reflect.String {
					lines[j] = value.Index(j).String()
				} else if value.Index(j).Kind() == reflect.Struct {
					a, b := getTableValuesFromInterface(value.Index(j).Interface())
					t = append(t, title.Name)
					v = append(v, []string{"================================"})
					t = append(t, a...)
					v = append(v, b...)
					t = append(t, " ")
					v = append(v, []string{"================================"})
				}
			}
			t = append(t, title.Name)
			v = append(v, lines)
		} else if time, ok := value.Interface().(time.Time); ok {
			t = append(t, title.Name)
			v = append(v, []string{time.Format("2006-01-02 15:04:05")})
		} else {
			if !strings.Contains(strings.ToLower(title.Name), "winner") {
				t = append(t, title.Name)
				v = append(v, []string{value.String()})
			}
		}
	}
	return t, v
}

//TODO add channel to abort
//takes title and values slices and draws them as table
func fillTableFromSlices(titles []string, values [][]string, abort chan bool) {
	select {
	case <-abort:
		//aborts previous call
	default:
		//so it doesn't lock
	}
	aborted := false
	go func() {
		//waits for abort signal
		abort <- true
		aborted = true
	}()
	infoTable.Clear()
	rowIndex := 0
	for index, title := range titles {
		//convert supported API IDs to reasonable strings
		lines := convertIDs(values[index])
		if aborted {
			return
		} else if len(values[index]) > 0 && len(values[index][0]) > 1 {
			//print title
			infoTable.SetCell(rowIndex, 1, tview.NewTableCell(title).SetAlign(tview.AlignRight).SetTextColor(tcell.ColorBlue))
			//print values
			for _, line := range lines {
				infoTable.SetCell(rowIndex, 2, tview.NewTableCell(line))
				rowIndex++
			}
			rowIndex++
		}
	}
	infoTable.ScrollToBeginning()
	app.Draw()
}

//takes year/race ID and returns full year and race nuber as strings
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
	if input[:4] == "2018" || input[:4] == "2019" {
		return input[:4], "0", nil
	}
	year := input[:2]
	intYear, _ := strconv.Atoi(year)
	//TODO: change before 2030
	if intYear < 30 {
		fullYear = "20" + year
	} else {
		fullYear = "19" + year
	}
	raceNumber = input[2:4]
	return fullYear, raceNumber, nil
}

//prints to debug window
func debugPrint(s string, x ...string) {
	y := s
	for _, str := range x {
		y += " " + str
	}
	fmt.Fprintf(debugText, y+"\n")
	debugText.ScrollToEnd()
}

//checks for driver or team IDs for the info table
func convertIDs(lines []string) []string {
	if len(lines) < 1 {
		return lines
	}
	if len(lines[0]) > 12 && lines[0][:12] == "/api/driver/" {
		lines = getDriverNames(lines)
	} else if len(lines[0]) > 12 && lines[0][:10] == "/api/team/" {
		lines = getTeamNames(lines)
	}
	return lines
}

//turns slice of driver IDs to their names
func getDriverNames(lines []string) []string {
	var wg sync.WaitGroup
	wg.Add(len(lines))
	//iterate over all lines
	for j := 0; j < len(lines); j++ {
		go func(j int) {
			//check if driver metadata is already cached
			driverMapMutex.RLock()
			driver, ok := driverMap[lines[j]]
			driverMapMutex.RUnlock()
			if !ok {
				//load driver metadata if not already cached
				driver = getDriver(lines[j])
				//add metadata to cache
				driverMapMutex.Lock()
				driverMap[lines[j]] = driver
				driverMapMutex.Unlock()
			}
			//change string to driver name + number from metadata
			name := addNumberToName(driver.DriverRacingnumber, driver.FirstName+" "+driver.LastName)
			lines[j] = name
			wg.Done()
		}(j)
	}
	wg.Wait()
	sort.Strings(lines)
	return lines
}

//turns array of team IDs to their names
func getTeamNames(lines []string) []string {
	var wg sync.WaitGroup
	wg.Add(len(lines))
	//iterate over all lines
	for j := 0; j < len(lines); j++ {
		go func(j int) {
			//check if team metadata is already cached
			teamMapMutex.RLock()
			team, ok := teamMap[lines[j]]
			teamMapMutex.RUnlock()
			if !ok {
				//load team metadata if not already cached
				team = getTeam(lines[j])
				//add metadata to cache
				teamMapMutex.Lock()
				teamMap[lines[j]] = team
				teamMapMutex.Unlock()
			}
			lines[j] = team.Name
			wg.Done()
		}(j)
	}
	wg.Wait()
	sort.Strings(lines)
	return lines
}

//adds all season to "Full Race Weekends" node
func addSeasons(parentNode *tview.TreeNode) allSeasonStruct {
	debugPrint("loading seasons")
	seasons := getSeasons()
	for _, s := range seasons.Seasons {
		seasonNode := tview.NewTreeNode(s.Name)
		seasonNode.SetReference(s)
		parentNode.AddChild(seasonNode)
	}
	parentNode.SetExpanded(true)
	return seasons
}

//returns node for every event (Australian GP, Bahrain GP, etc.)
func getEventNodes(season seasonStruct) []*tview.TreeNode {
	var wg1 sync.WaitGroup
	wg1.Add(len(season.EventoccurrenceUrls))
	events := make([]*tview.TreeNode, len(season.EventoccurrenceUrls))
	//iterate through events
	for m, eventID := range season.EventoccurrenceUrls {
		go func(eventID string, m int) {
			debugPrint("loading event")
			event := getEvent(eventID)
			//if the events actually has saved sassions add it to the tree
			if len(event.SessionoccurrenceUrls) > 0 {
				eventNode := tview.NewTreeNode(event.OfficialName).SetSelectable(true)
				eventNode.SetReference(event)
				events[m] = eventNode
			}
			wg1.Done()
		}(eventID, m)
	}
	wg1.Wait()
	return events
}

//returns node for every session (FP1, FP2, etc.)
func getSessionNodes(event eventStruct) []*tview.TreeNode {
	sessions := make([]*tview.TreeNode, len(event.SessionoccurrenceUrls))
	bonusIDs := make([][]string, len(event.SessionoccurrenceUrls))
	var wg2 sync.WaitGroup
	wg2.Add(len(event.SessionoccurrenceUrls))
	//iterate through sessions
	for n, sessionID := range event.SessionoccurrenceUrls {
		go func(sessionID string, n int) {
			debugPrint("loading session")
			session := getSession(sessionID)
			if session.Status != "upcoming" && session.Status != "expired" {
				debugPrint("loading session streams")
				streams := getSessionStreams(session.Slug)
				sessionNode := tview.NewTreeNode(session.Name).SetSelectable(true)
				bonusIDs[n] = session.ContentUrls
				if session.Status == "live" {
					sessionNode.SetText(session.Name + " - LIVE")
					sessionNode.SetColor(tcell.ColorRed)
				}
				sessionNode.SetReference(streams)
				sessionNode.SetExpanded(false)
				sessions[n] = sessionNode

				channels := getPerspectiveNodes(streams.Objects[0].ChannelUrls)
				for _, stream := range channels {
					sessionNode.AddChild(stream)
				}
			}
			wg2.Done()
		}(sessionID, n)
	}
	wg2.Wait()
	var allIDs []string
	for _, idList := range bonusIDs {
		allIDs = append(allIDs, idList...)
	}
	if len(allIDs) > 0 {
		bonusNode := tview.NewTreeNode("Bonus Content").SetSelectable(true).SetExpanded(false).SetReference("bonus")
		addEpisodes(bonusNode, allIDs)
		return append(sessions, bonusNode)
	}
	return sessions
}

//returns nodes for every perspective (main feed, data feed, drivers, etc.)
func getPerspectiveNodes(perspectives []channelUrlsStruct) []*tview.TreeNode {
	channels := make([]*tview.TreeNode, len(perspectives))
	var wg3 sync.WaitGroup
	wg3.Add(len(perspectives))
	//iterate through all available streams for the session
	for i := range perspectives {
		go func(i int) {
			streamPerspective := perspectives[i]
			name := streamPerspective.Name
			if len(streamPerspective.DriverUrls) > 0 {
				number := streamPerspective.DriverUrls[0].DriverRacingnumber
				name = addNumberToName(number, name)
			}
			switch name {
			case "WIF":
				name = "Main Feed"
			case "pit lane":
				name = "Pit Lane"
			case "driver":
				name = "Driver Tracker"
			case "data":
				name = "Data Channel"
			}
			streamNode := tview.NewTreeNode(name).SetSelectable(true)
			streamNode.SetReference(streamPerspective)
			streamNode.SetColor(tcell.ColorGreen)
			channels[i] = streamNode
			wg3.Done()
		}(i)
	}
	wg3.Wait()
	sort.Slice(channels, func(i, j int) bool {
		return !strings.Contains(channels[i].GetText(), "(")
	})
	return channels
}

//blinks node until bool is changed
//TODO replace done bool with channel?
func blinkNode(node *tview.TreeNode, done *bool, originalColor tcell.Color) {
	originalText := node.GetText()
	node.SetText("loading...")
	for !*done {
		node.SetColor(tcell.ColorBlue)
		app.Draw()
		time.Sleep(200 * time.Millisecond)
		node.SetColor(originalColor)
		app.Draw()
		time.Sleep(200 * time.Millisecond)
	}
	node.SetText(originalText)
	app.Draw()
}

//add episodes to VOD type
func addEpisodes(target *tview.TreeNode, IDs []string) {
	var episodes []episodeStruct
	var wg sync.WaitGroup
	wg.Add(len(IDs))
	//TODO: tweak number of threads
	guard := make(chan struct{}, 100)
	go func() {
		for i := range IDs {
			//wait for space in guard
			guard <- struct{}{}
			go func(i int) {
				epID := IDs[i]
				//check if episode metadata is already cached
				episodeMapMutex.RLock()
				ep, ok := episodeMap[epID]
				episodeMapMutex.RUnlock()
				if !ok {
					//load episode metadata and add to cache
					ep = getEpisode(epID)
					episodeMapMutex.Lock()
					episodeMap[epID] = ep
					episodeMapMutex.Unlock()
				}
				episodes = append(episodes, ep)
				//make room in guard
				<-guard
				defer wg.Done()
			}(i)
		}
	}()
	wg.Wait()
	sort.Slice(episodes, func(i, j int) bool {
		//TODO: move the checks to getYearAndRace
		if len(episodes[i].DataSourceID) >= 4 && len(episodes[j].DataSourceID) >= 4 {
			year1, race1, err := getYearAndRace(episodes[i].DataSourceID)
			year2, race2, err2 := getYearAndRace(episodes[j].DataSourceID)
			if err == nil && err2 == nil {
				//sort chronologically by year and race number
				if year1 != year2 {
					return year1 < year2
				} else if race1 != race2 {
					return race1 < race2
				}
			}
		}
		return episodes[i].Title < episodes[j].Title
	})
	//add loaded and sorted episodes to tree
	var skippedEpisodes []*tview.TreeNode
	for _, ep := range episodes {
		if len(ep.Items) < 1 {
			continue
		}
		node := tview.NewTreeNode(ep.Title).SetSelectable(true).
			SetReference(ep).
			SetColor(tcell.ColorGreen)
		//check for year/ race code
		if year, _, err := getYearAndRace(ep.DataSourceID); err == nil {
			//check if there is a node for the specified year, if not create one
			fatherFound := false
			var fatherNode *tview.TreeNode
			for _, subNode := range target.GetChildren() {
				if subNode.GetReference() == year {
					fatherNode = subNode
					fatherFound = true
				}
			}
			if !fatherFound {
				yearNode := tview.NewTreeNode(year).SetSelectable(true)
				yearNode.SetReference(year)
				yearNode.SetExpanded(false)
				target.AddChild(yearNode)
				fatherNode = yearNode
			}
			fatherNode.AddChild(node)
		} else {
			//save episodes with no year/race ID to be added at the end
			skippedEpisodes = append(skippedEpisodes, node)
		}
	}
	for _, ep := range skippedEpisodes {
		target.AddChild(ep)
	}
	app.Draw()
}

func addPlaybackNodes(node *tview.TreeNode, title string, epID string) {
	//add custom options
	if con.CustomPlaybackOptions != nil {
		for i := range con.CustomPlaybackOptions {
			com := con.CustomPlaybackOptions[i]
			if len(com.Commands) > 0 {
				var context nodeContext
				context.EpID = epID
				context.CustomOptions = com
				context.Title = title
				customNode := tview.NewTreeNode(com.Title)
				customNode.SetReference(context)
				node.AddChild(customNode)
			}
		}
	}

	playNode := tview.NewTreeNode("Play with MPV")
	playNode.SetReference(epID)
	node.AddChild(playNode)

	downloadNode := tview.NewTreeNode("Download .m3u8")
	downloadNode.SetReference([]string{epID, title})
	node.AddChild(downloadNode)

	if checkArgs("-d") {
		streamNode := tview.NewTreeNode("GET URL")
		streamNode.SetReference(epID)
		node.AddChild(streamNode)
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

func addNumberToName(number int, name string) string {
	if number >= 10 {
		name = "(" + strconv.Itoa(number) + ") " + name
	} else {
		name = " (" + strconv.Itoa(number) + ") " + name
	}
	return name
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

func switchNode(node *tview.TreeNode) {
	reference := node.GetReference()
	if index, ok := reference.(int); ok && index < len(vodTypes.Objects) {
		v, t := getTableValuesFromInterface(vodTypes.Objects[index])
		go fillTableFromSlices(v, t, abortWritingInfo)
	} else if event, ok := reference.(eventStruct); ok {
		v, t := getTableValuesFromInterface(event)
		go fillTableFromSlices(v, t, abortWritingInfo)
	} else if season, ok := reference.(seasonStruct); ok {
		v, t := getTableValuesFromInterface(season)
		go fillTableFromSlices(v, t, abortWritingInfo)
	} else if session, ok := reference.(sessionStreamsStruct); ok {
		v, t := getTableValuesFromInterface(session)
		go fillTableFromSlices(v, t, abortWritingInfo)
	} else if channel, ok := reference.(channelUrlsStruct); ok {
		v, t := getTableValuesFromInterface(channel)
		go fillTableFromSlices(v, t, abortWritingInfo)
	} else if seasons, ok := reference.(allSeasonStruct); ok {
		v, t := getTableValuesFromInterface(seasons)
		go fillTableFromSlices(v, t, abortWritingInfo)
	} else if ep, ok := reference.(episodeStruct); ok {
		//get name and value
		titles := make([]string, 1)
		values := make([][]string, 1)
		titles = append(titles, "Title")
		values = append(values, []string{ep.Title})
		titles = append(titles, "Subtitle")
		values = append(values, []string{ep.Subtitle})
		titles = append(titles, "Synopsis")
		values = append(values, []string{ep.Synopsis})
		titles = append(titles, "Drivers")
		values = append(values, ep.DriverUrls)
		titles = append(titles, "Teams")
		values = append(values, ep.TeamUrls)
		go fillTableFromSlices(titles, values, abortWritingInfo)
	} else if len(node.GetChildren()) != 0 {
		infoTable.Clear()
	}
	infoTable.ScrollToBeginning()
}

func nodeSelected(node *tview.TreeNode) {
	reference := node.GetReference()
	children := node.GetChildren()
	if reference == nil || node.GetText() == "loading..." {
		//Selecting the root node or a loading node does nothing
		return
	} else if len(children) > 0 {
		//Collapse if visible, expand if collapsed.
		node.SetExpanded(!node.IsExpanded())
	} else if ep, ok := reference.(episodeStruct); ok {
		//if regular episode is selected for the first time
		addPlaybackNodes(node, ep.Title, ep.Items[0])
	} else if ep, ok := reference.(channelUrlsStruct); ok {
		//if single perspective is selected (main feed, driver onboards, etc.) from full race weekends
		//TODO: better name
		addPlaybackNodes(node, node.GetText(), ep.Self)
	} else if event, ok := reference.(eventStruct); ok {
		//if event (eg. Australian GP 2018) is selected from full race weekends
		done := false
		hasSessions := false
		go func() {
			sessions := getSessionNodes(event)
			for _, session := range sessions {
				if session != nil && len(session.GetChildren()) > 0 {
					hasSessions = true
					node.AddChild(session)
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
		//if full season is selected from full race weekends
		done := false
		go func() {
			events := getEventNodes(season)
			for _, event := range events {
				layout := "2006-01-02"
				e := event.GetReference().(eventStruct)
				t, _ := time.Parse(layout, e.StartDate)
				if t.Before(time.Now()) {
					node.AddChild(event)
				}
			}
			done = true
		}()
		go blinkNode(node, &done, tcell.ColorWheat)
	} else if context, ok := reference.(nodeContext); ok {
		//custom command
		monitor := false
		com := context.CustomOptions
		if com.Watchphrase != "" && com.CommandToWatch >= 0 && com.CommandToWatch < len(com.Commands) {
			monitor = true
		}
		var stdoutIn io.ReadCloser
		url := getPlayableURL(context.EpID)
		var filepath string
		fileLoaded := false
		//run every command
		for j := range com.Commands {
			if len(com.Commands[j]) > 0 {
				tmpCommand := make([]string, len(com.Commands[j]))
				copy(tmpCommand, com.Commands[j])
				//replace $url and $file
				for x, s := range tmpCommand {
					tmpCommand[x] = s
					if strings.Contains(s, "$file") {
						if !fileLoaded {
							filepath = downloadAsset(url, context.Title)
							fileLoaded = true
						}
						tmpCommand[x] = strings.Replace(tmpCommand[x], "$file", filepath, -1)
					}
					tmpCommand[x] = strings.Replace(tmpCommand[x], "$url", url, -1)
				}
				//run command
				debugPrint("starting:", tmpCommand...)
				cmd := exec.Command(tmpCommand[0], tmpCommand[1:]...)
				stdoutIn, _ = cmd.StdoutPipe()
				err := cmd.Start()
				if err != nil {
					debugPrint(err.Error())
				}
				if monitor && com.CommandToWatch == j {
					go monitorCommand(node, com.Watchphrase, stdoutIn)
				}
			}
		}
		if !monitor {
			node.SetColor(tcell.ColorBlue)
		}
	} else if node.GetText() == "Play with MPV" {
		cmd := exec.Command("mpv", getPlayableURL(reference.(string)), "--alang="+con.Lang, "--start=0")
		stdoutIn, _ := cmd.StdoutPipe()
		err := cmd.Start()
		if err != nil {
			debugPrint(err.Error())
		}
		go monitorCommand(node, "Video", stdoutIn)
	} else if node.GetText() == "Download .m3u8" {
		node.SetColor(tcell.ColorBlue)
		urlAndTitle := reference.([]string)
		downloadAsset(getPlayableURL(urlAndTitle[0]), urlAndTitle[1])
	} else if node.GetText() == "GET URL" {
		debugPrint(getPlayableURL(reference.(string)))
	} else if i, ok := reference.(int); ok {
		//if episodes for category are not loaded yet
		if i < len(vodTypes.Objects) {
			go func() {
				doneLoading := false
				go blinkNode(node, &doneLoading, tcell.ColorYellow)
				addEpisodes(node, vodTypes.Objects[i].ContentUrls)
				doneLoading = true
			}()
		}
	} else if _, ok := reference.(allSeasonStruct); ok {
		done := false
		go func() {
			seasons := addSeasons(node)
			node.SetReference(seasons)
			done = true
		}()
		go blinkNode(node, &done, tcell.ColorYellow)
	}
}
