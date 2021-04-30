package ui

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/SoMuchForSubtlety/f1viewer/internal/cmd"
	"github.com/SoMuchForSubtlety/f1viewer/internal/util"
	"github.com/SoMuchForSubtlety/f1viewer/pkg/f1tv/v1"
	v2 "github.com/SoMuchForSubtlety/f1viewer/pkg/f1tv/v2"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var errNoSessions = errors.New("event has not past or live events")

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
		go func() {
			metadata.Lock()
			updateFunc := func() { s.updateEvent(node, metadata) }
			s.withBlink(node, updateFunc, metadata.Unlock)()
		}()
		return nil
	default:
		return keyEvent
	}
}

func (s *UIState) updateEvent(node *tview.TreeNode, metadata *NodeMetadata) {
	node.ClearChildren().SetSelectedFunc(nil)
	event, err := f1tv.GetEvent(metadata.id)
	if err != nil {
		s.logger.Error("Could not refresh event: ", err)
		return
	}

	sessions, err := s.getSessionNodes(metadata.metadata, event)
	if err != nil {
		s.logger.Error("Could not load sessions: ", err)
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

func (s *UIState) getV2Node() *tview.TreeNode {
	homepage := tview.NewTreeNode("V2").
		SetColor(activeTheme.CategoryNodeColor).
		SetReference(&NodeMetadata{nodeType: CategoryNode, metadata: cmd.MetaData{CategoryTitle: "V2"}})

	homepage.SetSelectedFunc(s.withBlink(homepage, func() {
		homepage.SetSelectedFunc(nil)
		seasons := s.getHomepageNodes()
		{
			appendNodes(homepage, seasons...)
		}
	}, nil))
	return homepage
}

func (s *UIState) getSeasonsNode() *tview.TreeNode {
	fullSessions := tview.NewTreeNode("Seasons").
		SetColor(activeTheme.CategoryNodeColor).
		SetReference(&NodeMetadata{nodeType: CategoryNode, metadata: cmd.MetaData{CategoryTitle: "Seasons"}})

	fullSessions.SetSelectedFunc(s.withBlink(fullSessions, func() {
		fullSessions.SetSelectedFunc(nil)
		seasons, err := s.getSeasonNodes()
		if err != nil {
			s.logger.Error(err)
		} else {
			appendNodes(fullSessions, seasons...)
		}
	}, nil))
	return fullSessions
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
		m := cmd.MetaData{
			EpisodeTitle: v.Metadata.Title,
		}
		streamNode := s.v2ContentNode(v, m)
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

func (s *UIState) getEventNodes(season f1tv.Season) ([]*tview.TreeNode, error) {
	// TODO: refactor
	errChan := make(chan error)
	events := make([]*tview.TreeNode, len(season.EventoccurrenceUrls))
	// iterate through events
	for m, eventID := range season.EventoccurrenceUrls {
		go func(eventID string, m int) {
			node, err := s.getEventNode(eventID, season.Name, m+1)
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

func (s *UIState) getEventNode(eventID string, seasonName string, sequenceNumber int) (*tview.TreeNode, error) {
	event, err := f1tv.GetEvent(eventID)
	if err != nil {
		return nil, err
	}
	// don't create node if there are no sessions
	if len(event.SessionoccurrenceUrls) == 0 {
		return nil, errNoSessions
	}

	titles := cmd.MetaData{SeasonTitle: seasonName, CategoryTitle: "Seasons", OrdinalNumber: sequenceNumber}

	eventNode := tview.NewTreeNode(event.Name).
		SetSelectable(true).
		SetReference(&NodeMetadata{nodeType: EventNode, id: eventID, metadata: titles})
	eventNode.SetSelectedFunc(s.withBlink(eventNode, func() {
		eventNode.SetSelectedFunc(nil)
		sessions, err := s.getSessionNodes(titles, event)
		if err != nil {
			s.logger.Error(err)
		} else {
			appendNodes(eventNode, sessions...)
		}
		if len(eventNode.GetChildren()) == 0 {
			eventNode.AddChild(nocontentNode())
		}
	}, nil))
	return eventNode, nil
}

func (s *UIState) getSessionNodes(t cmd.MetaData, event f1tv.Event) ([]*tview.TreeNode, error) {
	var sessions []*tview.TreeNode
	var bonusIDs []string
	sessionsData, err := f1tv.GetSessions(event.SessionoccurrenceUrls)
	if err != nil {
		return nil, err
	}
	t.EventTitle = event.Name
	legacy := event.EndDate.Before(time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC))
	for _, session := range sessionsData {
		if legacy && len(session.ContentUrls) == 0 {
			continue
		}
		st := t
		st.SessionTitle = session.Name
		if session.Status != "upcoming" && session.Status != "expired" {
			bonusIDs = append(bonusIDs, session.ContentUrls...)
			session := session
			if session.StartTime.IsZero() {
				st.Date = event.EndDate.Time
			} else {
				st.Date = session.StartTime
			}

			sessionNode := tview.NewTreeNode(session.Name).
				SetSelectable(true).
				SetReference(&NodeMetadata{nodeType: PlayableNode, id: session.UID, metadata: t})
			sessionNode.SetSelectedFunc(s.withBlink(sessionNode, func() {
				sessionNode.SetSelectedFunc(nil)
				streams, err := f1tv.GetSessionStreams(session.UID)
				if err != nil {
					s.logger.Error(err)
					return
				}
				if len(streams) == 0 {
					episodes, err := s.getEpisodeNodes(st, session.ContentUrls)
					if err != nil {
						s.logger.Error(err)
						return
					}
					appendNodes(sessionNode, episodes...)
				} else {
					channels := s.getPerspectiveNodes(st, streams)
					appendNodes(sessionNode, channels...)
				}
			}, nil))
			if session.Status == "live" {
				sessionNode.SetText(session.Name + " - LIVE").
					SetColor(activeTheme.LiveColor)
			}
			sessions = append(sessions, sessionNode)
		}
	}
	if len(bonusIDs) > 0 && !legacy {
		bonusNode := tview.NewTreeNode("Bonus Content").
			SetExpanded(false).SetReference(&NodeMetadata{nodeType: CollectionNode, metadata: t})
		episodes, err := s.getEpisodeNodes(t, bonusIDs)
		if err != nil {
			return nil, err
		}
		appendNodes(bonusNode, episodes...)
		return append(sessions, bonusNode), nil
	}
	return sessions, nil
}

func (s *UIState) getHomepageNodes() []*tview.TreeNode {
	headings, err := s.v2.GetVideoContainers()
	if err != nil {
		s.logger.Error(err)
		return nil
	}

	var headingNodes []*tview.TreeNode
	for _, h := range headings {
		h := h
		title := h.Metadata.Label
		if title == "" {
			title = h.RetrieveItems.ResultObj.MeetingName
		}
		metadata := cmd.MetaData{CategoryTitle: title}
		headingNode := tview.NewTreeNode(title).
			SetColor(activeTheme.FolderNodeColor).
			SetReference(&NodeMetadata{nodeType: CategoryNode, metadata: metadata}).
			SetExpanded(false)

		for _, v := range h.RetrieveItems.ResultObj.Containers {
			headingNode.AddChild(s.v2ContentNode(v, metadata))
		}
		headingNodes = append(headingNodes, headingNode)
	}

	return headingNodes
}

func (s *UIState) v2ContentNode(v v2.ContentContainer, meta cmd.MetaData) *tview.TreeNode {
	// TODO: more metadata
	meta.EpisodeTitle = v.Metadata.TitleBrief
	if meta.EpisodeTitle == "" {
		meta.EpisodeTitle = v.Metadata.Title
	}

	streamNode := tview.NewTreeNode(meta.EpisodeTitle).
		SetColor(activeTheme.ItemNodeColor).
		SetReference(&NodeMetadata{nodeType: StreamNode, id: strconv.Itoa(v.Metadata.ContentID), metadata: meta})

	streamNode.SetSelectedFunc(func() {
		streamNode.SetSelectedFunc(nil)
		details, err := s.v2.ContentDetails(v.Metadata.ContentID)
		if err != nil {
			s.logger.Errorf("could not get content details for '%d': %v", v.Metadata.ContentID, err)
		}
		// fall back to just the main stream if there was an error getting details
		// or there are no more streams
		if err != nil || len(details.Metadata.AdditionalStreams) == 0 {
			nodes := s.getPlaybackNodes(meta, func() (string, error) { return s.v2.GetPlaybackURL(v2.BIG_SCREEN_HLS, v.Metadata.ContentID) })
			appendNodes(streamNode, nodes...)
			return
		}

		streams := details.Metadata.AdditionalStreams
		perspectives := make([]*tview.TreeNode, len(streams)+1)

		sort.Slice(details.Metadata.AdditionalStreams, func(i, j int) bool {
			if streams[i].TeamName != "" && streams[j].TeamName != "" {
				return streams[i].TeamName < streams[j].TeamName
			}
			if streams[i].TeamName == "" && streams[j].TeamName == "" {
				return streams[i].Title < streams[j].Title
			}
			return streams[i].TeamName == ""
		})
		// TODO: paralellize loading?
		// TODO: animation
		for i, p := range streams {
			meta2 := meta
			title := fmt.Sprintf("[%2d] %s %s", p.RacingNumber, p.DriverFirstName, p.DriverLastName)
			if p.DriverLastName == "" {
				title = p.Title
			}
			meta2.PerspectiveTitle = title

			color := util.HexStringToColor(p.Hex)
			if p.Hex == "" {
				color = activeTheme.ItemNodeColor
			}

			node := tview.NewTreeNode(title).
				SetColor(color).
				SetReference(&NodeMetadata{nodeType: PlayableNode, metadata: meta2})

			node.SetSelectedFunc(func() {
				node.SetSelectedFunc(nil)
				playbackNodes := s.getPlaybackNodes(meta2, func() (string, error) { return s.v2.GetPerspectivePlaybackURL(v2.BIG_SCREEN_HLS, p.PlaybackURL) })
				appendNodes(node, playbackNodes...)
			})
			perspectives[i+1] = node
		}
		node := tview.NewTreeNode("World Feed").
			SetColor(activeTheme.ItemNodeColor).
			SetReference(&NodeMetadata{nodeType: PlayableNode, metadata: meta})
		node.SetSelectedFunc(func() {
			node.SetSelectedFunc(nil)
			playbackNodes := s.getPlaybackNodes(meta, func() (string, error) { return s.v2.GetPlaybackURL(v2.BIG_SCREEN_HLS, v.Metadata.ContentID) })
			appendNodes(node, playbackNodes...)
		})
		perspectives[0] = node
		appendNodes(streamNode, perspectives...)
	})

	return streamNode
}

// func (s *UIState) getMultiCommandNodes(perspectives []f1tv.Channel) []*tview.TreeNode {
// 	if len(s.cfg.MultiCommand) == 0 {
// 		return nil
// 	}

// 	var nodes []*tview.TreeNode

// 	for _, multi := range s.cfg.MultiCommand {
// 		var commands []cmd.CommandContext
// 		for _, target := range multi.Targets {
// 			cmd, err := s.getCommand(target)
// 			if err != nil {
// 				s.logger.Error("could not add target to multi command: ", err)
// 				continue
// 			}

// 			perspective, err := findPerspectiveByName(target.MatchTitle, perspectives)
// 			if err != nil {
// 				continue
// 			}

// 			// If we have a match, run the given command!
// 			context := cmd.CommandContext{
// 				MetaData:      cmd.MetaData{PerspectiveTitle: multi.Title},
// 				EpID:          perspective.Self,
// 				CustomOptions: cmd,
// 			}
// 			commands = append(commands, context)
// 		}
// 		// If no favorites are found, continue
// 		if len(commands) == 0 {
// 			continue
// 		}

// 		multiNode := tview.NewTreeNode(multi.Title).
// 			SetColor(activeTheme.MultiCommandColor).
// 			SetReference(&NodeMetadata{nodeType: ActionNode})
// 		multiNode.SetSelectedFunc(s.withBlink(multiNode, func() {
// 			multiNode.SetSelectedFunc(nil)
// 			for _, context := range commands {
// 				err := s.cmd.RunCommand(context, func() (string, error) { s.v1.GetPlayableURL(per) })
// 				if err != nil {
// 					s.logger.Error(err)
// 				}
// 			}
// 		}, nil))
// 		nodes = append(nodes, multiNode)
// 	}

// 	return nodes
// }

func findPerspectiveByName(name string, perspectives []f1tv.Channel) (f1tv.Channel, error) {
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
	return f1tv.Channel{}, fmt.Errorf("found no perspective matching '%s'", name)
}

// func (s *UIState) getCommand(matcher channelMatcher) (cmd.Command, error) {
// 	if len(matcher.Command) > 0 {
// 		return cmd.Command{Command: matcher.Command}, nil
// 	}

// 	if matcher.CommandKey == "" {
// 		return cmd.Command{}, fmt.Errorf("No command for matcher '%s' provided", matcher.MatchTitle)
// 	}
// 	for _, cmd := range s.cfg.CustomPlaybackOptions {
// 		if cmd.Title == matcher.CommandKey {
// 			return cmd, nil
// 		}
// 	}
// 	return cmd.Command{}, fmt.Errorf("found no command matching '%s'", matcher.CommandKey)
// }

func (s *UIState) getPerspectiveNodes(title cmd.MetaData, perspectives []f1tv.Channel) []*tview.TreeNode {
	var channels []*tview.TreeNode
	var teamsContasiner *tview.TreeNode

	// add the multi commands at the top
	// TODO
	// multiCommands := s.getMultiCommandNodes(perspectives)
	// channels = append(channels, multiCommands...)

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
			nodes := s.getPlaybackNodes(newTitle, func() (string, error) { return s.v1.GetPlayableURL(streamPerspective.Self) })
			appendNodes(streamNode, nodes...)
		})
		channels = append(channels, streamNode)
	}
	if teamsContasiner != nil {
		channels = append(channels, teamsContasiner)
	}
	return channels
}

func (s *UIState) getSeasonNodes() ([]*tview.TreeNode, error) {
	seasons, err := f1tv.GetSeasons()
	if err != nil {
		return nil, err
	}
	decades := make(map[int][]*tview.TreeNode)
	for _, season := range seasons {
		if season.HasContent {
			season := season
			seasonNode := tview.NewTreeNode(strconv.Itoa(season.Year)).SetReference(&NodeMetadata{nodeType: CategoryNode, id: season.UID})
			seasonNode.SetSelectedFunc(s.withBlink(seasonNode, func() {
				seasonNode.SetSelectedFunc(nil)
				events, err := s.getEventNodes(season)
				if err != nil {
					s.logger.Error(err)
				}
				appendNodes(seasonNode, events...)
			}, nil))
			decade := season.Year / 10 * 10 // PepoThink
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

func (s *UIState) getEpisodeNodes(title cmd.MetaData, ids []string) ([]*tview.TreeNode, error) {
	var nodes []*tview.TreeNode
	yearNodesMap := make(map[string][]*tview.TreeNode)
	eps, err := f1tv.LoadEpisodes(ids)
	if err != nil {
		return nil, err
	}
	for _, ep := range eps {
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

		playbackNodes := s.getPlaybackNodes(tempTitle, func() (string, error) { return s.v1.GetPlayableURL(ep.Items[0]) })
		appendNodes(node, playbackNodes...)

		if year, _, err := util.GetYearAndRace(ep.DataSourceID); err == nil {
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

func (s *UIState) getVodTypeNodes() ([]*tview.TreeNode, error) {
	var nodes []*tview.TreeNode
	vodTypes, err := f1tv.GetVodTypes()
	if err != nil {
		return nil, err
	}
	for _, vType := range vodTypes.Objects {
		vType := vType
		if len(vType.ContentUrls) > 0 {
			titles := cmd.MetaData{CategoryTitle: vType.Name}
			node := tview.NewTreeNode(vType.Name).
				SetColor(activeTheme.CategoryNodeColor).
				SetReference(&NodeMetadata{nodeType: CategoryNode, id: vType.UID, metadata: titles})
			node.SetSelectedFunc(s.withBlink(node, func() {
				node.SetSelectedFunc(nil)
				episodes, err := s.getEpisodeNodes(titles, vType.ContentUrls)
				if err != nil {
					s.logger.Error(err)
				} else {
					appendNodes(node, episodes...)
				}
			}, nil))
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (s *UIState) getCollectionsNode() *tview.TreeNode {
	node := tview.NewTreeNode("Collections").
		SetColor(activeTheme.CategoryNodeColor).
		SetReference(&NodeMetadata{nodeType: CategoryNode})
	node.SetSelectedFunc(s.withBlink(node, func() {
		node.SetSelectedFunc(nil)
		list, err := f1tv.GetCollectionList()
		if err != nil {
			s.logger.Error("could not load collections: ", err)
		}
		for _, coll := range list {
			collID := coll.UID
			child := tview.NewTreeNode(coll.Title).SetReference(&NodeMetadata{nodeType: CollectionNode, id: collID})
			child.SetSelectedFunc(s.withBlink(child, func() {
				child.SetSelectedFunc(nil)
				var nodes []*tview.TreeNode
				nodes, err = s.getCollectionContent(collID)
				if err != nil {
					s.logger.Error(err)
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

func (s *UIState) getCollectionContent(id string) ([]*tview.TreeNode, error) {
	coll, err := f1tv.GetCollection(id)
	if err != nil {
		return nil, err
	}
	var epIDs []string
	for _, ep := range coll.Items {
		epIDs = append(epIDs, ep.ContentURL)
	}
	return s.getEpisodeNodes(cmd.MetaData{CategoryTitle: coll.Title}, epIDs)
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

func (s *UIState) toggleVisibility(node *tview.TreeNode) {
	if len(node.GetChildren()) > 0 {
		node.SetExpanded(!node.IsExpanded())
	}
}
