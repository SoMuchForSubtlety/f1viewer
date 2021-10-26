package ui

import "time"

func (s *UIState) checkLive() {
	for {
		s.logger.Info("checking for live session")
		isLive, liveNode, err := s.getLiveNode()
		switch {
		case err != nil:
			s.logger.Error("error looking for live session: ", err)
			if s.cfg.LiveRetryTimeout <= 0 {
				return
			}
		case isLive:
			s.addLiveNode(liveNode)
			s.logger.Info("found live event")
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
