package main

import (
	"bufio"
	"fmt"
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
	//TODO: add confit for preserred audio language

	//cache for loaded episodes
	episodeMap = make(map[string]episodeStruct)
	driverMap = make(map[string]driverStruct)
	teamMap = make(map[string]teamStruct)

	//build base tree
	rootDir := "VOD-Types"
	root := tview.NewTreeNode(rootDir)
	root.SetColor(tcell.ColorBlue)
	root.SetSelectable(false)
	tree = tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	//add vod types to tree
	vodTypes := getVodTypes()
	maxi := 0
	for i, vType := range vodTypes.Objects {
		node := tview.NewTreeNode(vType.Name).SetSelectable(true)
		node.SetReference(i)
		node.SetColor(tcell.ColorYellow)
		root.AddChild(node)
		maxi = i
	}

	//TODO: add support for future seasons
	//TODO: add content for info table
	fullSessions := tview.NewTreeNode("Full Race Weekends").SetSelectable(true)
	fullSessions.SetReference(maxi + 1)
	fullSessions.SetExpanded(false)
	fullSessions.SetColor(tcell.ColorYellow)
	root.AddChild(fullSessions)

	season2018 := getSeason("/api/race-season/race_21081096222b4cbb89ae828d37035d1a/")
	var wg1 sync.WaitGroup
	wg1.Add(len(season2018.EventoccurrenceUrls))

	//TODO: move out of main (?)
	//slice of all events (for thread safety)
	events := make([]*tview.TreeNode, len(season2018.EventoccurrenceUrls))
	//iterate through events (GPs)
	for m, eventID := range season2018.EventoccurrenceUrls {
		go func(eventID string, m int) {
			event := getEvent(eventID)
			//if the events actually has saved sassions add it to the tree
			if len(event.SessionoccurrenceUrls) > 0 {
				eventNode := tview.NewTreeNode(event.OfficialName).SetSelectable(true)
				eventNode.SetReference(event.Slug)
				eventNode.SetExpanded(false)
				events[m] = eventNode
				//slice of all sessions (for thread safety)
				sessions := make([]*tview.TreeNode, len(event.SessionoccurrenceUrls))
				var wg2 sync.WaitGroup
				wg2.Add(len(event.SessionoccurrenceUrls))
				//iterate through sessions (FP1, FP2, etc.)
				for n, sessionID := range event.SessionoccurrenceUrls {
					go func(sessionID string, n int) {
						session := getSession(sessionID)
						sessionNode := tview.NewTreeNode(session.Name).SetSelectable(true)
						sessionNode.SetReference(session.Slug)
						sessionNode.SetExpanded(false)
						sessions[n] = sessionNode

						streams := getSessionStreams(session.Slug)
						//slice of all channels (for thread safety)
						channels := make([]*tview.TreeNode, len(streams.Objects[0].ChannelUrls))
						var wg3 sync.WaitGroup
						wg3.Add(len(streams.Objects[0].ChannelUrls))
						//iterate through all available streams for the session (Main feed and driver feeds)
						for i := range streams.Objects[0].ChannelUrls {
							go func(f int) {
								streamPerspective := streams.Objects[0].ChannelUrls[f]
								name := streamPerspective.Name
								if name == "WIF" {
									name = "Main Feed"
								}
								//get url and check for url pattern where separate request needs to be made for the url to be accessible
								url := streamPerspective.Ovps[0].FullStreamURL
								if strings.Contains(url, "https://f1tv.secure.footprint.net/live/") || url == "" {
									//TODO: see if high speed tests can be streamed (curretly return empty string)
									newURL := getProperURL(streamPerspective.Self)
									if len(newURL) > 5 {
										url = newURL
									}
								}
								streamNode := tview.NewTreeNode(name).SetSelectable(true)
								streamNode.SetReference("kjasdlkjhasdlkjhasdlkj")
								streamNode.SetExpanded(false)
								streamNode.SetColor(tcell.ColorGreen)
								channels[f] = streamNode

								//add download and play nodes
								playNode := tview.NewTreeNode("Play with MPV")
								playNode.SetReference(url)
								streamNode.AddChild(playNode)

								downloadNode := tview.NewTreeNode("Download .m3u8")
								downloadNode.SetReference([]string{streamPerspective.Self, event.OfficialName + " - " + session.Name + " - " + name})
								streamNode.AddChild(downloadNode)
								wg3.Done()
							}(i)
						}
						wg3.Wait()
						for _, stream := range channels {
							sessionNode.AddChild(stream)
						}
						wg2.Done()
					}(sessionID, n)
				}
				wg2.Wait()
				for _, session := range sessions {
					eventNode.AddChild(session)
				}
			}
			wg1.Done()
		}(eventID, m)
	}
	wg1.Wait()
	for _, event := range events {
		fullSessions.AddChild(event)
	}

	//load episodes for VOD type
	addEpisodes := func(target *tview.TreeNode, parentType int) {
		//check if episodes of the selected type are not available
		if len(vodTypes.Objects[parentType].ContentUrls) == 0 {
			node := tview.NewTreeNode("no content available")
			node.SetSelectable(false)
			node.SetColor(tcell.ColorRed)
			target.AddChild(node)
			app.Draw()
		} else {
			//store loaded episodes to be sorted at the end
			var episodes []episodeStruct

			//waitgroup so sorting doesn't get skipped
			var wg sync.WaitGroup
			wg.Add(len(vodTypes.Objects[parentType].ContentUrls))

			//blink category node until loading is complete
			doneLoading := false
			go func() {
				for !doneLoading {
					target.SetColor(tcell.ColorBlue)
					app.Draw()
					time.Sleep(200 * time.Millisecond)
					target.SetColor(tcell.ColorYellow)
					app.Draw()
					time.Sleep(200 * time.Millisecond)
				}
			}()

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
			//stop blinking
			doneLoading = true
			app.Draw()
		}
	}

	//display info for the episode or VOD type the cursor is on
	//TODO: are linebreaks/ multiline cells possible?
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		titles := make([]string, 1)
		values := make([][]string, 1)

		reference := node.GetReference()
		if index, ok := reference.(int); ok && index <= maxi {
			//check if selected node is a vod type
			vodTypeStruct := vodTypes.Objects[index]
			fields := reflect.TypeOf(vodTypeStruct)
			values := reflect.ValueOf(vodTypeStruct)
			go fillTable(fields, values)
		} else if ep, ok := reference.(episodeStruct); ok {
			//check if selected node is an episode
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
		if reference == nil {
			return //Selecting the root node does nothing.
		} else if ep, ok := reference.(episodeStruct); ok && len(children) < 1 {
			//TODO: create these on episode node creation?
			//if episode is selected for the first time
			//add nodes to download or play directly
			playNode := tview.NewTreeNode("Play with MPV")
			playNode.SetReference(getProperURL(ep.Items[0]))
			node.AddChild(playNode)

			downloadNode := tview.NewTreeNode("Download .m3u8")
			downloadNode.SetReference([]string{ep.Items[0], ep.Title})
			node.AddChild(downloadNode)
		} else if node.GetText() == "Play with MPV" {
			//if "play" node is selected
			//open URL in MPV
			//TODO: handle mpv not installed
			//TODO: move language selection to config file
			debugPrint(reference.(string))
			cmd := exec.Command("mpv", reference.(string), "--alang=en")
			//create pipe with command output
			stdoutIn, _ := cmd.StdoutPipe()
			//launch command
			cmd.Start()
			//check if window is launched
			scanner := bufio.NewScanner(stdoutIn)
			go func() {
				//check if MPV is opened
				done := false
				go func() {
					for scanner.Scan() {
						sText := scanner.Text()
						if strings.Contains(sText, "Video") {
							break
						}
					}
					done = true
				}()
				//blink the current node from white to blue until MPV is opened
				for !done {
					node.SetColor(tcell.ColorBlue)
					app.Draw()
					time.Sleep(300 * time.Millisecond)
					node.SetColor(tcell.ColorWhite)
					app.Draw()
					time.Sleep(300 * time.Millisecond)
				}
			}()
		} else if node.GetText() == "Download .m3u8" {
			//if "download" node is selected
			node.SetColor(tcell.ColorBlue)
			//download .m3u8
			ref := node.GetReference().([]string)
			downloadAsset(ref[0], ref[1])
		} else if len(children) == 0 {
			//if episodes for category are not loaded yet
			go func() {
				addEpisodes(node, reference.(int))
			}()
		} else {
			//Collapse if visible, expand if collapsed.
			node.SetExpanded(!node.IsExpanded())
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
	flex.AddItem(rowFlex, 0, 3, false)
	rowFlex.AddItem(infoTable, 0, 2, false)
	//flag -d enables debug window
	if len(os.Args) > 1 && os.Args[1] == "-d" {
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

//turns array of driver IDs to their names
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
