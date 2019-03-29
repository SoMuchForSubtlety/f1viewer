package main

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	Lang    string `json:"preferred_language"`
	VLCport string `json:"vlc_telnet_port"`
	VLCpass string `json:"vlc_telnet_pass"`
}

var vodTypes vodTypesStruct
var con config
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
		//log.Fatalln("no config file found, try \"cp sample-config.json config.json\"")
		con.Lang = "en"
		con.VLCpass = ""
		con.VLCport = ""
	} else {
		err = json.Unmarshal(file, &con)
		if err != nil {
			log.Fatalf("malformed configuration file: %v\n", err)
		}
	}

	//TODO: add config for preserred audio language
	//cache
	episodeMap = make(map[string]episodeStruct)
	driverMap = make(map[string]driverStruct)
	teamMap = make(map[string]teamStruct)
	//build base tree
	root := tview.NewTreeNode("VOD-Types")
	root.SetColor(tcell.ColorBlue)
	root.SetSelectable(false)
	tree = tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	//add vod types to tree
	vodTypes = getVodTypes()
	for i, vType := range vodTypes.Objects {
		if len(vType.ContentUrls) > 0 {
			node := tview.NewTreeNode(vType.Name).SetSelectable(true)
			node.SetReference(i)
			node.SetColor(tcell.ColorYellow)
			root.AddChild(node)
		}
	}
	//TODO: add content for info table
	fullSessions := tview.NewTreeNode("Full Race Weekends").SetSelectable(true)
	fullSessions.SetReference(len(vodTypes.Objects))
	fullSessions.SetExpanded(false)
	fullSessions.SetColor(tcell.ColorYellow)
	root.AddChild(fullSessions)

	//display info for the episode or VOD type the cursor is on
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		titles := make([]string, 1)
		values := make([][]string, 1)
		reference := node.GetReference()
		if index, ok := reference.(int); ok && index < len(vodTypes.Objects) { //check if selected node is a vod type
			vodTypeStruct := vodTypes.Objects[index]
			fields := reflect.TypeOf(vodTypeStruct)
			values := reflect.ValueOf(vodTypeStruct)
			go fillTable(fields, values)
		} else if ep, ok := reference.(episodeStruct); ok { //check if selected node is an episode
			//get name and value
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
			go fillTableFromSlices(titles, values)
		} else {
			infoTable.Clear()
		}
		infoTable.ScrollToBeginning()
	})

	//what happens when a node is selected
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		children := node.GetChildren()
		if reference == nil || node.GetText() == "loading..." {
			return //Selecting the root node or a loading node does nothing
		} else if len(children) > 0 {
			node.SetExpanded(!node.IsExpanded()) //Collapse if visible, expand if collapsed.
		} else if ep, ok := reference.(episodeStruct); ok {
			//if regular episode is selected for the first time
			addPlaybackNodes(node, getProperURL(ep.Items[0]), ep.Title, ep.Items[0])
		} else if ep, ok := reference.(channelUrlsStruct); ok {
			//if single perspective is selected (main feed, driver onboards, etc.) from full race weekends
			url := ep.Ovps[0].FullStreamURL
			//TODO: see if high speed tests can be streamed (curretly return empty string)
			newURL := getProperURL(ep.Self)
			if len(newURL) > 5 {
				url = newURL
			}
			//TODO: better name
			addPlaybackNodes(node, url, ep.Name, ep.Self)
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
					node.AddChild(event)
				}
				done = true
			}()
			go blinkNode(node, &done, tcell.ColorWheat)
		} else if node.GetText() == "Play with MPV" {
			//if "play" node is selected
			//TODO: handle mpv not installed
			//TODO: move language selection to config file
			debugPrint(reference.(string))
			cmd := exec.Command("mpv", reference.(string), "--alang="+con.Lang, "--start=0")
			stdoutIn, _ := cmd.StdoutPipe()
			cmd.Start()
			scanner := bufio.NewScanner(stdoutIn)
			//parse command output to see if MPV has opened
			go func() {
				done := false
				go func() {
					for scanner.Scan() {
						sText := scanner.Text()
						debugPrint(sText)
						if strings.Contains(sText, "Video") {
							break
						}
					}
					done = true
				}()
				blinkNode(node, &done, tcell.ColorWhite)
				node.SetText("Play with MPV")
				app.Draw()

			}()
		} else if node.GetText() == "Download .m3u8" {
			node.SetColor(tcell.ColorBlue)
			urlAndTitle := reference.([]string)
			downloadAsset(urlAndTitle[0], urlAndTitle[1])
		} else if node.GetText() == "Stream with VLC" {
			cmd := exec.Command("vlc", "--intf", "telnet", "--telnet-port", con.VLCport, "--telnet-password", con.VLCpass, reference.(string), "--sout", "'#duplicate{dst=std{access=http,mux=ts,dst=:8080}'")
			cmd.Start()
			debugPrint("sent VLC command")
			node.SetColor(tcell.ColorBlue)
		} else if node.GetText() == "GET URL" {
			debugPrint(getProperURL(reference.(string)))
		} else if i, ok := reference.(int); ok {
			//if episodes for category are not loaded yet
			if i < len(vodTypes.Objects) {
				go func() {
					addEpisodes(node, i)
				}()
			} else if i == len(vodTypes.Objects) {
				//special case for full weekends
				done := false
				go func() {
					addSeasons(fullSessions)
					done = true
				}()
				go blinkNode(node, &done, tcell.ColorYellow)
			}
		}
	})

	//start UI
	app = tview.NewApplication()
	//flex containing everything
	flex := tview.NewFlex()
	//flex containing metadata and debug
	rowFlex := tview.NewFlex()
	rowFlex.SetDirection(tview.FlexRow)
	//metadate window
	infoTable = tview.NewTable()
	infoTable.SetBorder(true).SetTitle(" Info ")
	//debug window
	debugText = tview.NewTextView()
	debugText.SetBorder(true)
	debugText.SetTitle("Debug")
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
func fillTable(titles reflect.Type, values reflect.Value) {
	t := make([]string, 1)
	v := make([][]string, 1)

	//iterate through titles and values and add them to the slices
	for i := 0; i < titles.NumField(); i++ {
		title := titles.Field(i)
		value := values.Field(i)

		if value.Kind() == reflect.String {
			//if velue is a string
			t = append(t, title.Name)
			v = append(v, []string{value.String()})
		} else if value.Kind() == reflect.Slice {
			//if value is a slice of strings
			lines := make([]string, value.Len())
			for j := 0; j < value.Len(); j++ {
				lines[j] = value.Index(j).String()
			}
			t = append(t, title.Name)
			v = append(v, lines)
		}
	}
	fillTableFromSlices(t, v)
}

//takes title and values slices and draws them as table
func fillTableFromSlices(titles []string, values [][]string) {
	infoTable.Clear()
	rowIndex := 0
	for index, title := range titles {
		//convert supported API IDs to reasonable strings
		lines := convertIDs(values[index])
		//print to info table
		if len(values[index]) > 0 && len(values[index][0]) > 1 {
			//print title
			infoTable.SetCell(rowIndex, 1, tview.NewTableCell(title).SetAlign(tview.AlignRight).SetTextColor(tcell.ColorBlue))
			//print values
			for _, line := range lines {
				infoTable.SetCell(rowIndex, 2, tview.NewTableCell(line))
				rowIndex++
			}
			rowIndex++
		}
		app.Draw()
	}
	infoTable.ScrollToBeginning()
}

//takes year/race ID and returns full year and race nuber as strings
func getYearAndRace(input string) (string, string) {
	year := input[:2]
	intYear, _ := strconv.Atoi(year)
	var fullYear string
	//TODO: change before 2030
	if intYear < 30 {
		fullYear = "20" + year
	} else {
		fullYear = "19" + year
	}
	raceNumber := input[2:4]
	return fullYear, raceNumber
}

//prints to debug window
func debugPrint(s string) {
	fmt.Fprintf(debugText, s+"\n")
	debugText.ScrollToEnd()
}

//parses multiline values for the info table
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
			//TODO: do the same for teams, etc.
			//check if driver metadata is already cached
			driverMapMutex.RLock()
			driver, ok := driverMap[lines[j]]
			driverMapMutex.RUnlock()
			if !ok {
				debugPrint("team loaded from web")
				//load driver metadata if not already cached
				driver = getDriver(lines[j])
				//add metadata to cache
				driverMapMutex.Lock()
				driverMap[lines[j]] = driver
				driverMapMutex.Unlock()
			} else {
				debugPrint("team loaded from cache")
			}
			//change string to driver name + number from metadata
			number := driver.DriverRacingnumber
			//TODO: move number to back of string
			strNumber := ""
			if number < 10 {
				strNumber = " (" + strconv.Itoa(driver.DriverRacingnumber) + ") "
			} else {
				strNumber = "(" + strconv.Itoa(driver.DriverRacingnumber) + ") "

			}
			lines[j] = strNumber + driver.FirstName + " " + driver.LastName

			//add string to slice
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
				debugPrint("driver loaded from web")
				//load team metadata if not already cached
				team = getTeam(lines[j])
				//add metadata to cache
				teamMapMutex.Lock()
				teamMap[lines[j]] = team
				teamMapMutex.Unlock()
			} else {
				debugPrint("driver loaded from cache")
			}
			//add string to slice
			lines[j] = team.Name
			wg.Done()
		}(j)
	}
	wg.Wait()
	sort.Strings(lines)
	return lines
}

//adds all season to "Full Race Weekends" node
func addSeasons(parentNode *tview.TreeNode) {
	debugPrint("loading seasons")
	seasons := getSeasons()
	seasonList := make([]*tview.TreeNode, len(seasons.Seasons)-68)
	//load 2018
	season1Node := tview.NewTreeNode(seasons.Seasons[1].Name)
	season1Node.SetReference(seasons.Seasons[1])
	seasonList[0] = season1Node
	//load 2019 and onwards
	for i := 69; i < len(seasons.Seasons); i++ {
		seasonNode := tview.NewTreeNode(seasons.Seasons[i].Name)
		seasonNode.SetReference(seasons.Seasons[i])
		seasonList[i-68] = seasonNode
	}
	for _, season := range seasonList {
		parentNode.AddChild(season)
	}
	parentNode.SetExpanded(true)
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
	var wg2 sync.WaitGroup
	wg2.Add(len(event.SessionoccurrenceUrls))
	//iterate through sessions
	for n, sessionID := range event.SessionoccurrenceUrls {
		go func(sessionID string, n int) {
			debugPrint("loading session")
			session := getSession(sessionID)
			if session.Status != "upcoming" {
				debugPrint("loading session streams")
				streams := getSessionStreams(session.Slug)
				sessionNode := tview.NewTreeNode(session.Name).SetSelectable(true)
				if session.Status == "live" {
					sessionNode.SetText(session.Name + " - LIVE")
					sessionNode.SetColor(tcell.ColorRed)
				}
				sessionNode.SetReference(streams.Objects[0].ChannelUrls)
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
			if name == "WIF" {
				name = "Main Feed"
			}
			streamNode := tview.NewTreeNode(name).SetSelectable(true)
			streamNode.SetReference(streamPerspective)
			streamNode.SetColor(tcell.ColorGreen)
			channels[i] = streamNode

			wg3.Done()
		}(i)
	}
	wg3.Wait()
	return channels
}

//blinks node until bool is changed
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
func addEpisodes(target *tview.TreeNode, parentType int) {
	//store loaded episodes to be sorted at the end
	var episodes []episodeStruct

	//waitgroup so sorting doesn't get skipped
	var wg sync.WaitGroup
	wg.Add(len(vodTypes.Objects[parentType].ContentUrls))

	//blink category node until loading is complete
	doneLoading := false
	go blinkNode(target, &doneLoading, tcell.ColorYellow)

	//load every episode
	//TODO: tweak number of threads
	guard := make(chan struct{}, 100)
	go func() {
		for i := range vodTypes.Objects[parentType].ContentUrls {
			//multithread loading the apisodes
			//wait for space in guard
			guard <- struct{}{}
			go func(i int) {
				epID := vodTypes.Objects[parentType].ContentUrls[i]
				//check if episode metadata is already cached
				episodeMapMutex.RLock()
				ep, ok := episodeMap[epID]
				episodeMapMutex.RUnlock()
				if !ok {
					//load episode metadata if not already cached
					ep = getEpisode(epID)
					//add metadata to cache
					episodeMapMutex.Lock()
					episodeMap[epID] = ep
					episodeMapMutex.Unlock()
				}
				//temporarily save loaded episodes
				episodes = append(episodes, ep)
				//make room in guard
				<-guard
				defer wg.Done()
			}(i)
		}
	}()
	//wait for loading to complete
	wg.Wait()
	//sort episodes
	sort.Slice(episodes, func(i, j int) bool {
		_, err := strconv.Atoi(episodes[i].DataSourceID[:4])
		_, err2 := strconv.Atoi(episodes[j].DataSourceID[:4])
		//if one of the episodes doesn't start with a date/race code just compare titles
		if err != nil || err2 != nil {
			return episodes[i].Title < episodes[j].Title
		}
		year1, race1 := getYearAndRace(episodes[i].DataSourceID)
		year2, race2 := getYearAndRace(episodes[j].DataSourceID)
		//sort chronologically by year and race number
		return year1 < year2 || ((year1 == year2) && (race1 < race2))
	})
	//add loaded and sorted episodes to tree
	var skippedEpisodes []*tview.TreeNode
	for _, ep := range episodes {
		node := tview.NewTreeNode(ep.Title).SetSelectable(true)
		node.SetReference(ep)
		node.SetColor(tcell.ColorGreen)
		yearRaceID := ep.DataSourceID[:4]
		//check for year/ race code
		if _, err := strconv.Atoi(yearRaceID); err == nil {
			year := ""
			//TODO: better solution for "2018/19[..]" IDs before
			//special case for IDs that start with 2018/19 since they don't  match the pattern
			if yearRaceID != "2018" && yearRaceID != "2019" {
				year, _ = getYearAndRace(ep.DataSourceID)
			} else {
				year = yearRaceID
			}
			fatherFound := false
			var fatherNode *tview.TreeNode
			//check if there is a node for the specified year
			for _, subNode := range target.GetChildren() {
				if subNode.GetReference() == year {
					fatherNode = subNode
					fatherFound = true
				}
			}
			//if there is no node for the year, create one
			if !fatherFound {
				yearNode := tview.NewTreeNode(year).SetSelectable(true)
				yearNode.SetReference(year)
				yearNode.SetExpanded(false)
				target.AddChild(yearNode)
				fatherNode = yearNode
			}
			//add episode to mathcing year
			fatherNode.AddChild(node)
		} else {
			//save episodes with no year/race ID to be added at the end
			skippedEpisodes = append(skippedEpisodes, node)
		}
	}

	//add skipped episodes to tree
	for _, ep := range skippedEpisodes {
		target.AddChild(ep)
	}
	doneLoading = true
	app.Draw()
}
func addPlaybackNodes(node *tview.TreeNode, url string, title string, epID string) {
	playNode := tview.NewTreeNode("Play with MPV")
	playNode.SetReference(url)
	node.AddChild(playNode)

	if checkArgs("-vlc") {
		streamNode := tview.NewTreeNode("Stream with VLC")
		streamNode.SetReference(url)
		node.AddChild(streamNode)
	}

	if checkArgs("-d") {
		streamNode := tview.NewTreeNode("GET URL")
		streamNode.SetReference(epID)
		node.AddChild(streamNode)
	}

	downloadNode := tview.NewTreeNode("Download .m3u8")
	downloadNode.SetReference([]string{url, title})
	node.AddChild(downloadNode)
}

func checkArgs(searchArg string) bool {
	for _, arg := range os.Args {
		if arg == searchArg {
			return true
		}
	}
	return false
}
