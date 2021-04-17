package ui

import "time"

func (s *UIState) checkLive() {
	for {
		s.logger.Info("checking for live session")
		isLive, liveNode, err := s.getLiveNode()
		if err != nil {
			s.logger.Error("error looking for live session: ", err)
			if s.cfg.LiveRetryTimeout <= 0 {
				return
			}
		} else if isLive {
			s.changes <- liveNode
			return
		} else if s.cfg.LiveRetryTimeout <= 0 {
			s.logger.Info("no live session found")
			return
		} else {
			s.logger.Info("no live session found")
		}
		time.Sleep(time.Second * time.Duration(s.cfg.LiveRetryTimeout))
	}
}
