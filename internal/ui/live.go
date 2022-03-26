package ui

import (
	"time"

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/cmd"
	"github.com/SoMuchForSubtlety/f1viewer/v2/pkg/f1tv/v2"
)

func (s *UIState) checkLive() {
	for {
		s.logger.Info("checking for live session")
		isLive, liveNode, newSessions, err := s.getLiveNode()
		switch {
		case err != nil:
			s.logger.Error("error looking for live session: ", err)
			if s.cfg.LiveRetryTimeout <= 0 {
				return
			}
		case isLive:
			if len(newSessions) == 0 {
				break
			}
			s.addLiveNode(liveNode)
			s.logger.Info("found live event")

			for _, session := range newSessions {
				meta := s.extractMetadata(session.Metadata, session.Properties)
				details, err := s.v2.ContentDetails(session.Metadata.ContentID)
				if err != nil {
					s.logger.Errorf("failed to load details for session %s: %v", meta.Title, err)
					continue
				}
				for _, liveHook := range s.cfg.LiveSessionHooks {
					s.runLiveHook(liveHook, session, details.Metadata.AdditionalStreams, meta)
				}
			}
		case s.cfg.LiveRetryTimeout <= 0:
			s.logger.Info("no live session found")
			return
		default:
			s.addLiveNode(nil) // remove live node
			s.logger.Info("no live session found")
		}
		time.Sleep(time.Second * time.Duration(s.cfg.LiveRetryTimeout))
	}
}

func (s *UIState) runLiveHook(hook cmd.MultiCommand, mainStream f1tv.ContentContainer, perspectives []f1tv.AdditionalStream, meta cmd.MetaData) {
	commands := s.extractCommands(hook, perspectives, mainStream)
	// If no streams are matched, continue
	if len(commands) == 0 {
		return
	}

	for _, context := range commands {
		err := s.cmd.RunCommand(context)
		if err != nil {
			s.logger.Error(err)
		}
	}
}
