package ui

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/cmd"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/util"
	"github.com/SoMuchForSubtlety/f1viewer/v2/pkg/f1tv/v2"

	"github.com/rivo/tview"
)

func (s *UIState) extractMetadata(metadata f1tv.Metadata, properties []f1tv.Properties) cmd.MetaData {
	meta := cmd.MetaData{
		Event:         util.FirstNonEmptyString(metadata.EmfAttributes.MeetingName, metadata.EmfAttributes.GlobalMeetingName),
		Title:         util.FirstNonEmptyString(metadata.Title, metadata.EmfAttributes.GlobalTitle, metadata.TitleBrief),
		Circuit:       util.FirstNonEmptyString(metadata.EmfAttributes.CircuitShortName, metadata.EmfAttributes.CircuitOfficialName),
		Year:          metadata.Year,
		EpisodeNumber: metadata.EpisodeNumber,
		Country:       util.FirstNonEmptyString(metadata.EmfAttributes.GlobalMeetingCountryName, metadata.EmfAttributes.MeetingCountryName, metadata.Country),
		Series:        metadata.EmfAttributes.Series,
		Session:       metadata.TitleBrief,
		Source:        map[string]interface{}{"metadata": metadata, "properties": properties},
	}
	if len(metadata.Genres) > 0 {
		meta.Category = metadata.Genres[0]
	}
	if len(properties) > 0 {
		meta.Date = time.Unix(properties[0].SessionStartDate/1000, properties[0].SessionStartDate%1000*1000000)
		meta.OrdinalNumber = properties[0].MeetingNumber
	}

	return meta
}

func (s *UIState) v2ContentNode(v f1tv.ContentContainer) *tview.TreeNode {
	streamNode := tview.NewTreeNode(util.FirstNonEmptyString(
		v.Metadata.Title,
		v.Metadata.TitleBrief,
		v.Metadata.EmfAttributes.GlobalTitle,
		v.Metadata.ShortDescription,
		v.Metadata.LongDescription,
	)).SetColor(activeTheme.ItemNodeColor).
		SetReference(&NodeMetadata{nodeType: StreamNode, id: v.Metadata.ContentID.String(), metadata: s.extractMetadata(v.Metadata, v.Properties)})
	streamNode.SetSelectedFunc(func() {
		streamNode.SetSelectedFunc(nil)

		perspectives := s.v2PerspectiveNodes(v)
		appendNodes(streamNode, perspectives...)
	})

	return streamNode
}

func (s *UIState) v2PerspectiveNodes(v f1tv.ContentContainer) []*tview.TreeNode {
	meta := s.extractMetadata(v.Metadata, v.Properties)
	s.logger.Infof("loading details for %s (%d)", meta.Title, v.Metadata.ContentID)
	details, err := s.v2.ContentDetails(v.Metadata.ContentID)
	if err != nil {
		s.logger.Errorf("could not get content details for '%d': %v", v.Metadata.ContentID, err)
	} else {
		meta = s.extractMetadata(details.Metadata, details.Properties)
	}

	// fall back to just the main stream if there was an error getting details
	// or there are no more streams
	if err != nil || len(details.Metadata.AdditionalStreams) == 0 {
		nodes := s.getPlaybackNodes(meta, func() (string, error) { return s.v2.GetPlaybackURL(f1tv.BIG_SCREEN_HLS, v.Metadata.ContentID, nil) })
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
			playbackNodes := s.getPlaybackNodes(meta2, func() (string, error) {
				return s.v2.GetPlaybackURL(f1tv.BIG_SCREEN_HLS, details.ContentID, &p.ChannelID)
			})
			appendNodes(node, playbackNodes...)
		})
		perspectives[i+1] = node
	}
	node := tview.NewTreeNode("World Feed").
		SetColor(activeTheme.ItemNodeColor).
		SetReference(&NodeMetadata{nodeType: PlayableNode, metadata: meta})
	node.SetSelectedFunc(func() {
		node.SetSelectedFunc(nil)
		playbackNodes := s.getPlaybackNodes(meta, func() (string, error) { return s.v2.GetPlaybackURL(f1tv.BIG_SCREEN_HLS, v.Metadata.ContentID, nil) })
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
		s.logger.Infof("checking %q", multi.Title)
		commands := s.extractCommands(multi, perspectives, mainStream)

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

func (s *UIState) extractCommands(multi cmd.MultiCommand, perspectives []f1tv.AdditionalStream, mainStream f1tv.ContentContainer) []cmd.CommandContext {
	var commands []cmd.CommandContext
	for _, target := range multi.Targets {
		mainFeed, perspective, err := findPerspectiveByName(target.MatchTitle, perspectives, mainStream)
		if err != nil {
			continue
		}

		var urlFunc func() (string, error)
		if mainFeed != nil {
			urlFunc = func() (string, error) {
				return s.v2.GetPlaybackURL(f1tv.BIG_SCREEN_HLS, mainStream.Metadata.ContentID, nil)
			}
		} else {
			urlFunc = func() (string, error) {
				return s.v2.GetPlaybackURL(f1tv.BIG_SCREEN_HLS, mainFeed.Metadata.ContentID, &perspective.ChannelID)
			}
		}
		targetCmd := s.cmd.GetCommand(target)
		if len(targetCmd.Command) == 0 {
			s.logger.Errorf("could not determine command for %q - %q", multi.Title, target.MatchTitle)
			continue
		}

		meta := s.extractMetadata(mainStream.Metadata, mainStream.Properties)
		if perspective != nil {
			meta.PerspectiveTitle = perspective.PrettyName()
		}
		// If we have a match, run the given command!
		context := cmd.CommandContext{
			MetaData:      meta,
			CustomOptions: targetCmd,
			URL:           urlFunc,
		}
		commands = append(commands, context)
	}
	return commands
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
