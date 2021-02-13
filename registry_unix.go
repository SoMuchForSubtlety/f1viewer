// +build !windows

package main

func (s *viewerSession) checkRegistry(c command) (command, bool) {
	_ = c.registry
	_ = c.registry32
	// noop
	return c, false
}
