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
var episodeMapMutex = sync.RWMutex{}

var app *tview.Application
var infoTable *tview.Table
var debugText *tview.TextView

func main() {
	episodeMap = make(map[string]episodeStruct)

	rootDir := "VOD-Types"
	root := tview.NewTreeNode(rootDir).
		SetColor(tcell.ColorBlue)
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	vodTypes := getVodTypes()

	//TODO does not need to be a function (?)
	//for the base category nodes
	add := func(target *tview.TreeNode) {
		//iterate through vod-types
		for i, vType := range vodTypes.Objects {
			node := tview.NewTreeNode(vType.Name).SetSelectable(true)
			node.SetReference(i)
			target.AddChild(node)
		}
	}

	//build base tree
	add(root)

	//TODO add year subnodes
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
				//sort chronologically
				return year1 < year2 || ((year1 == year2) && (race1 < race2))
			})
			var skippedEpisodes []*tview.TreeNode
			//add episodes to tree
			for _, ep := range episodes {
				node := tview.NewTreeNode(ep.Title).SetSelectable(true)
				node.SetReference(ep)
				node.SetColor(tcell.ColorGreen)
				yearRaceID := ep.DataSourceID[:4]
				//check for year/ race code
				if _, err := strconv.Atoi(yearRaceID); err == nil {
					year := ""
					//TODO: better solution for "2018/19[..]" IDs
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
			app.Draw()
		}
	}

	//display info for the episode or VOD type the cursor is on
	//TODO: are linebreaks/ multiline cells possible?
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if index, ok := reference.(int); ok {
			//check if selected node is a vod type
			vodTypeStruct := vodTypes.Objects[index]
			fields := reflect.TypeOf(vodTypeStruct)
			values := reflect.ValueOf(vodTypeStruct)
			fillTable(fields, values)
		} else if ep, ok := reference.(episodeStruct); ok {
			//check if selected node is an episode
			//get name and value
			fields := reflect.TypeOf(ep)
			values := reflect.ValueOf(ep)
			fillTable(fields, values)
		} else if node.GetText() == "VOD-Types" {
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
			//if episode is selected for the first time
			//add nodes to download or play directly
			playNode := tview.NewTreeNode("Play with MPV")
			playNode.SetReference(getM3U8URL(ep.Items[0]))
			node.AddChild(playNode)

			downloadNode := tview.NewTreeNode("Download .m3u8")
			downloadNode.SetReference([]string{ep.Items[0], ep.Slug})
			node.AddChild(downloadNode)
		} else if node.GetText() == "Play with MPV" {
			//if "play" node is selected
			//open URL in MPV
			//TODO: handle mpv not installed

			cmd := exec.Command("mpv", reference.(string))
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

//takes struct reflect Types and Values and draws them as a table
func fillTable(fields reflect.Type, values reflect.Value) {
	infoTable.Clear()
	rowIndex := 0

	//iterate through  fields
	for fieldIndex := 0; fieldIndex < fields.NumField(); fieldIndex++ {
		field := fields.Field(fieldIndex)
		value := values.Field(fieldIndex)
		//if value is a single string
		if value.Kind() == reflect.String {
			infoTable.SetCell(rowIndex, 1, tview.NewTableCell(field.Name).SetAlign(tview.AlignRight).SetTextColor(tcell.ColorBlue))
			infoTable.SetCell(rowIndex, 2, tview.NewTableCell(value.String()))
			rowIndex++
		} else if value.Kind() == reflect.Slice {
			//if value is a string slice iterate through that too
			infoTable.SetCell(rowIndex, 1, tview.NewTableCell(field.Name).SetAlign(tview.AlignRight).SetTextColor(tcell.ColorRed))
			for j := 0; j < value.Len(); j++ {
				item := value.Index(j)
				infoTable.SetCell(rowIndex, 2, tview.NewTableCell(item.String()))
				rowIndex++
			}
		}
	}
}

//TODO: change before 2030
func fullYear(lastDigits string) string {
	intYear, _ := strconv.Atoi(lastDigits)
	fullYear := ""
	//TODO: change before 2030
	if intYear < 30 {
		fullYear = "20" + lastDigits
	} else {
		fullYear = "19" + lastDigits
	}
	return fullYear
}

//takes year/race ID and returns full year and race nuber as strings
func getYearAndRace(input string) (string, string) {
	year := fullYear(input[:2])
	raceNumber := input[2:4]
	return year, raceNumber
}

//prints to debug window
func debugPrint(s string) {
	fmt.Fprintf(debugText, s+"\n")
}
