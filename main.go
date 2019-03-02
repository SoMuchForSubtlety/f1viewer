package main

import (
	"bufio"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var app *tview.Application
var infoTable *tview.Table

func main() {

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
			for i := range vodTypes.Objects[parentType].ContentUrls {
				//multithread loading the apisodes and add them to the tree dynamically
				//TODO: limit the number of threads
				go func(i int) {
					ep := getEpisode(vodTypes.Objects[parentType].ContentUrls[i])
					episodes = append(episodes, ep)
					node := tview.NewTreeNode(ep.Title).SetSelectable(true)
					node.SetReference(ep)
					node.SetColor(tcell.ColorGreen)
					target.AddChild(node)
					defer wg.Done()
					app.Draw()
				}(i)
			}
			//wait for loading to complete, then sort
			wg.Wait()
			sort.Slice(episodes, func(i, j int) bool {
				return episodes[i].Title < episodes[j].Title
			})
			//purge childrean and re-add them in sorted order
			target.ClearChildren()
			for _, ep := range episodes {
				node := tview.NewTreeNode(ep.Title).SetSelectable(true)
				node.SetReference(ep)
				node.SetColor(tcell.ColorGreen)
				target.AddChild(node)
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
	flex := tview.NewFlex()
	infoTable = tview.NewTable()
	infoTable.SetBorder(true).SetTitle(" Info ")
	flex.AddItem(tree, 0, 2, true)
	flex.AddItem(infoTable, 0, 3, false)
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
