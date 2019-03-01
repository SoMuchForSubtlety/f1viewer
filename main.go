package main

import (
	"fmt"
	"os/exec"
	"reflect"
	"sort"
	"sync"

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

	//display info for the episode the cursor is on
	//TODO: are linebreaks possible?
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		infoTable.Clear()
		//check if selected node is an episode
		reference := node.GetReference()
		if ep, ok := reference.(episodeStruct); ok {
			//get name and value
			fields := reflect.TypeOf(ep)
			values := reflect.ValueOf(ep)
			num := fields.NumField()
			num2 := 0

			//write to table
			for i := 0; i < num; i++ {
				field := fields.Field(i)
				value := values.Field(i)
				//if value is a single string
				if value.Kind() == reflect.String {
					infoTable.SetCell(num2, 1, tview.NewTableCell(field.Name).SetAlign(tview.AlignRight).SetTextColor(tcell.ColorBlue))
					infoTable.SetCell(num2, 2, tview.NewTableCell(value.String()))
					num2++
				} else if value.Kind() == reflect.Slice {
					//if value is a string slice iterate through that too
					infoTable.SetCell(num2, 1, tview.NewTableCell(field.Name).SetAlign(tview.AlignRight).SetTextColor(tcell.ColorRed))
					for j := 0; j < value.Len(); j++ {
						item := value.Index(j)
						infoTable.SetCell(num2, 2, tview.NewTableCell(item.String()))
						num2++
					}
				}
			}
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
			//add nodes to download or play directly
			playNode := tview.NewTreeNode("Play with MPV")
			playNode.SetReference(getM3U8URL(ep.Items[0]))
			node.AddChild(playNode)

			downloadNode := tview.NewTreeNode("download .m3u8")
			downloadNode.SetReference([]string{ep.Items[0], ep.Title})
			node.AddChild(downloadNode)
		} else if node.GetText() == "Play with MPV" {
			//if "play" node is selected
			//open URL in MPV
			mpvPlay(reference.(string))
		} else if node.GetText() == "download .m3u8" {
			//if "download" node is selected
			//download .m3u8
			slice1 := node.GetReference().([]string)
			fmt.Println(len(slice1))
			downloadAsset(slice1[0], slice1[1])
		} else if len(children) == 0 {
			//if episodes for category are not loaded yet
			//TODO: limit threads that are spawned
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
	infoTable.SetBorder(true).SetTitle(" Episode Info ")
	flex.AddItem(tree, 0, 2, true)
	flex.AddItem(infoTable, 0, 3, false)
	app.SetRoot(flex, true).Run()
}

//launches MPV
func mpvPlay(m3u8URL string) {
	cmd := exec.Command("mpv", m3u8URL)
	cmd.Start()
}
