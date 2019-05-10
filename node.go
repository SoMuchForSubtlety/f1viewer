package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func getPlaybackNodes(title string, epID string) []*tview.TreeNode {
	nodes := make([]*tview.TreeNode, 0)

	// add custom options
	if con.CustomPlaybackOptions != nil {
		for i := range con.CustomPlaybackOptions {
			com := con.CustomPlaybackOptions[i]
			if len(com.Commands) > 0 {
				var context commandContext
				context.EpID = epID
				context.CustomOptions = com
				context.Title = title
				customNode := tview.NewTreeNode(com.Title)
				customNode.SetReference(context)
				nodes = append(nodes, customNode)
			}
		}
	}

	playNode := tview.NewTreeNode("Play with MPV")
	playNode.SetReference(epID)
	nodes = append(nodes, playNode)

	downloadNode := tview.NewTreeNode("Download .m3u8")
	downloadNode.SetReference([]string{epID, title})
	nodes = append(nodes, downloadNode)

	if checkArgs("-d") {
		streamNode := tview.NewTreeNode("GET URL")
		streamNode.SetReference(epID)
		nodes = append(nodes, streamNode)
	}
	return nodes
}

func getLiveNode() (bool, *tview.TreeNode, error) {
	var sessionNode *tview.TreeNode
	home, err := getHomepageContent()
	if err != nil {
		return false, sessionNode, err
	}
	firstContent := home.Objects[0].Items[0].ContentURL.Items[0].ContentURL.Self
	if strings.Contains(firstContent, "/api/event-occurrence/") {
		event, err := getEvent(firstContent)
		if err != nil {
			return false, sessionNode, err
		}
		for _, sessionID := range event.SessionoccurrenceUrls {
			session, err := getSession(sessionID)
			if err != nil {
				return false, sessionNode, err
			}
			if session.Status == "live" {
				streams, err := getSessionStreams(session.Slug)
				if err != nil {
					return false, sessionNode, err
				}
				sessionNode = tview.NewTreeNode(session.Name + " - LIVE").
					SetSelectable(true).
					SetColor(tcell.ColorRed).
					SetReference(streams).
					SetExpanded(false)
				channels := getPerspectiveNodes(streams.Objects[0].ChannelUrls)
				for _, stream := range channels {
					sessionNode.AddChild(stream)
				}
				return true, sessionNode, nil
			}
		}
	}
	return false, sessionNode, nil
}

// blinks node until bool is changed
// TODO replace done bool with channel?
func blinkNode(node *tview.TreeNode, done *bool, originalColor tcell.Color) {
	colors := []tcell.Color{tcell.ColorRed, tcell.ColorOrange, tcell.ColorYellow, tcell.ColorGreen, tcell.ColorBlue, tcell.ColorIndigo, tcell.ColorViolet}
	originalText := node.GetText()
	node.SetText("loading...")
	for !*done {
		for _, color := range colors {
			if *done {
				break
			}
			node.SetColor(color)
			app.Draw()
			time.Sleep(100 * time.Millisecond)
		}
	}
	node.SetColor(originalColor)
	node.SetText(originalText)
	app.Draw()
}

// returns node for every event (Australian GP, Bahrain GP, etc.)
func getEventNodes(season seasonStruct) ([]*tview.TreeNode, error) {
	errChan := make(chan error)
	events := make([]*tview.TreeNode, len(season.EventoccurrenceUrls))
	// iterate through events
	for m, eventID := range season.EventoccurrenceUrls {
		go func(eventID string, m int) {
			event, err := getEvent(eventID)
			if err != nil {
				errChan <- err
				return
			}
			// if the events actually has saved sassions add it to the tree
			if len(event.SessionoccurrenceUrls) > 0 {
				eventNode := tview.NewTreeNode(strings.Replace(event.OfficialName, "™", "", -1)).SetSelectable(true)
				eventNode.SetReference(event)
				events[m] = eventNode
			}
			errChan <- nil
		}(eventID, m)
	}
	for index := 0; index < len(season.EventoccurrenceUrls); index++ {
		select {
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		}
	}
	return events, nil
}

// returns node for every session (FP1, FP2, etc.)
func getSessionNodes(event eventStruct) ([]*tview.TreeNode, error) {
	errChan := make(chan error)
	sessions := make([]*tview.TreeNode, len(event.SessionoccurrenceUrls))
	bonusIDs := make([][]string, len(event.SessionoccurrenceUrls))
	// iterate through sessions
	for n, sessionID := range event.SessionoccurrenceUrls {
		go func(sessionID string, n int) {
			session, err := getSession(sessionID)
			if err != nil {
				errChan <- err
				return
			}
			bonusIDs[n] = session.ContentUrls
			if session.Status != "upcoming" && session.Status != "expired" {
				streams, err := getSessionStreams(session.Slug)
				if err != nil {
					errChan <- err
					return
				}
				sessionNode := tview.NewTreeNode(session.Name).
					SetSelectable(true).
					SetReference(streams).
					SetExpanded(false)
				if session.Status == "live" {
					sessionNode.SetText(session.Name + " - LIVE").
						SetColor(tcell.ColorRed)
				}
				sessions[n] = sessionNode

				channels := getPerspectiveNodes(streams.Objects[0].ChannelUrls)
				for _, stream := range channels {
					sessionNode.AddChild(stream)
				}
			}
			errChan <- nil
		}(sessionID, n)
	}
	for index := 0; index < len(event.SessionoccurrenceUrls); index++ {
		select {
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		}
	}
	var allIDs []string
	for _, idList := range bonusIDs {
		allIDs = append(allIDs, idList...)
	}
	if len(allIDs) > 0 {
		bonusNode := tview.NewTreeNode("Bonus Content").SetSelectable(true).SetExpanded(false).SetReference("bonus")
		episodes, err := getEpisodeNodes(allIDs)
		if err != nil {
			return nil, err
		}
		appendNodes(bonusNode, episodes...)
		return append(sessions, bonusNode), nil
	}
	return sessions, nil
}

// returns nodes for every perspective (main feed, data feed, drivers, etc.)
func getPerspectiveNodes(perspectives []channelUrlsStruct) []*tview.TreeNode {
	channels := make([]*tview.TreeNode, len(perspectives))
	// iterate through all available streams for the session
	for i := range perspectives {
		streamPerspective := perspectives[i]
		name := streamPerspective.Name
		if len(streamPerspective.DriverUrls) > 0 {
			number := streamPerspective.DriverUrls[0].DriverRacingnumber
			name = fmt.Sprintf("%4v "+name, "("+strconv.Itoa(number)+")")
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
		streamNode := tview.NewTreeNode(name).
			SetSelectable(true).
			SetReference(streamPerspective).
			SetColor(tcell.ColorGreen)
		channels[i] = streamNode
	}
	sort.Slice(channels, func(i, j int) bool {
		return !strings.Contains(channels[i].GetText(), "(")
	})
	return channels
}

// returns nodes for every season of "Full Race Weekends"
func getSeasonNodes() ([]*tview.TreeNode, error) {
	seasons, err := getSeasons()
	if err != nil {
		return nil, err
	}
	var nodes []*tview.TreeNode
	for _, s := range seasons.Seasons {
		if s.HasContent {
			seasonNode := tview.NewTreeNode(s.Name).
				SetReference(s)
			nodes = append(nodes, seasonNode)
		}
	}
	return nodes, nil
}

// add episodes to VOD type
func getEpisodeNodes(IDs []string) ([]*tview.TreeNode, error) {
	var skippedEpisodes []*tview.TreeNode
	var yearNodes []*tview.TreeNode

	eps, err := loadEpisodes(IDs)
	if err != nil {
		return nil, err
	}
	episodes := sortEpisodes(eps)
	// add loaded and sorted episodes to tree
	for _, ep := range episodes {
		if len(ep.Items) < 1 {
			continue
		}
		node := tview.NewTreeNode(ep.Title).SetSelectable(true).
			SetReference(ep).
			SetColor(tcell.ColorGreen)
		// check for year/ race code
		if year, _, err := getYearAndRace(ep.DataSourceID); err == nil {
			// check if there is a node for the specified year, if not create one
			fatherFound := false
			var fatherNode *tview.TreeNode
			for _, subNode := range yearNodes {
				if subNode.GetReference() == year {
					fatherNode = subNode
					fatherFound = true
				}
			}
			if !fatherFound {
				yearNode := tview.NewTreeNode(year).
					SetSelectable(true).
					SetReference(year).
					SetExpanded(false)
				yearNodes = append(yearNodes, yearNode)
				fatherNode = yearNode
			}
			fatherNode.AddChild(node)
		} else {
			// save episodes with no year/race ID to be added at the end
			skippedEpisodes = append(skippedEpisodes, node)
		}
	}
	if len(yearNodes) == 1 {
		return append(yearNodes[0].GetChildren(), skippedEpisodes...), nil
	}
	return append(yearNodes, skippedEpisodes...), nil
}

func getVodTypeNodes() ([]*tview.TreeNode, error) {
	var nodes []*tview.TreeNode
	var err error
	vodTypes, err = getVodTypes()
	if err != nil {
		return nil, err
	}
	for i, vType := range vodTypes.Objects {
		if len(vType.ContentUrls) > 0 {
			node := tview.NewTreeNode(vType.Name).
				SetSelectable(true).
				SetReference(i).
				SetColor(tcell.ColorYellow)
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

// appends children to parent
func appendNodes(parent *tview.TreeNode, children ...*tview.TreeNode) {
	if children != nil {
		for _, node := range children {
			if node != nil {
				parent.AddChild(node)
			}
		}
	}
}

// probably needs mutex
func insertNodeAtTop(parentNode *tview.TreeNode, childNode *tview.TreeNode) {
	children := parentNode.GetChildren()
	children = append([]*tview.TreeNode{childNode}, children...)
	parentNode.SetChildren(children)
}

func removeChild(node, childNode *tview.TreeNode) error {
	if node == nil || childNode == nil {
		return errors.New("node can't be nil")
	}
	children := node.GetChildren()
	for i, child := range children {
		if child != nil && (child.GetReference() == childNode.GetReference()) && (child.GetText() == childNode.GetText()) {
			copy(children[i:], children[i+1:])
			children = children[:len(children)-1]
			node.ClearChildren()
			node.SetChildren(children)
			return nil
		}
	}
	return errors.New("child not found")
}

// adds and removes live session nodes to root
func liveHandler(root *tview.TreeNode) {
	for {
		debugPrint("checking for live session")
		isLive, liveNode, err := getLiveNode()
		if err != nil {
			debugPrint("error looking for live session:")
			debugPrint(err.Error())
			time.Sleep(time.Second * time.Duration(con.LiveRetryTimeout))
		} else if isLive {
			insertNodeAtTop(root, liveNode)
			return
		} else if con.LiveRetryTimeout < 0 {
			debugPrint("no live session found")
			return
		} else {
			debugPrint("no live session found")
			time.Sleep(time.Second * time.Duration(con.LiveRetryTimeout))
		}
	}
}
