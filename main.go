package main

import (
	"sort"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

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

	//for episode nodes
	addEpisodes := func(target *tview.TreeNode, parentType int) {

		//create slice of episode IDs and sort it by date
		var episodes []episodeStruct
		for _, episode := range vodTypes.Objects[parentType].ContentUrls {
			episodes = append(episodes, getEpisode(episode))
		}

		sort.Slice(episodes, func(i, j int) bool {
			return episodes[i].Title < episodes[j].Title
		})

		//add them to the tree/ add no content available note
		if len(episodes) == 0 {
			node := tview.NewTreeNode("no content available")
			node.SetSelectable(false)
			node.SetColor(tcell.ColorRed)
			target.AddChild(node)
		} else {
			for _, episode := range episodes {
				node := tview.NewTreeNode(episode.Title).SetSelectable(true)
				node.SetReference(episode)
				node.SetColor(tcell.ColorGreen)
				target.AddChild(node)
			}
		}
	}

	//build base tree
	add(root)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		children := node.GetChildren()
		if reference == nil {
			return // Selecting the root node does nothing.
		} else if ep, ok := reference.(episodeStruct); ok && len(children) < 1 {
			downloadAsset(ep.Items[0], ep.Title)
			newNode := tview.NewTreeNode("content downloaded")
			node.AddChild(newNode)
			return
		}
		// Load and show files in this directory.
		if len(children) == 0 {
			addEpisodes(node, reference.(int))
		} else {
			// Collapse if visible, expand if collapsed.
			node.SetExpanded(!node.IsExpanded())
		}
	})

	//start UI
	if err := tview.NewApplication().SetRoot(tree, true).Run(); err != nil {
		panic(err)
	}
}
