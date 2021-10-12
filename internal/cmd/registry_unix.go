//go:build !windows
// +build !windows

package cmd

import "os/exec"

func checkRegistry(c Command) (Command, bool) {
	_ = c.registry
	_ = c.registry32
	return c, false
}

func checkFlatpak(c Command) (Command, bool) {
	if c.flatpakAppID == "" {
		// command is not flatpak
		return c, false
	}

	_, err := exec.LookPath("flatpak")
	if err != nil {
		// flatpak not installed
		return c, false
	}

	err = exec.Command("flatpak", "info", c.flatpakAppID).Run()
	if err != nil {
		// package not installed
		return c, false
	}

	c.Command[0] = c.flatpakAppID
	c.Command = append([]string{"flatpak", "run"}, c.Command...)

	// optional update title

	c.Title = c.Title + " Flatpak"

	return c, true
}
