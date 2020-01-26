package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func (session *viewerSession) getFullSessionsNode() *tview.TreeNode {
	fullSessions := tview.NewTreeNode("Full Race Weekends").
		SetColor(activeTheme.CategoryNodeColor)
	fullSessions.SetSelectedFunc(session.withBlink(fullSessions, func() {
		fullSessions.SetSelectedFunc(nil)
		seasons, err := session.getSeasonNodes()
		if err != nil {
			session.logError(err)
		} else {
			appendNodes(fullSessions, seasons...)
		}
	}))
	return fullSessions
}

func (session *viewerSession) getPlaybackNodes(sessionTitles titles, epID string) []*tview.TreeNode {
	nodes := make([]*tview.TreeNode, 0)

	// add custom options
	if session.cfg.CustomPlaybackOptions != nil {
		for i := range session.cfg.CustomPlaybackOptions {
			com := session.cfg.CustomPlaybackOptions[i]
			if len(com.Command) > 0 {
				context := commandContext{
					Titles:        sessionTitles,
					EpID:          epID,
					CustomOptions: com,
				}
				customNode := tview.NewTreeNode(com.Title).
					SetColor(activeTheme.ActionNodeColor)
				customNode.SetSelectedFunc(func() {
					go func() {
						err := session.runCustomCommand(context)
						if err != nil {
							session.logError(err)
						}
					}()
				})
				nodes = append(nodes, customNode)
			}
		}
	}

	playNode := tview.NewTreeNode("Play with MPV").
		SetColor(activeTheme.ActionNodeColor)
	playNode.SetSelectedFunc(func() {
		url, err := getPlayableURL(epID)
		if err != nil {
			session.logError(err)
			return
		}
		cmd := exec.Command("mpv", url, "--alang="+session.cfg.Lang, "--start=0", "--quiet")
		err = session.runCmd(cmd)
		if err != nil {
			session.logError(err)
		}
	})
	nodes = append(nodes, playNode)

	downloadNode := tview.NewTreeNode("Download .m3u8").
		SetColor(activeTheme.ActionNodeColor)
	downloadNode.SetSelectedFunc(func() {
		downloadNode.SetColor(activeTheme.ActionNodeColor)
		url, err := getPlayableURL(epID)
		if err != nil {
			session.logError(err)
			return
		}
		path, _, err := session.cfg.downloadAsset(url, sessionTitles.String())
		if err != nil {
			session.logError(err)
			return
		}
		session.logInfo("Saved \"", filepath.Base(path), "\"")
	})
	nodes = append(nodes, downloadNode)

	streamNode := tview.NewTreeNode("Copy URL to clipboard").
		SetColor(activeTheme.ActionNodeColor)
	streamNode.SetSelectedFunc(func() {
		url, err := getPlayableURL(epID)
		if err != nil {
			session.logError(err)
			return
		}
		err = clipboard.WriteAll(url)
		if err != nil {
			session.logError(err)
			return
		}
		session.logInfo("URL copied to clipboard")
	})
	nodes = append(nodes, streamNode)
	return nodes
}

func (session *viewerSession) getLiveNode() (bool, *tview.TreeNode, error) {
	var sessionNode *tview.TreeNode
	home, err := getHomepageContent()
	if err != nil {
		return false, sessionNode, err
	}
	var contentURL string
	found := false
	for _, item := range home.Items {
		contentURL = item.ContentURL
		if strings.Contains(contentURL, "/api/event-occurrence/") {
			found = true
			break
		}
	}
	if found {
		var t titles
		event, err := getEvent(contentURL)
		if err != nil {
			return false, sessionNode, err
		}
		t.EventTitle = event.Name
		for _, sessionID := range event.SessionoccurrenceUrls {
			s, err := getSession(sessionID)
			if err != nil {
				return false, sessionNode, err
			}
			st := t
			st.SessionTitle = s.Name
			if s.Status == "live" {
				streams, err := getSessionStreams(s.UID)
				if err != nil {
					return false, sessionNode, err
				}
				sessionNode = tview.NewTreeNode(s.Name + " - LIVE").
					SetColor(activeTheme.LiveColor).
					SetExpanded(false)
				channels := session.getPerspectiveNodes(st, streams)
				for _, stream := range channels {
					sessionNode.AddChild(stream)
				}
				return true, sessionNode, nil
			}
		}
	}
	return false, sessionNode, nil
}

func (session *viewerSession) getEventNodes(season seasonStruct) ([]*tview.TreeNode, error) {
	// TODO: refactor
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
				eventNode := tview.NewTreeNode(event.Name).SetSelectable(true)
				eventNode.SetSelectedFunc(session.withBlink(eventNode, func() {
					eventNode.SetSelectedFunc(nil)
					sessions, err := session.getSessionNodes(titles{SeasonTitle: season.Name, CategoryTitle: "Full Race Weekends"}, event)
					if err != nil {
						session.logError(err)
					} else {
						appendNodes(eventNode, sessions...)
					}
					if len(eventNode.GetChildren()) == 0 {
						eventNode.AddChild(tview.NewTreeNode("no content").SetColor(activeTheme.NoContentColor))
					}
				}))
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

func (session *viewerSession) getSessionNodes(t titles, event eventStruct) ([]*tview.TreeNode, error) {
	sessions := make([]*tview.TreeNode, 0)
	bonusIDs := make([]string, 0)
	sessionsData, err := getSessions(event.SessionoccurrenceUrls)
	if err != nil {
		return nil, err
	}
	t.EventTitle = event.Name
	for _, s := range sessionsData {
		st := t
		st.SessionTitle = s.Name
		bonusIDs = append(bonusIDs, s.ContentUrls...)
		if s.Status != "upcoming" && s.Status != "expired" {
			s := s
			sessionNode := tview.NewTreeNode(s.Name).
				SetSelectable(true)
			sessionNode.SetSelectedFunc(session.withBlink(sessionNode, func() {
				sessionNode.SetSelectedFunc(nil)
				streams, err := getSessionStreams(s.UID)
				if err != nil {
					session.logError(err)
					return
				}
				channels := session.getPerspectiveNodes(st, streams)
				appendNodes(sessionNode, channels...)
			}))
			if s.Status == "live" {
				sessionNode.SetText(s.Name + " - LIVE").
					SetColor(activeTheme.LiveColor)
			}
			sessions = append(sessions, sessionNode)
		}
	}
	if len(bonusIDs) > 0 {
		bonusNode := tview.NewTreeNode("Bonus Content").SetExpanded(false)
		episodes, err := session.getEpisodeNodes(t, bonusIDs)
		if err != nil {
			return nil, err
		}
		appendNodes(bonusNode, episodes...)
		return append(sessions, bonusNode), nil
	}
	return sessions, nil
}

func (session *viewerSession) getPerspectiveNodes(title titles, perspectives []channel) []*tview.TreeNode {
	var channels []*tview.TreeNode
	teams := make(map[string]*tview.TreeNode)
	var teamsContasiner *tview.TreeNode
	for _, streamPerspective := range perspectives {
		streamPerspective := streamPerspective
		name := streamPerspective.Name
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
		newTitle := title
		newTitle.PerspectiveTitle = name
		if len(streamPerspective.DriverUrls) > 0 {
			number := streamPerspective.DriverUrls[0].DriverRacingnumber
			name = fmt.Sprintf("(%2d) %s", number, name)
		}

		streamNode := tview.NewTreeNode(name).
			SetColor(activeTheme.ItemNodeColor)

		streamNode.SetSelectedFunc(func() {
			streamNode.SetSelectedFunc(nil)
			nodes := session.getPlaybackNodes(newTitle, streamPerspective.Self)
			appendNodes(streamNode, nodes...)
		})
		if len(streamPerspective.DriverUrls) > 0 {
			if teamsContasiner == nil {
				teamsContasiner = tview.NewTreeNode("Teams").SetExpanded(false)
			}
			team := streamPerspective.DriverUrls[0].TeamURL
			teamNode, ok := teams[team.Name]
			if !ok {
				teamNode = tview.NewTreeNode(team.Name).SetExpanded(false)
				color, err := strconv.ParseInt(team.Colour[1:], 16, 32)
				if err == nil {
					teamNode.SetColor(tcell.NewHexColor(int32(color)))
				}
				teams[team.Name] = teamNode
				teamsContasiner.AddChild(teamNode)
			}
			teamNode.AddChild(streamNode)
		} else {
			channels = append(channels, streamNode)
		}
	}
	if teamsContasiner != nil {
		channels = append(channels, teamsContasiner)
	}
	return channels
}

func (session *viewerSession) getSeasonNodes() ([]*tview.TreeNode, error) {
	seasons, err := getSeasons()
	if err != nil {
		return nil, err
	}
	var nodes []*tview.TreeNode
	for _, s := range seasons.Seasons {
		if s.HasContent {
			s := s
			seasonNode := tview.NewTreeNode(s.Name)
			seasonNode.SetSelectedFunc(session.withBlink(seasonNode, func() {
				seasonNode.SetSelectedFunc(nil)
				events, err := session.getEventNodes(s)
				if err != nil {
					session.logError(err)
				}
				appendNodes(seasonNode, events...)
			}))
			nodes = append(nodes, seasonNode)
		}
	}
	return nodes, nil
}

func (session *viewerSession) getEpisodeNodes(title titles, IDs []string) ([]*tview.TreeNode, error) {
	var nodes []*tview.TreeNode
	var yearNodes []*tview.TreeNode
	yearNodesMap := make(map[string]*tview.TreeNode)
	eps, err := session.loadEpisodes(IDs)
	if err != nil {
		return nil, err
	}
	episodes := sortEpisodes(eps)
	for _, ep := range episodes {
		ep := ep
		if len(ep.Items) < 1 {
			continue
		}
		node := tview.NewTreeNode(ep.Title).
			SetColor(activeTheme.ItemNodeColor)
		tempTitle := title
		tempTitle.EpisodeTitle = ep.Title
		node.SetSelectedFunc(func() {
			node.SetSelectedFunc(nil)
			nodes := session.getPlaybackNodes(tempTitle, ep.Items[0])
			appendNodes(node, nodes...)
		})
		if year, _, err := getYearAndRace(ep.DataSourceID); err == nil {
			yearNode, ok := yearNodesMap[year]
			if !ok {
				yearNode = tview.NewTreeNode(year).
					SetExpanded(false)
				yearNodesMap[year] = yearNode
				yearNodes = append(yearNodes, yearNode)
			}
			yearNode.AddChild(node)
		} else {
			nodes = append(nodes, node)
		}
	}
	return append(yearNodes, nodes...), nil
}

func (session *viewerSession) getVodTypeNodes() ([]*tview.TreeNode, error) {
	var nodes []*tview.TreeNode
	vodTypes, err := getVodTypes()
	if err != nil {
		return nil, err
	}
	for _, vType := range vodTypes.Objects {
		vType := vType
		catTitle := vType.Name
		if len(vType.ContentUrls) > 0 {
			node := tview.NewTreeNode(vType.Name).
				SetColor(activeTheme.CategoryNodeColor)
			node.SetSelectedFunc(session.withBlink(node, func() {
				node.SetSelectedFunc(nil)
				episodes, err := session.getEpisodeNodes(titles{CategoryTitle: catTitle}, vType.ContentUrls)
				if err != nil {
					session.logError(err)
				} else {
					appendNodes(node, episodes...)
				}
			}))
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (session *viewerSession) getCollectionsNode() *tview.TreeNode {
	node := tview.NewTreeNode("Collections").SetColor(activeTheme.CategoryNodeColor)
	node.SetSelectedFunc(session.withBlink(node, func() {
		node.SetSelectedFunc(nil)
		list, err := getCollectionList()
		if err != nil {
			session.logError("could not load collections: ", err)
		}
		for _, coll := range list.Objects {
			child := tview.NewTreeNode(coll.Title)
			collID := coll.UID
			child.SetSelectedFunc(session.withBlink(child, func() {
				child.SetSelectedFunc(nil)
				var nodes []*tview.TreeNode
				nodes, err = session.getCollectionContent(collID)
				if err != nil {
					session.logError(err)
				} else if len(nodes) > 0 {
					appendNodes(child, nodes...)
				} else {
					child.AddChild(tview.NewTreeNode("no content").SetColor(activeTheme.NoContentColor))
				}
			}))
			node.AddChild(child)
		}
	}))
	return node
}

func (session *viewerSession) getCollectionContent(id string) ([]*tview.TreeNode, error) {
	coll, err := getCollection(id)
	if err != nil {
		return nil, err
	}
	var epIDs []string
	for _, ep := range coll.Items {
		epIDs = append(epIDs, ep.ContentURL)
	}
	return session.getEpisodeNodes(titles{CategoryTitle: coll.Title}, epIDs)
}

func appendNodes(parent *tview.TreeNode, children ...*tview.TreeNode) {
	if children != nil {
		for _, node := range children {
			if node != nil {
				parent.AddChild(node)
			}
		}
	}
}

func insertNodeAtTop(parentNode *tview.TreeNode, childNode *tview.TreeNode) {
	children := parentNode.GetChildren()
	children = append([]*tview.TreeNode{childNode}, children...)
	parentNode.SetChildren(children)
}

func (session *viewerSession) toggleVisibility(node *tview.TreeNode) {
	if len(node.GetChildren()) > 0 {
		node.SetExpanded(!node.IsExpanded())
	}
}
