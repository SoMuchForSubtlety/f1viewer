package ui

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/cmd"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/util"
	"github.com/SoMuchForSubtlety/f1viewer/v2/pkg/f1tv/v2"

	"github.com/rivo/tview"
)

func (s *UIState) v2ContentNode(v f1tv.ContentContainer, meta cmd.MetaData) *tview.TreeNode {
	// TODO: more metadata
	meta.EpisodeTitle = v.Metadata.TitleBrief
	if meta.EpisodeTitle == "" {
		meta.EpisodeTitle = v.Metadata.Title
	}

	streamNode := tview.NewTreeNode(meta.EpisodeTitle).
		SetColor(activeTheme.ItemNodeColor).
		SetReference(&NodeMetadata{nodeType: StreamNode, id: strconv.FormatInt(v.Metadata.ContentID, 10), metadata: meta})

	streamNode.SetSelectedFunc(func() {
		streamNode.SetSelectedFunc(nil)

		perspectives := s.v2PerspectiveNodes(v, meta)
		appendNodes(streamNode, perspectives...)
	})

	return streamNode
}

func (s *UIState) v2PerspectiveNodes(v f1tv.ContentContainer, meta cmd.MetaData) []*tview.TreeNode {
	details, err := s.v2.ContentDetails(v.Metadata.ContentID)
	if err != nil {
		s.logger.Errorf("could not get content details for '%d': %v", v.Metadata.ContentID, err)
	}
	// fall back to just the main stream if there was an error getting details
	// or there are no more streams
	if err != nil || len(details.Metadata.AdditionalStreams) == 0 {
		nodes := s.getPlaybackNodes(meta, func() (string, error) { return s.v2.GetPlaybackURL(f1tv.BIG_SCREEN_HLS, v.Metadata.ContentID) })
		return nodes
	}

	streams := details.Metadata.AdditionalStreams
	perspectives := make([]*tview.TreeNode, len(streams)+1)

	sort.Slice(streams, func(i, j int) bool {
		if streams[i].TeamName != "" && streams[j].TeamName != "" {
			return streams[i].TeamName < streams[j].TeamName
		}
		if streams[i].TeamName == "" && streams[j].TeamName == "" {
			return streams[i].Title < streams[j].Title
		}
		return streams[i].TeamName == ""
	})

	for i, p := range streams {
		p := p
		meta2 := meta
		meta2.PerspectiveTitle = p.PrettyName()

		color := util.HexStringToColor(p.Hex)
		if p.Hex == "" || s.cfg.DisableTeamColors {
			color = activeTheme.ItemNodeColor
		}

		node := tview.NewTreeNode(p.PrettyName()).
			SetColor(color).
			SetReference(&NodeMetadata{nodeType: PlayableNode, metadata: meta2})

		node.SetSelectedFunc(func() {
			node.SetSelectedFunc(nil)
			playbackNodes := s.getPlaybackNodes(meta2, func() (string, error) { return s.v2.GetPerspectivePlaybackURL(f1tv.BIG_SCREEN_HLS, p.PlaybackURL) })
			appendNodes(node, playbackNodes...)
		})
		perspectives[i+1] = node
	}
	node := tview.NewTreeNode("World Feed").
		SetColor(activeTheme.ItemNodeColor).
		SetReference(&NodeMetadata{nodeType: PlayableNode, metadata: meta})
	node.SetSelectedFunc(func() {
		node.SetSelectedFunc(nil)
		playbackNodes := s.getPlaybackNodes(meta, func() (string, error) { return s.v2.GetPlaybackURL(f1tv.BIG_SCREEN_HLS, v.Metadata.ContentID) })
		appendNodes(node, playbackNodes...)
	})
	perspectives[0] = node

	multicommands := s.v2MultiCommandNodes(streams, v)

	return append(multicommands, perspectives...)
}

func (s *UIState) v2MultiCommandNodes(perspectives []f1tv.AdditionalStream, mainStream f1tv.ContentContainer) []*tview.TreeNode {
	s.logger.Info("checking for multi commands")
	if len(s.cfg.MultiCommand) == 0 {
		return nil
	}

	var nodes []*tview.TreeNode

	for _, multi := range s.cmd.MultiCommads {
		s.logger.Info("chcking " + multi.Title)
		var commands []cmd.CommandContext
		for _, target := range multi.Targets {
			mainFeed, perspective, err := findPerspectiveByName(target.MatchTitle, perspectives, mainStream)
			if err != nil {
				continue
			}

			var urlFunc func() (string, error)
			if mainFeed != nil {
				urlFunc = func() (string, error) { return s.v2.GetPlaybackURL(f1tv.BIG_SCREEN_HLS, mainStream.Metadata.ContentID) }
			} else {
				urlFunc = func() (string, error) {
					return s.v2.GetPerspectivePlaybackURL(f1tv.BIG_SCREEN_HLS, perspective.PlaybackURL)
				}
			}
			// If we have a match, run the given command!
			context := cmd.CommandContext{
				MetaData:      cmd.MetaData{PerspectiveTitle: multi.Title},
				CustomOptions: s.cmd.GetCommand(target),
				URL:           urlFunc,
			}
			commands = append(commands, context)
		}
		// If no streams are matched, continue
		if len(commands) == 0 {
			continue
		}

		multiNode := tview.NewTreeNode(multi.Title).
			SetColor(activeTheme.MultiCommandColor).
			SetReference(&NodeMetadata{nodeType: ActionNode})
		multiNode.SetSelectedFunc(s.withBlink(multiNode, func() {
			multiNode.SetSelectedFunc(nil)
			for _, context := range commands {
				err := s.cmd.RunCommand(context)
				if err != nil {
					s.logger.Error(err)
				}
			}
		}, nil))
		nodes = append(nodes, multiNode)
	}

	return nodes
}

func findPerspectiveByName(name string, perspectives []f1tv.AdditionalStream, mainStream f1tv.ContentContainer) (*f1tv.ContentContainer, *f1tv.AdditionalStream, error) {
	notFound := fmt.Errorf("found no perspective matching '%s'", name)
	for _, perspective := range perspectives {
		if perspective.PrettyName() == name {
			return nil, &perspective, nil
		}
		// if the string doesn't match try regex
		r, err := regexp.Compile(name)
		if err != nil {
			continue
		}
		if r.MatchString(perspective.PrettyName()) {
			return nil, &perspective, nil
		}
	}
	if strings.EqualFold(name, "World Feed") {
		return &mainStream, nil, nil
	}
	r, err := regexp.Compile(name)
	if err != nil {
		return nil, nil, notFound
	}
	if r.MatchString("World Feed") {
		return &mainStream, nil, nil
	}
	return nil, nil, notFound
}
