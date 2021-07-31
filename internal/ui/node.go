package ui

import (
	"errors"
	"fmt"
	"sync"

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/cmd"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/util"
	"github.com/SoMuchForSubtlety/f1viewer/v2/pkg/f1tv/v2"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NodeMetadata is used for treenode references and holds metadata about a node
type NodeMetadata struct {
	nodeType NodeType
	id       string
	metadata cmd.MetaData
	sync.Mutex
}

// NodeType indicates a treeview node's type
type NodeType int

// Node types for Metadata
const (
	CategoryNode NodeType = iota
	EventNode
	PlayableNode
	StreamNode
	ActionNode
	MiscNode
	CollectionNode
)

func (s *UIState) TreeInputHanlder(keyEvent *tcell.EventKey) *tcell.EventKey {
	// only listen for 'r' key
	if keyEvent.Key() != tcell.KeyRune || (keyEvent.Rune() != 'r' && keyEvent.Rune() != 'q') {
		return keyEvent
	}

	if keyEvent.Rune() == 'q' {
		s.app.Stop()
		return nil
	}

	node := s.treeView.GetCurrentNode()
	metadata, err := getMetadata(node)
	if err != nil {
		s.logger.Error(err)
	}

	switch metadata.nodeType {
	case EventNode:
		// TODO: implement refreshing again
		return nil
	default:
		return keyEvent
	}
}

func getMetadata(node *tview.TreeNode) (*NodeMetadata, error) {
	if node == nil {
		return &NodeMetadata{}, errors.New("node is nil")
	}
	switch v := node.GetReference().(type) {
	case *NodeMetadata:
		return v, nil
	default:
		return &NodeMetadata{}, fmt.Errorf("Node has reference of unexpected type %T", v)
	}
}

func (s *UIState) getPlaybackNodes(sessionTitles cmd.MetaData, getURL func() (string, error)) []*tview.TreeNode {
	nodes := make([]*tview.TreeNode, 0)

	for _, c := range s.cmd.Commands {
		nodes = append(nodes, s.createCommandNode(sessionTitles, getURL, c))
	}

	// for _, c := range s.cmd.MultiCommads {
	// 	nodes = append(nodes, s.createCommandNode(sessionTitles, getURL, c))
	// }

	clipboardNode := tview.NewTreeNode("Copy URL to clipboard").
		SetColor(activeTheme.ActionNodeColor).
		SetReference(&NodeMetadata{nodeType: ActionNode, metadata: sessionTitles})
	clipboardNode.SetSelectedFunc(func() {
		url, err := getURL()
		if err != nil {
			s.logger.Error(err)
			return
		}
		err = clipboard.WriteAll(url)
		if err != nil {
			s.logger.Error(err)
			return
		}
		s.logger.Info("URL copied to clipboard")
	})
	nodes = append(nodes, clipboardNode)
	return nodes
}

func (s *UIState) createCommandNode(t cmd.MetaData, getURL func() (string, error), c cmd.Command) *tview.TreeNode {
	context := cmd.CommandContext{
		MetaData:      t,
		CustomOptions: c,
		URL:           getURL,
	}
	node := tview.NewTreeNode(c.Title).
		SetColor(activeTheme.ActionNodeColor).
		SetReference(&NodeMetadata{nodeType: ActionNode, metadata: t})
	node.SetSelectedFunc(func() {
		go func() {
			err := s.cmd.RunCommand(context)
			if err != nil {
				s.logger.Error(err)
			}
		}()
	})

	return node
}

func (s *UIState) getLiveNode() (bool, *tview.TreeNode, error) {
	liveVideos, err := s.v2.GetLiveVideoContainers()
	if err != nil || len(liveVideos) == 0 {
		return false, nil, err
	}

	var nodes []*tview.TreeNode
	for _, v := range liveVideos {
		streamNode := s.v2ContentNode(v)
		streamNode.SetText(v.Metadata.Title + " - LIVE").SetColor(activeTheme.LiveColor)
		nodes = append(nodes, streamNode)
	}

	if len(nodes) > 1 {
		allLive := tview.NewTreeNode("LIVE").
			SetColor(activeTheme.LiveColor).
			SetExpanded(false)
		appendNodes(allLive, nodes...)
		return true, allLive, nil
	} else {
		return true, nodes[0], nil
	}
}

func (s *UIState) getPageNodes(id f1tv.PageID) []*tview.TreeNode {
	s.logger.Infof("loading page %d", id)
	headings, bundles, err := s.v2.GetPageContent(id)
	if err != nil {
		s.logger.Error(err)
		return nil
	}

	seenIDs := make(map[f1tv.PageID]struct{})

	var headingNodes []*tview.TreeNode
	for _, h := range headings {
		h := h
		title := util.FirstNonEmptyString(h.Metadata.Label, h.Metadata.Title, h.RetrieveItems.ResultObj.MeetingName)

		if title == "" {
			for _, v := range h.RetrieveItems.ResultObj.Containers {
				headingNodes = append(headingNodes, s.v2ContentNode(v))
			}
		} else {
			headingNode := tview.NewTreeNode(title).
				SetColor(activeTheme.CategoryNodeColor).
				SetReference(&NodeMetadata{nodeType: CategoryNode}).
				SetExpanded(false)

			for _, v := range h.RetrieveItems.ResultObj.Containers {
				headingNode.AddChild(s.v2ContentNode(v))
			}
			headingNodes = append(headingNodes, headingNode)
		}
	}
	for _, b := range bundles {
		b := b
		// don't show the same page
		_, ok := seenIDs[b.ID]
		if !ok {
			seenIDs[b.ID] = struct{}{}
		} else {
			continue
		}
		headingNode := tview.NewTreeNode(b.Title).
			SetColor(activeTheme.FolderNodeColor).
			SetReference(&NodeMetadata{nodeType: CategoryNode}).
			SetExpanded(false)

		headingNode.SetSelectedFunc(s.withBlink(headingNode, func() {
			headingNode.SetSelectedFunc(nil)
			appendNodes(headingNode, s.getPageNodes(b.ID)...)
			headingNode.SetExpanded(true)
		}, nil))
		headingNodes = append(headingNodes, headingNode)
	}

	return headingNodes
}

func appendNodes(parent *tview.TreeNode, children ...*tview.TreeNode) {
	for _, node := range children {
		if node != nil {
			parent.AddChild(node)
		}
	}
}

func insertNodeAtTop(parentNode *tview.TreeNode, childNode *tview.TreeNode) {
	children := parentNode.GetChildren()
	children = append([]*tview.TreeNode{childNode}, children...)
	parentNode.SetChildren(children)
}

func (s *UIState) toggleVisibility(node *tview.TreeNode) {
	if len(node.GetChildren()) > 0 {
		node.SetExpanded(!node.IsExpanded())
	}
}
