// +build !windows

package cmd

func checkRegistry(c Command) (Command, bool) {
	_ = c.registry
	_ = c.registry32
	return c, false
}
