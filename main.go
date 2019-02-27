package main

import (
	"sort"
	"sync"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var app *tview.Application

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

	//what happens when a node is selected
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		children := node.GetChildren()
		if reference == nil {
			return //Selecting the root node does nothing.
		} else if ep, ok := reference.(episodeStruct); ok && len(children) < 1 {
			//TODO replace ep.title with proper title
			downloadAsset(ep.Items[0], ep.Title)
			newNode := tview.NewTreeNode("content downloaded")
			node.AddChild(newNode)
			return
		}
		if len(children) == 0 {
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
	app.SetRoot(tree, true).Run()
}
