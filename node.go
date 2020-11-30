package main

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var errNoSessions = errors.New("event has not past or live events")

// NodeMetadata is used for treenode references and holds metadata about a node
type NodeMetadata struct {
	nodeType NodeType
	id       string
	metadata MetaData
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

func (session *viewerSession) nodeRefresh(keyEvent *tcell.EventKey) *tcell.EventKey {
	// only listen for 'r' key
	if keyEvent.Key() != tcell.KeyRune || keyEvent.Rune() != 'r' {
		return keyEvent
	}

	node := session.tree.GetCurrentNode()
	metadata, err := getMetadata(node)
	if err != nil {
		session.logError(err)
	}

	switch metadata.nodeType {
	case EventNode:
		go func() {
			metadata.Lock()
			updateFunc := func() { session.updateEvent(node, metadata) }
			session.withBlink(node, updateFunc, metadata.Unlock)()
		}()
		return nil
	default:
		return keyEvent
	}
}

func (session *viewerSession) updateEvent(node *tview.TreeNode, metadata *NodeMetadata) {
	node.ClearChildren().SetSelectedFunc(nil)
	event, err := getEvent(metadata.id)
	if err != nil {
		session.logError("Could not refresh event: ", err)
		return
	}

	sessions, err := session.getSessionNodes(metadata.metadata, event)
	if err != nil {
		session.logError("Could not load sessions: ", err)
		return
	}

	appendNodes(node, sessions...)

	if len(node.GetChildren()) == 0 {
		node.AddChild(nocontentNode())
	}
	node.SetExpanded(true)
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

func (session *viewerSession) getSeasonsNode() *tview.TreeNode {
	fullSessions := tview.NewTreeNode("Seasons").
		SetColor(activeTheme.CategoryNodeColor).
		SetReference(&NodeMetadata{nodeType: CategoryNode, metadata: MetaData{CategoryTitle: "Seasons"}})

	fullSessions.SetSelectedFunc(session.withBlink(fullSessions, func() {
		fullSessions.SetSelectedFunc(nil)
		seasons, err := session.getSeasonNodes()
		if err != nil {
			session.logError(err)
		} else {
			appendNodes(fullSessions, seasons...)
		}
	}, nil))
	return fullSessions
}

func (session *viewerSession) getPlaybackNodes(sessionTitles MetaData, epID string) []*tview.TreeNode {
	nodes := make([]*tview.TreeNode, 0)

	// add custom options
	if session.cfg.CustomPlaybackOptions != nil {
		for i := range session.cfg.CustomPlaybackOptions {
			com := session.cfg.CustomPlaybackOptions[i]
			if len(com.Command) > 0 {
				nodes = append(nodes, session.createCommandNode(sessionTitles, epID, com))
			}
		}
	}

	for _, c := range session.commands {
		nodes = append(nodes, session.createCommandNode(sessionTitles, epID, c))
	}

	clipboardNode := tview.NewTreeNode("Copy URL to clipboard").
		SetColor(activeTheme.ActionNodeColor).
		SetReference(&NodeMetadata{nodeType: ActionNode, metadata: sessionTitles})
	clipboardNode.SetSelectedFunc(func() {
		var url string
		var err error
		if epID != "" {
			url, err = getPlayableURL(epID, session.authtoken)
			if err != nil {
				session.logError(err)
				return
			}
		} else {
			url, err = getBackupStream()
			if err != nil {
				session.logError(err)
				return
			}
		}
		err = clipboard.WriteAll(url)
		if err != nil {
			session.logError(err)
			return
		}
		session.logInfo("URL copied to clipboard")
	})
	nodes = append(nodes, clipboardNode)
	return nodes
}

func (session *viewerSession) createCommandNode(t MetaData, epID string, c command) *tview.TreeNode {
	context := commandContext{
		MetaData:      t,
		EpID:          epID,
		CustomOptions: c,
	}
	node := tview.NewTreeNode(c.Title).
		SetColor(activeTheme.ActionNodeColor).
		SetReference(&NodeMetadata{nodeType: ActionNode, metadata: t})
	node.SetSelectedFunc(func() {
		go func() {
			err := session.runCustomCommand(context)
			if err != nil {
				session.logError(err)
			}
		}()
	})

	return node
}

func (session *viewerSession) getLiveNode() (bool, *tview.TreeNode, error) {
	var sessionNode *tview.TreeNode

	event, eventFound, err := getLiveWeekendEvent()
	if err != nil || !eventFound {
		return false, sessionNode, err
	}

	var t MetaData
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
			sessionNode = tview.NewTreeNode(s.SessionName + " - LIVE").
				SetColor(activeTheme.LiveColor).
				SetExpanded(false).
				SetReference(&NodeMetadata{nodeType: PlayableNode, id: event.UID, metadata: t})
			channels := session.getPerspectiveNodes(st, streams)
			appendNodes(sessionNode, channels...)
			sessionNode.AddChild(session.getBackupNode(t))
			return true, sessionNode, nil
		}
	}
	return false, sessionNode, nil
}

func (session *viewerSession) getBackupNode(sessionTitles MetaData) *tview.TreeNode {
	node := tview.NewTreeNode("backup stream").
		SetExpanded(false).
		SetReference(&NodeMetadata{nodeType: StreamNode, id: "", metadata: sessionTitles})
	nodes := session.getPlaybackNodes(sessionTitles, "")
	appendNodes(node, nodes...)

	return node
}

func (session *viewerSession) getEventNodes(season seasonStruct) ([]*tview.TreeNode, error) {
	// TODO: refactor
	errChan := make(chan error)
	events := make([]*tview.TreeNode, len(season.EventoccurrenceUrls))
	// iterate through events
	for m, eventID := range season.EventoccurrenceUrls {
		go func(eventID string, m int) {
			node, err := session.getEventNode(eventID, season.Name, m+1)
			if err == nil {
				events[m] = node
			}

			if err == errNoSessions {
				err = nil
			}

			errChan <- err
		}(eventID, m)
	}
	for index := 0; index < len(season.EventoccurrenceUrls); index++ {
		err := <-errChan
		if err != nil {
			return nil, err
		}
	}
	return events, nil
}

func (session *viewerSession) getEventNode(eventID string, seasonName string, sequenceNumber int) (*tview.TreeNode, error) {
	event, err := getEvent(eventID)
	if err != nil {
		return nil, err
	}
	// don't create node if there are no sessions
	if len(event.SessionoccurrenceUrls) == 0 {
		return nil, errNoSessions
	}

	titles := MetaData{SeasonTitle: seasonName, CategoryTitle: "Seasons", OrdinalNumber: sequenceNumber}

	eventNode := tview.NewTreeNode(event.Name).
		SetSelectable(true).
		SetReference(&NodeMetadata{nodeType: EventNode, id: eventID, metadata: titles})
	eventNode.SetSelectedFunc(session.withBlink(eventNode, func() {
		eventNode.SetSelectedFunc(nil)
		sessions, err := session.getSessionNodes(titles, event)
		if err != nil {
			session.logError(err)
		} else {
			appendNodes(eventNode, sessions...)
		}
		if len(eventNode.GetChildren()) == 0 {
			eventNode.AddChild(nocontentNode())
		}
	}, nil))
	return eventNode, nil

}

func (session *viewerSession) getSessionNodes(t MetaData, event eventStruct) ([]*tview.TreeNode, error) {
	var sessions []*tview.TreeNode
	var bonusIDs []string
	sessionsData, err := getSessions(event.SessionoccurrenceUrls)
	if err != nil {
		return nil, err
	}
	t.EventTitle = event.Name
	legacy := event.EndDate.Before(time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC))
	for _, s := range sessionsData {
		if legacy && len(s.ContentUrls) == 0 {
			continue
		}
		st := t
		st.SessionTitle = s.Name
		if s.Status != "upcoming" && s.Status != "expired" {
			bonusIDs = append(bonusIDs, s.ContentUrls...)
			s := s
			if s.StartTime.IsZero() {
				st.Date = event.EndDate.Time
			} else {
				st.Date = s.StartTime
			}

			sessionNode := tview.NewTreeNode(s.Name).
				SetSelectable(true).
				SetReference(&NodeMetadata{nodeType: PlayableNode, id: s.UID, metadata: t})
			sessionNode.SetSelectedFunc(session.withBlink(sessionNode, func() {
				sessionNode.SetSelectedFunc(nil)
				streams, err := getSessionStreams(s.UID)
				if err != nil {
					session.logError(err)
					return
				}
				if len(streams) == 0 {
					episodes, err := session.getEpisodeNodes(st, s.ContentUrls)
					if err != nil {
						session.logError(err)
						return
					}
					appendNodes(sessionNode, episodes...)
				} else {
					channels := session.getPerspectiveNodes(st, streams)
					appendNodes(sessionNode, channels...)
				}
			}, nil))
			if s.Status == "live" {
				sessionNode.SetText(s.Name + " - LIVE").
					SetColor(activeTheme.LiveColor)
			}
			sessions = append(sessions, sessionNode)
		}
	}
	if len(bonusIDs) > 0 && !legacy {
		bonusNode := tview.NewTreeNode("Bonus Content").
			SetExpanded(false).SetReference(&NodeMetadata{nodeType: CollectionNode, metadata: t})
		episodes, err := session.getEpisodeNodes(t, bonusIDs)
		if err != nil {
			return nil, err
		}
		appendNodes(bonusNode, episodes...)
		return append(sessions, bonusNode), nil
	}
	return sessions, nil
}

func (session *viewerSession) getMultiCommandNodes(perspectives []channel) []*tview.TreeNode {
	if len(session.cfg.MultiCommand) == 0 {
		return nil
	}

	var nodes []*tview.TreeNode

	for _, multi := range session.cfg.MultiCommand {
		var commands []commandContext
		for _, target := range multi.Targets {
			cmd, err := session.getCommand(target)
			if err != nil {
				session.logError("could not add target to multi command: ", err)
				continue
			}

			perspective, err := findPerspectiveByName(target.MatchTitle, perspectives)
			if err != nil {
				continue
			}

			// If we have a match, run the given command!
			context := commandContext{
				MetaData:      MetaData{PerspectiveTitle: multi.Title},
				EpID:          perspective.Self,
				CustomOptions: cmd,
			}
			commands = append(commands, context)
		}
		// If no favorites are found, continue
		if len(commands) == 0 {
			continue
		}

		multiNode := tview.NewTreeNode(multi.Title).
			SetColor(activeTheme.MultiCommandColor).
			SetReference(&NodeMetadata{nodeType: ActionNode})
		multiNode.SetSelectedFunc(session.withBlink(multiNode, func() {
			multiNode.SetSelectedFunc(nil)
			for _, context := range commands {
				err := session.runCustomCommand(context)
				if err != nil {
					session.logError(err)
				}
			}
		}, nil))
		nodes = append(nodes, multiNode)
	}

	return nodes
}

func findPerspectiveByName(name string, perspectives []channel) (channel, error) {
	for _, perspective := range perspectives {
		if perspective.PrettyName() == name {
			return perspective, nil
		}
		// if the string doesn't match try regex
		r, err := regexp.Compile(name)
		if err != nil {
			continue
		}
		if r.MatchString(perspective.PrettyName()) {
			return perspective, nil
		}
	}
	return channel{}, fmt.Errorf("found no perspective matching '%s'", name)
}

func (session *viewerSession) getCommand(matcher channelMatcher) (command, error) {
	if len(matcher.Command) > 0 {
		return command{Command: matcher.Command}, nil
	}

	if matcher.CommandKey == "" {
		return command{}, fmt.Errorf("No command for matcher '%s' provided", matcher.MatchTitle)
	}
	for _, cmd := range session.cfg.CustomPlaybackOptions {
		if cmd.Title == matcher.CommandKey {
			return cmd, nil
		}
	}
	return command{}, fmt.Errorf("found no command matching '%s'", matcher.CommandKey)
}

func (session *viewerSession) getPerspectiveNodes(title MetaData, perspectives []channel) []*tview.TreeNode {
	var channels []*tview.TreeNode
	var teamsContasiner *tview.TreeNode

	// add the multi commands at the top
	multiCommands := session.getMultiCommandNodes(perspectives)
	channels = append(channels, multiCommands...)

	for _, streamPerspective := range perspectives {
		streamPerspective := streamPerspective
		name := streamPerspective.PrettyName()

		newTitle := title
		newTitle.PerspectiveTitle = name

		streamNode := tview.NewTreeNode(name).
			SetColor(activeTheme.ItemNodeColor).
			SetReference(&NodeMetadata{nodeType: StreamNode, id: streamPerspective.Self, metadata: newTitle})

		streamNode.SetSelectedFunc(func() {
			streamNode.SetSelectedFunc(nil)
			nodes := session.getPlaybackNodes(newTitle, streamPerspective.Self)
			appendNodes(streamNode, nodes...)
		})
		channels = append(channels, streamNode)
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
	decades := make(map[int][]*tview.TreeNode)
	for _, s := range seasons.Seasons {
		if s.HasContent {
			s := s
			seasonNode := tview.NewTreeNode(strconv.Itoa(s.Year)).SetReference(&NodeMetadata{nodeType: CategoryNode, id: s.UID})
			seasonNode.SetSelectedFunc(session.withBlink(seasonNode, func() {
				seasonNode.SetSelectedFunc(nil)
				events, err := session.getEventNodes(s)
				if err != nil {
					session.logError(err)
				}
				appendNodes(seasonNode, events...)
			}, nil))
			decade := s.Year / 10 * 10 // PepoThink
			if _, ok := decades[decade]; !ok {
				decades[decade] = []*tview.TreeNode{seasonNode}
			} else {
				decades[decade] = append(decades[decade], seasonNode)
			}
		}
	}
	var nodes []*tview.TreeNode
	for decade, seasons := range decades {
		decade := tview.NewTreeNode(strconv.Itoa(decade) + "s").
			SetExpanded(false).
			SetReference(&NodeMetadata{nodeType: CollectionNode, id: strconv.Itoa(decade) + "s"})
		appendNodes(decade, seasons...)
		nodes = append(nodes, decade)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].GetText() > nodes[j].GetText() })
	return nodes, nil
}

func (session *viewerSession) getEpisodeNodes(title MetaData, ids []string) ([]*tview.TreeNode, error) {
	var nodes []*tview.TreeNode
	yearNodesMap := make(map[string][]*tview.TreeNode)
	eps, err := session.loadEpisodes(ids)
	if err != nil {
		return nil, err
	}
	episodes := sortEpisodes(eps)
	for _, ep := range episodes {
		ep := ep
		if len(ep.Items) < 1 {
			continue
		}
		tempTitle := title
		tempTitle.EpisodeTitle = ep.Title
		node := tview.NewTreeNode(ep.Title).
			SetColor(activeTheme.ItemNodeColor).
			SetReference(&NodeMetadata{nodeType: PlayableNode, id: ep.UID, metadata: tempTitle}).
			SetExpanded(false)

		playbackNodes := session.getPlaybackNodes(tempTitle, ep.Items[0])
		appendNodes(node, playbackNodes...)

		if year, _, err := getYearAndRace(ep.DataSourceID); err == nil {
			yearEps, ok := yearNodesMap[year]
			if !ok {
				yearNodesMap[year] = []*tview.TreeNode{node}
			} else {
				yearNodesMap[year] = append(yearEps, node)
			}
		} else {
			nodes = append(nodes, node)
		}
	}

	if len(yearNodesMap) > 1 {
		for year, eps := range yearNodesMap {
			yearNode := tview.NewTreeNode(year).
				SetExpanded(false).
				SetReference(&NodeMetadata{nodeType: CollectionNode, id: year, metadata: title})
			appendNodes(yearNode, eps...)
			nodes = append(nodes, yearNode)
		}
	} else {
		for _, eps := range yearNodesMap {
			nodes = append(nodes, eps...)
		}
	}

	if len(nodes) == 1 {
		return nodes[0].GetChildren(), nil
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].GetText() > nodes[j].GetText() })
	return nodes, nil
}

func (session *viewerSession) getVodTypeNodes() ([]*tview.TreeNode, error) {
	var nodes []*tview.TreeNode
	vodTypes, err := getVodTypes()
	if err != nil {
		return nil, err
	}
	for _, vType := range vodTypes.Objects {
		vType := vType
		if len(vType.ContentUrls) > 0 {
			titles := MetaData{CategoryTitle: vType.Name}
			node := tview.NewTreeNode(vType.Name).
				SetColor(activeTheme.CategoryNodeColor).
				SetReference(&NodeMetadata{nodeType: CategoryNode, id: vType.UID, metadata: titles})
			node.SetSelectedFunc(session.withBlink(node, func() {
				node.SetSelectedFunc(nil)
				episodes, err := session.getEpisodeNodes(titles, vType.ContentUrls)
				if err != nil {
					session.logError(err)
				} else {
					appendNodes(node, episodes...)
				}
			}, nil))
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (session *viewerSession) getCollectionsNode() *tview.TreeNode {
	node := tview.NewTreeNode("Collections").
		SetColor(activeTheme.CategoryNodeColor).
		SetReference(&NodeMetadata{nodeType: CategoryNode})
	node.SetSelectedFunc(session.withBlink(node, func() {
		node.SetSelectedFunc(nil)
		list, err := getCollectionList()
		if err != nil {
			session.logError("could not load collections: ", err)
		}
		for _, coll := range list.Objects {
			collID := coll.UID
			child := tview.NewTreeNode(coll.Title).SetReference(&NodeMetadata{nodeType: CollectionNode, id: collID})
			child.SetSelectedFunc(session.withBlink(child, func() {
				child.SetSelectedFunc(nil)
				var nodes []*tview.TreeNode
				nodes, err = session.getCollectionContent(collID)
				if err != nil {
					session.logError(err)
				} else if len(nodes) > 0 {
					appendNodes(child, nodes...)
				} else {
					child.AddChild(nocontentNode())
				}
			}, nil))
			node.AddChild(child)
		}
	}, nil))
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
	return session.getEpisodeNodes(MetaData{CategoryTitle: coll.Title}, epIDs)
}

func nocontentNode() *tview.TreeNode {
	return tview.NewTreeNode("no content").
		SetColor(activeTheme.NoContentColor).
		SetReference(&NodeMetadata{nodeType: MiscNode})
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

func (session *viewerSession) toggleVisibility(node *tview.TreeNode) {
	if len(node.GetChildren()) > 0 {
		node.SetExpanded(!node.IsExpanded())
	}
}
