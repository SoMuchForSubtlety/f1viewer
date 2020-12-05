// +build !windows

package main

func (s *viewerSession) checkRegistry(c command) (command, bool) {
	// noop
	return c, false
}
