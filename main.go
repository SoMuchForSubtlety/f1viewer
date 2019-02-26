package main

import (
	"sort"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {

	rootDir := "VOD-Types"
	root := tview.NewTreeNode(rootDir).
		SetColor(tcell.ColorRed)
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	vodTypes := getVodTypes()

	add := func(target *tview.TreeNode) {
		for i, vType := range vodTypes.Objects {
			node := tview.NewTreeNode(vType.Name).SetSelectable(true)
			node.SetReference(i)
			target.AddChild(node)
		}
	}

	addEpisodes := func(target *tview.TreeNode, parentType int) {

		var episodes []episodeStruct

		for _, episode := range vodTypes.Objects[parentType].ContentUrls {
			episodes = append(episodes, getEpisode(episode))
		}

		sort.Slice(episodes, func(i, j int) bool {
			return episodes[i].Title < episodes[j].Title
		})

		if len(episodes) == 0 {
			node := tview.NewTreeNode("no content available")
			target.AddChild(node)
		} else {
			for _, episode := range episodes {
				node := tview.NewTreeNode(episode.Title).SetSelectable(true)
				node.SetReference(episode)
				target.AddChild(node)
			}
		}
	}

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

	if err := tview.NewApplication().SetRoot(tree, true).Run(); err != nil {
		panic(err)
	}
}
